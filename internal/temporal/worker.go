// Package temporal provides Temporal client and worker infrastructure
package temporal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// WorkerPool manages multiple Temporal workers
type WorkerPool struct {
	client         *Client
	workers        map[string]*ManagedWorker
	registry       *Registry
	config         []WorkerConfig
	mu             sync.RWMutex
	stopChan       chan struct{}
	restartChan    chan struct{}
	wg             sync.WaitGroup
	metricsHandler *MetricsReporter
}

// ManagedWorker wraps a Temporal worker with lifecycle management
type ManagedWorker struct {
	worker.Worker
	config       WorkerConfig
	taskQueue    string
	isRunning    bool
	startTime    time.Time
	stopTime     time.Time
	errorCount   int
	lastError    error
	mu           sync.RWMutex
	stopChan     chan struct{}
	restartChan  chan struct{}
	wg           sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(client *Client, registry *Registry, configs []WorkerConfig) *WorkerPool {
	return &WorkerPool{
		client:      client,
		workers:     make(map[string]*ManagedWorker),
		registry:    registry,
		config:      configs,
		stopChan:    make(chan struct{}),
		restartChan: make(chan struct{}, len(configs)),
	}
}

// Start starts all configured workers
func (p *WorkerPool) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Starting worker pool with %d workers", len(p.config))

	for _, cfg := range p.config {
		if !cfg.Enabled {
			log.Printf("Worker %s is disabled, skipping", cfg.Name)
			continue
		}

		if cfg.AutoStart {
			if err := p.startWorker(cfg); err != nil {
				return fmt.Errorf("failed to start worker %s: %w", cfg.Name, err)
			}
		}
	}

	// Start worker monitor
	p.wg.Add(1)
	go p.monitorWorkers()

	log.Printf("Worker pool started successfully")
	return nil
}

// startWorker starts a single worker
func (p *WorkerPool) startWorker(cfg WorkerConfig) error {
	if _, exists := p.workers[cfg.Name]; exists {
		return fmt.Errorf("worker %s already exists", cfg.Name)
	}

	// Create worker options
	opts := cfg.Options.ToWorkerOptions()

	// Create the worker
	w := worker.New(p.client.Client, cfg.TaskQueue, opts)

	// Register workflows and activities from registry
	if err := p.registerWorkflowsAndActivities(w, cfg.TaskQueue); err != nil {
		return fmt.Errorf("failed to register workflows and activities: %w", err)
	}

	// Create managed worker
	managed := &ManagedWorker{
		Worker:      w,
		config:      cfg,
		taskQueue:   cfg.TaskQueue,
		stopChan:    make(chan struct{}),
		restartChan: make(chan struct{}, 1),
	}

	// Start the worker
	if err := managed.Start(); err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	p.workers[cfg.Name] = managed

	log.Printf("Started worker %s on task queue %s", cfg.Name, cfg.TaskQueue)
	return nil
}

// registerWorkflowsAndActivities registers workflows and activities with a worker
func (p *WorkerPool) registerWorkflowsAndActivities(w worker.Worker, taskQueue string) error {
	// Register workflows for this task queue
	workflows := p.registry.GetWorkflowsForTaskQueue(taskQueue)
	for name, wf := range workflows {
		w.RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{
			Name: name,
		})
		log.Printf("Registered workflow %s on task queue %s", name, taskQueue)
	}

	// Register activities for this task queue
	activities := p.registry.GetActivitiesForTaskQueue(taskQueue)
	for name, act := range activities {
		w.RegisterActivityWithOptions(act, activity.RegisterOptions{
			Name: name,
		})
		log.Printf("Registered activity %s on task queue %s", name, taskQueue)
	}

	return nil
}

// Stop stops all workers in the pool
func (p *WorkerPool) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Stopping worker pool...")

	// Signal shutdown
	close(p.stopChan)

	// Stop all workers
	var errors []error
	for name, w := range p.workers {
		log.Printf("Stopping worker %s", name)
		if err := w.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop worker %s: %w", name, err))
		}
	}

	// Wait for all goroutines
	p.wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping workers: %v", errors)
	}

	log.Printf("Worker pool stopped successfully")
	return nil
}

