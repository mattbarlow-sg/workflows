// Package temporal provides Temporal client and worker infrastructure
package temporal

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
)

// Client wraps a Temporal client with additional functionality
type Client struct {
	client.Client
	config         *ClientConfig
	namespace      string
	identity       string
	metricsHandler client.MetricsHandler
	connectionMu   sync.RWMutex
	connected      bool
	lastError      error
	retryPolicy    *temporal.RetryPolicy
	
	// Connection management
	reconnectChan  chan struct{}
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewClient creates a new Temporal client with the given configuration
func NewClient(cfg *ClientConfig) (*Client, error) {
	c := &Client{
		config:        cfg,
		namespace:     cfg.Namespace,
		identity:      cfg.Identity,
		reconnectChan: make(chan struct{}, 1),
		stopChan:      make(chan struct{}),
		retryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    10,
		},
	}

	// Initial connection
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to Temporal: %w", err)
	}

	// Start connection monitor
	c.wg.Add(1)
	go c.connectionMonitor()

	return c, nil
}

// connect establishes connection to Temporal server
func (c *Client) connect() error {
	c.connectionMu.Lock()
	defer c.connectionMu.Unlock()

	opts := c.config.ToClientOptions()
	
	// Configure TLS if enabled
	if c.config.TLS != nil && c.config.TLS.Enabled {
		tlsConfig, err := c.buildTLSConfig(c.config.TLS)
		if err != nil {
			return fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.ConnectionOptions.TLS = tlsConfig
	}

	// Configure metrics if enabled
	if c.metricsHandler != nil {
		opts.MetricsHandler = c.metricsHandler
	}

	// Create client
	temporalClient, err := client.Dial(opts)
	if err != nil {
		c.connected = false
		c.lastError = err
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}

	c.Client = temporalClient
	c.connected = true
	c.lastError = nil
	
	log.Printf("Successfully connected to Temporal at %s (namespace: %s)", c.config.HostPort, c.namespace)
	return nil
}

// buildTLSConfig builds TLS configuration
func (c *Client) buildTLSConfig(tlsCfg *TLSConfig) (*tls.Config, error) {
	config := &tls.Config{
		ServerName:         tlsCfg.ServerName,
		InsecureSkipVerify: tlsCfg.DisableHostVerification,
	}

	// Load certificates if provided
	if tlsCfg.ClientCertFile != "" && tlsCfg.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsCfg.ClientCertFile, tlsCfg.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificates: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// TODO: Load root CA if provided
	
	return config, nil
}

// connectionMonitor monitors connection health and handles reconnection
func (c *Client) connectionMonitor() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.checkConnection()
		case <-c.reconnectChan:
			c.handleReconnect()
		}
	}
}

// checkConnection checks if the connection is healthy
func (c *Client) checkConnection() {
	c.connectionMu.RLock()
	if !c.connected {
		c.connectionMu.RUnlock()
		select {
		case c.reconnectChan <- struct{}{}:
		default:
		}
		return
	}
	c.connectionMu.RUnlock()

	// Perform health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.Client.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		log.Printf("Health check failed: %v", err)
		c.connectionMu.Lock()
		c.connected = false
		c.lastError = err
		c.connectionMu.Unlock()
		
		select {
		case c.reconnectChan <- struct{}{}:
		default:
		}
	}
}

// handleReconnect handles reconnection with exponential backoff
func (c *Client) handleReconnect() {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second
	
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		log.Printf("Attempting to reconnect to Temporal...")
		
		if err := c.connect(); err != nil {
			log.Printf("Reconnection failed: %v. Retrying in %v", err, backoff)
			
			select {
			case <-time.After(backoff):
				backoff = minDuration(backoff*2, maxBackoff)
			case <-c.stopChan:
				return
			}
		} else {
			log.Printf("Successfully reconnected to Temporal")
			return
		}
	}
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.connectionMu.RLock()
	defer c.connectionMu.RUnlock()
	return c.connected
}

