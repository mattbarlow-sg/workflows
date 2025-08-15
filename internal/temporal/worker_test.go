package temporal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

func TestWorkerPoolCreation(t *testing.T) {
	// Create mock client
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}
	
	// Note: This will fail to connect in test environment
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("Cannot create client in test environment")
		return
	}
	defer client.Close()

	// Create registry
	registry := NewRegistry()

	// Create worker configs
	configs := []WorkerConfig{
		{
			Name:      "worker1",
			TaskQueue: "queue1",
			Options:   WorkerOptions{},
			Enabled:   true,
			AutoStart: false,
		},
		{
			Name:      "worker2",
			TaskQueue: "queue2",
			Options:   WorkerOptions{},
			Enabled:   false,
			AutoStart: false,
		},
	}

	// Create worker pool
	pool := NewWorkerPool(client, registry, configs)
	assert.NotNil(t, pool)
	assert.Equal(t, 2, len(pool.config))
}

func TestWorkerOptions(t *testing.T) {
	opts := WorkerOptions{
		MaxConcurrentWorkflowTaskExecutionSize:  50,
		MaxConcurrentActivityExecutionSize:      100,
		MaxConcurrentLocalActivityExecutionSize: 50,
		WorkerRateLimit:                         10.0,
		TaskQueueActivitiesPerSecond:            20.0,
		MaxWorkerActivitiesPerSecond:            30.0,
		EnableSessionWorker:                     true,
		MaxConcurrentSessionExecutionSize:       500,
		WorkflowPanicPolicy:                     "BlockWorkflow",
		WorkerStopTimeout:                       30 * time.Second,
		EnableLoggingInReplay:                   true,
		StickyScheduleToStartTimeout:            10 * time.Second,
		DisableEagerActivities:                  true,
		MaxHeartbeatThrottleInterval:            60 * time.Second,
		DefaultHeartbeatThrottleInterval:        30 * time.Second,
	}

	workerOpts := opts.ToWorkerOptions()
	assert.Equal(t, 50, workerOpts.MaxConcurrentWorkflowTaskExecutionSize)
	assert.Equal(t, 100, workerOpts.MaxConcurrentActivityExecutionSize)
	assert.Equal(t, 50, workerOpts.MaxConcurrentLocalActivityExecutionSize)
	assert.Equal(t, 20.0, workerOpts.TaskQueueActivitiesPerSecond)
	assert.Equal(t, 30.0, workerOpts.WorkerActivitiesPerSecond)
	assert.Equal(t, true, workerOpts.EnableSessionWorker)
	assert.Equal(t, 500, workerOpts.MaxConcurrentSessionExecutionSize)
	assert.Equal(t, worker.BlockWorkflow, workerOpts.WorkflowPanicPolicy)
	assert.Equal(t, 30*time.Second, workerOpts.WorkerStopTimeout)
	assert.Equal(t, true, workerOpts.EnableLoggingInReplay)
	assert.Equal(t, 10*time.Second, workerOpts.StickyScheduleToStartTimeout)
	assert.Equal(t, true, workerOpts.DisableEagerActivities)
	assert.Equal(t, 60*time.Second, workerOpts.MaxHeartbeatThrottleInterval)
	assert.Equal(t, 30*time.Second, workerOpts.DefaultHeartbeatThrottleInterval)
}

func TestWorkerPanicPolicy(t *testing.T) {
	tests := []struct {
		policy   string
		expected worker.WorkflowPanicPolicy
	}{
		{"BlockWorkflow", worker.BlockWorkflow},
		{"FailWorkflow", worker.FailWorkflow},
		{"", worker.FailWorkflow}, // Default
		{"Unknown", worker.FailWorkflow}, // Default for unknown
	}

	for _, tt := range tests {
		opts := WorkerOptions{
			WorkflowPanicPolicy: tt.policy,
		}
		workerOpts := opts.ToWorkerOptions()
		assert.Equal(t, tt.expected, workerOpts.WorkflowPanicPolicy)
	}
}

func TestManagedWorker(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	defer env.AssertExpectations(t)

	config := WorkerConfig{
		Name:      "test-worker",
		TaskQueue: "test-queue",
		Options:   WorkerOptions{},
		Enabled:   true,
		AutoStart: true,
	}

	managed := &ManagedWorker{
		config:      config,
		taskQueue:   config.TaskQueue,
		stopChan:    make(chan struct{}),
		restartChan: make(chan struct{}, 1),
	}

	// Test status
	status := managed.GetStatus()
	assert.Equal(t, "test-worker", status.Name)
	assert.Equal(t, "test-queue", status.TaskQueue)
	assert.False(t, status.IsRunning)
	assert.Equal(t, 0, status.ErrorCount)

	// Test IsRunning
	assert.False(t, managed.IsRunning())
}