// monitorWorkers monitors worker health and handles restarts
func (p *WorkerPool) monitorWorkers() {
	defer p.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.checkWorkerHealth()
		case <-p.restartChan:
			p.handleWorkerRestart()
		}
	}
}

// checkWorkerHealth checks the health of all workers
func (p *WorkerPool) checkWorkerHealth() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for name, w := range p.workers {
		if !w.IsRunning() {
			log.Printf("Worker %s is not running, scheduling restart", name)
			select {
			case p.restartChan <- struct{}{}:
			default:
			}
		}
	}
}

// handleWorkerRestart handles restarting failed workers
func (p *WorkerPool) handleWorkerRestart() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, w := range p.workers {
		if !w.IsRunning() && w.config.AutoStart {
			log.Printf("Attempting to restart worker %s", name)
			
			// Stop the old worker
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_ = w.Stop(ctx)
			cancel()

			// Remove from pool
			delete(p.workers, name)

			// Try to restart
			if err := p.startWorker(w.config); err != nil {
				log.Printf("Failed to restart worker %s: %v", name, err)
				// Schedule another restart attempt
				time.AfterFunc(30*time.Second, func() {
					select {
					case p.restartChan <- struct{}{}:
					default:
					}
				})
			} else {
				log.Printf("Successfully restarted worker %s", name)
			}
		}
	}
}

// GetWorker returns a specific worker by name
func (p *WorkerPool) GetWorker(name string) (*ManagedWorker, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	w, exists := p.workers[name]
	return w, exists
}

// ListWorkers returns all workers in the pool
func (p *WorkerPool) ListWorkers() map[string]*ManagedWorker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	result := make(map[string]*ManagedWorker)
	for k, v := range p.workers {
		result[k] = v
	}
	return result
}

// AddWorker adds a new worker to the pool
func (p *WorkerPool) AddWorker(cfg WorkerConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.workers[cfg.Name]; exists {
		return fmt.Errorf("worker %s already exists", cfg.Name)
	}

	return p.startWorker(cfg)
}

// RemoveWorker removes a worker from the pool
func (p *WorkerPool) RemoveWorker(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	w, exists := p.workers[name]
	if !exists {
		return fmt.Errorf("worker %s not found", name)
	}

	// Stop the worker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := w.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop worker %s: %w", name, err)
	}

	delete(p.workers, name)
	log.Printf("Removed worker %s", name)
	return nil
}

// RestartWorker restarts a specific worker
func (p *WorkerPool) RestartWorker(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	w, exists := p.workers[name]
	if !exists {
		return fmt.Errorf("worker %s not found", name)
	}

	log.Printf("Restarting worker %s", name)

	// Stop the worker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = w.Stop(ctx)
	cancel()

	// Remove from pool
	delete(p.workers, name)

	// Restart
	if err := p.startWorker(w.config); err != nil {
		return fmt.Errorf("failed to restart worker %s: %w", name, err)
	}

	log.Printf("Successfully restarted worker %s", name)
	return nil
}

// ManagedWorker methods

// Start starts the managed worker
func (w *ManagedWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	// Start worker in background
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.startTime = time.Now()
		w.isRunning = true
		
		err := w.Worker.Run(worker.InterruptCh())
		if err != nil {
			w.mu.Lock()
			w.lastError = err
			w.errorCount++
			w.mu.Unlock()
			log.Printf("Worker stopped with error: %v", err)
		}
		
		w.mu.Lock()
		w.isRunning = false
		w.stopTime = time.Now()
		w.mu.Unlock()
	}()

	return nil
}

// Stop stops the managed worker
func (w *ManagedWorker) Stop(ctx context.Context) error {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	// Signal stop
	close(w.stopChan)

	// Stop the underlying worker
	w.Worker.Stop()

	// Wait for worker to stop
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns true if the worker is running
func (w *ManagedWorker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isRunning
}

// GetStatus returns the worker's status
func (w *ManagedWorker) GetStatus() WorkerStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return WorkerStatus{
		Name:       w.config.Name,
		TaskQueue:  w.taskQueue,
		IsRunning:  w.isRunning,
		StartTime:  w.startTime,
		StopTime:   w.stopTime,
		ErrorCount: w.errorCount,
		LastError:  w.lastError,
	}
}