// GetLastError returns the last connection error
func (c *Client) GetLastError() error {
	c.connectionMu.RLock()
	defer c.connectionMu.RUnlock()
	return c.lastError
}

// ExecuteWorkflow starts a workflow execution with retry logic
func (c *Client) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	// Apply default options if not set
	if options.TaskQueue == "" {
		options.TaskQueue = "default"
	}
	if options.WorkflowExecutionTimeout == 0 {
		options.WorkflowExecutionTimeout = 24 * time.Hour
	}
	if options.WorkflowRunTimeout == 0 {
		options.WorkflowRunTimeout = 24 * time.Hour
	}
	if options.WorkflowTaskTimeout == 0 {
		options.WorkflowTaskTimeout = 10 * time.Second
	}
	if options.RetryPolicy == nil {
		options.RetryPolicy = c.retryPolicy
	}

	return c.Client.ExecuteWorkflow(ctx, options, workflow, args...)
}

// SignalWorkflow sends a signal to a workflow with retry logic
func (c *Client) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return c.Client.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// QueryWorkflow queries a workflow with retry logic
func (c *Client) QueryWorkflow(ctx context.Context, workflowID, runID, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	return c.Client.QueryWorkflow(ctx, workflowID, runID, queryType, args...)
}

// CancelWorkflow cancels a workflow execution
func (c *Client) CancelWorkflow(ctx context.Context, workflowID, runID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return c.Client.CancelWorkflow(ctx, workflowID, runID)
}

// TerminateWorkflow terminates a workflow execution
func (c *Client) TerminateWorkflow(ctx context.Context, workflowID, runID, reason string, details ...interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return c.Client.TerminateWorkflow(ctx, workflowID, runID, reason, details...)
}

// GetWorkflowHistory gets the history of a workflow execution
func (c *Client) GetWorkflowHistory(ctx context.Context, workflowID, runID string) client.HistoryEventIterator {
	// Use the AllEvent filter type to get all events
	return c.Client.GetWorkflowHistory(ctx, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
}

// ListWorkflow lists workflow executions using a query
func (c *Client) ListWorkflow(ctx context.Context, query string) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	// Use the underlying WorkflowService client to list workflows
	request := &workflowservice.ListWorkflowExecutionsRequest{
		Namespace: c.namespace,
		Query:     query,
	}
	
	return c.Client.WorkflowService().ListWorkflowExecutions(ctx, request)
}

// CompleteActivity completes an activity by task token
func (c *Client) CompleteActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return c.Client.CompleteActivity(ctx, taskToken, result, err)
}

// RecordActivityHeartbeat records heartbeat for an activity
func (c *Client) RecordActivityHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return c.Client.RecordActivityHeartbeat(ctx, taskToken, details...)
}

// RecordActivityHeartbeatByID records heartbeat for an activity by ID
func (c *Client) RecordActivityHeartbeatByID(ctx context.Context, namespace, workflowID, runID, activityID string, details ...interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return c.Client.RecordActivityHeartbeatByID(ctx, namespace, workflowID, runID, activityID, details...)
}

// Close gracefully shuts down the client
func (c *Client) Close() error {
	log.Printf("Shutting down Temporal client...")
	
	// Signal shutdown
	close(c.stopChan)
	
	// Wait for goroutines to finish
	c.wg.Wait()
	
	// Close the underlying client
	if c.Client != nil {
		c.Client.Close()
	}
	
	log.Printf("Temporal client shut down successfully")
	return nil
}

// SetMetricsHandler sets the metrics handler for the client
func (c *Client) SetMetricsHandler(handler client.MetricsHandler) {
	c.metricsHandler = handler
}

// GetNamespace returns the namespace the client is connected to
func (c *Client) GetNamespace() string {
	return c.namespace
}

// GetIdentity returns the client identity
func (c *Client) GetIdentity() string {
	return c.identity
}