func TestWorkerPoolStatus(t *testing.T) {
	// Create mock client
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}
	
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("Cannot create client in test environment")
		return
	}
	defer client.Close()

	registry := NewRegistry()
	configs := []WorkerConfig{
		{
			Name:      "worker1",
			TaskQueue: "queue1",
			Enabled:   true,
		},
	}

	pool := NewWorkerPool(client, registry, configs)
	
	// Add a mock worker
	managed := &ManagedWorker{
		config: configs[0],
		taskQueue: "queue1",
		isRunning: true,
		startTime: time.Now(),
	}
	pool.workers["worker1"] = managed

	// Get status
	status := pool.GetStatus()
	assert.Equal(t, 1, status.TotalWorkers)
	assert.Equal(t, 1, status.RunningWorkers)
	assert.Equal(t, 0, status.StoppedWorkers)
	assert.Contains(t, status.Workers, "worker1")
}

func TestWorkerManager(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}
	
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("Cannot create client in test environment")
		return
	}
	defer client.Close()

	registry := NewRegistry()
	configs := []WorkerConfig{
		{
			Name:      "worker1",
			TaskQueue: "queue1",
			Enabled:   true,
		},
	}

	manager := NewWorkerManager(client, registry, configs)
	assert.NotNil(t, manager)

	// Test metrics
	metrics := manager.GetMetrics()
	assert.Equal(t, 0, metrics.TotalWorkers)
	assert.Equal(t, 0, metrics.RunningWorkers)
	assert.Equal(t, 0, metrics.StoppedWorkers)
	assert.NotNil(t, metrics.TaskQueues)
	assert.NotNil(t, metrics.ErrorCounts)
}

func TestWorkerScaling(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}
	
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("Cannot create client in test environment")
		return
	}
	defer client.Close()

	registry := NewRegistry()
	configs := []WorkerConfig{}

	manager := NewWorkerManager(client, registry, configs)

	// Test scaling up
	err := manager.ScaleWorkers("test-queue", 3)
	// Will fail without actual connection, but should not panic
	_ = err

	// Test scaling down
	err = manager.ScaleWorkers("test-queue", 1)
	_ = err
}

func TestWorkerHealthCheck(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}
	
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("Cannot create client in test environment")
		return
	}
	defer client.Close()

	registry := NewRegistry()
	configs := []WorkerConfig{}

	manager := NewWorkerManager(client, registry, configs)

	// Health check should fail with no workers
	err := manager.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no workers are running")
}

func TestWorkerPoolOperations(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}
	
	client, _ := NewClient(cfg)
	if client == nil {
		t.Skip("Cannot create client in test environment")
		return
	}
	defer client.Close()

	registry := NewRegistry()
	configs := []WorkerConfig{}

	pool := NewWorkerPool(client, registry, configs)

	// Test GetWorker
	worker, exists := pool.GetWorker("nonexistent")
	assert.Nil(t, worker)
	assert.False(t, exists)

	// Test ListWorkers
	workers := pool.ListWorkers()
	assert.Empty(t, workers)

	// Test AddWorker
	newConfig := WorkerConfig{
		Name:      "new-worker",
		TaskQueue: "new-queue",
		Enabled:   true,
		AutoStart: false,
	}
	err := pool.AddWorker(newConfig)
	// Will fail without actual connection
	_ = err

	// Test RemoveWorker
	err = pool.RemoveWorker("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worker nonexistent not found")

	// Test RestartWorker
	err = pool.RestartWorker("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worker nonexistent not found")
}

func TestWorkerStopTimeout(t *testing.T) {
	managed := &ManagedWorker{
		config: WorkerConfig{
			Name:      "test-worker",
			TaskQueue: "test-queue",
		},
		stopChan: make(chan struct{}),
		isRunning: false,
	}

	// Stop should return immediately if not running
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := managed.Stop(ctx)
	assert.NoError(t, err)
}

func TestWorkerMetrics(t *testing.T) {
	metrics := WorkerMetrics{
		TotalWorkers:   5,
		RunningWorkers: 3,
		StoppedWorkers: 2,
		TaskQueues: map[string]int{
			"queue1": 2,
			"queue2": 3,
		},
		ErrorCounts: map[string]int{
			"worker1": 1,
			"worker3": 2,
		},
	}

	assert.Equal(t, 5, metrics.TotalWorkers)
	assert.Equal(t, 3, metrics.RunningWorkers)
	assert.Equal(t, 2, metrics.StoppedWorkers)
	assert.Equal(t, 2, metrics.TaskQueues["queue1"])
	assert.Equal(t, 3, metrics.TaskQueues["queue2"])
	assert.Equal(t, 1, metrics.ErrorCounts["worker1"])
	assert.Equal(t, 2, metrics.ErrorCounts["worker3"])
}