// WorkerStatus represents the status of a worker
type WorkerStatus struct {
	Name       string
	TaskQueue  string
	IsRunning  bool
	StartTime  time.Time
	StopTime   time.Time
	ErrorCount int
	LastError  error
}

// WorkerPoolStatus represents the status of the worker pool
type WorkerPoolStatus struct {
	TotalWorkers   int
	RunningWorkers int
	StoppedWorkers int
	Workers        map[string]WorkerStatus
}

// GetStatus returns the status of the worker pool
func (p *WorkerPool) GetStatus() WorkerPoolStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := WorkerPoolStatus{
		TotalWorkers: len(p.workers),
		Workers:      make(map[string]WorkerStatus),
	}

	for name, w := range p.workers {
		workerStatus := w.GetStatus()
		status.Workers[name] = workerStatus
		
		if workerStatus.IsRunning {
			status.RunningWorkers++
		} else {
			status.StoppedWorkers++
		}
	}

	return status
}

// WorkerManager provides high-level worker management
type WorkerManager struct {
	pool     *WorkerPool
	client   *Client
	registry *Registry
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(client *Client, registry *Registry, configs []WorkerConfig) *WorkerManager {
	return &WorkerManager{
		pool:     NewWorkerPool(client, registry, configs),
		client:   client,
		registry: registry,
	}
}

// Start starts the worker manager
func (m *WorkerManager) Start(ctx context.Context) error {
	return m.pool.Start(ctx)
}

// Stop stops the worker manager
func (m *WorkerManager) Stop(ctx context.Context) error {
	return m.pool.Stop(ctx)
}

// ScaleWorkers scales the number of workers for a task queue
func (m *WorkerManager) ScaleWorkers(taskQueue string, count int) error {
	// Get current workers for this task queue
	currentWorkers := 0
	for _, w := range m.pool.ListWorkers() {
		if w.taskQueue == taskQueue {
			currentWorkers++
		}
	}

	if count > currentWorkers {
		// Scale up
		for i := currentWorkers; i < count; i++ {
			cfg := WorkerConfig{
				Name:      fmt.Sprintf("%s-worker-%d", taskQueue, i),
				TaskQueue: taskQueue,
				Options:   WorkerOptions{}, // Use defaults
				Enabled:   true,
				AutoStart: true,
			}
			if err := m.pool.AddWorker(cfg); err != nil {
				return fmt.Errorf("failed to add worker: %w", err)
			}
		}
	} else if count < currentWorkers {
		// Scale down
		workers := m.pool.ListWorkers()
		removed := 0
		for name, w := range workers {
			if w.taskQueue == taskQueue && removed < (currentWorkers-count) {
				if err := m.pool.RemoveWorker(name); err != nil {
					return fmt.Errorf("failed to remove worker %s: %w", name, err)
				}
				removed++
			}
		}
	}

	return nil
}

// GetMetrics returns worker metrics
func (m *WorkerManager) GetMetrics() WorkerMetrics {
	status := m.pool.GetStatus()
	
	metrics := WorkerMetrics{
		TotalWorkers:   status.TotalWorkers,
		RunningWorkers: status.RunningWorkers,
		StoppedWorkers: status.StoppedWorkers,
		TaskQueues:     make(map[string]int),
		ErrorCounts:    make(map[string]int),
	}

	for _, w := range status.Workers {
		metrics.TaskQueues[w.TaskQueue]++
		if w.ErrorCount > 0 {
			metrics.ErrorCounts[w.Name] = w.ErrorCount
		}
	}

	return metrics
}

// WorkerMetrics contains worker metrics
type WorkerMetrics struct {
	TotalWorkers   int
	RunningWorkers int
	StoppedWorkers int
	TaskQueues     map[string]int
	ErrorCounts    map[string]int
}

// HealthCheck performs a health check on all workers
func (m *WorkerManager) HealthCheck() error {
	status := m.pool.GetStatus()
	
	if status.RunningWorkers == 0 {
		return fmt.Errorf("no workers are running")
	}

	if float64(status.StoppedWorkers) > float64(status.TotalWorkers)*0.5 {
		return fmt.Errorf("more than 50%% of workers are stopped")
	}

	return nil
}