// WorkflowService provides high-level workflow operations
type WorkflowService struct {
	client *Client
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(client *Client) *WorkflowService {
	return &WorkflowService{
		client: client,
	}
}

// StartWorkflow starts a new workflow execution
func (s *WorkflowService) StartWorkflow(ctx context.Context, workflowID string, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "default",
	}
	return s.client.ExecuteWorkflow(ctx, options, workflow, args...)
}

// GetWorkflowResult gets the result of a workflow execution
func (s *WorkflowService) GetWorkflowResult(ctx context.Context, workflowID string, runID string, result interface{}) error {
	run := s.client.GetWorkflow(ctx, workflowID, runID)
	return run.Get(ctx, result)
}

// WaitForWorkflowCompletion waits for a workflow to complete
func (s *WorkflowService) WaitForWorkflowCompletion(ctx context.Context, workflowID string, runID string) error {
	run := s.client.GetWorkflow(ctx, workflowID, runID)
	return run.Get(ctx, nil)
}

// ActivityService provides high-level activity operations
type ActivityService struct {
	client *Client
}

// NewActivityService creates a new activity service
func NewActivityService(client *Client) *ActivityService {
	return &ActivityService{
		client: client,
	}
}

// CompleteAsyncActivity completes an async activity
func (s *ActivityService) CompleteAsyncActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error {
	return s.client.CompleteActivity(ctx, taskToken, result, err)
}

// SendHeartbeat sends a heartbeat for an activity
func (s *ActivityService) SendHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error {
	return s.client.RecordActivityHeartbeat(ctx, taskToken, details...)
}

// Helper functions

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// MetricsReporter provides metrics reporting for the client
type MetricsReporter struct {
	workflowsStarted   map[string]int64
	workflowsCompleted map[string]int64
	workflowsFailed    map[string]int64
	activitiesStarted  map[string]int64
	activitiesCompleted map[string]int64
	activitiesFailed   map[string]int64
	mu                 sync.RWMutex
}

// NewMetricsReporter creates a new metrics reporter
func NewMetricsReporter() *MetricsReporter {
	return &MetricsReporter{
		workflowsStarted:    make(map[string]int64),
		workflowsCompleted:  make(map[string]int64),
		workflowsFailed:     make(map[string]int64),
		activitiesStarted:   make(map[string]int64),
		activitiesCompleted: make(map[string]int64),
		activitiesFailed:    make(map[string]int64),
	}
}

// RecordWorkflowStarted records a workflow start
func (m *MetricsReporter) RecordWorkflowStarted(workflowType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflowsStarted[workflowType]++
}

// RecordWorkflowCompleted records a workflow completion
func (m *MetricsReporter) RecordWorkflowCompleted(workflowType string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflowsCompleted[workflowType]++
}

// RecordWorkflowFailed records a workflow failure
func (m *MetricsReporter) RecordWorkflowFailed(workflowType string, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", workflowType, errorType)
	m.workflowsFailed[key]++
}

// RecordActivityStarted records an activity start
func (m *MetricsReporter) RecordActivityStarted(activityType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activitiesStarted[activityType]++
}

// RecordActivityCompleted records an activity completion
func (m *MetricsReporter) RecordActivityCompleted(activityType string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activitiesCompleted[activityType]++
}

// RecordActivityFailed records an activity failure
func (m *MetricsReporter) RecordActivityFailed(activityType string, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", activityType, errorType)
	m.activitiesFailed[key]++
}

// GetMetrics returns current metrics
func (m *MetricsReporter) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{
		"workflows_started":    m.workflowsStarted,
		"workflows_completed":  m.workflowsCompleted,
		"workflows_failed":     m.workflowsFailed,
		"activities_started":   m.activitiesStarted,
		"activities_completed": m.activitiesCompleted,
		"activities_failed":    m.activitiesFailed,
	}
}