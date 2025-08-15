package temporal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestClientConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ClientConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &ClientConfig{
				HostPort:  "localhost:7233",
				Namespace: "default",
				Identity:  "test-worker",
			},
			wantErr: false,
		},
		{
			name: "with TLS",
			config: &ClientConfig{
				HostPort:  "localhost:7233",
				Namespace: "default",
				Identity:  "test-worker",
				TLS: &TLSConfig{
					Enabled:    true,
					ServerName: "temporal.example.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.config.ToClientOptions()
			assert.Equal(t, tt.config.HostPort, opts.HostPort)
			assert.Equal(t, tt.config.Namespace, opts.Namespace)
			assert.Equal(t, tt.config.Identity, opts.Identity)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test loading config with defaults
	cfg, err := LoadConfig("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Client.HostPort)
	assert.NotEmpty(t, cfg.Client.Namespace)
	assert.NotEmpty(t, cfg.Workers)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Client: ClientConfig{
					HostPort:  "localhost:7233",
					Namespace: "default",
				},
				Workers: []WorkerConfig{
					{
						Name:      "worker1",
						TaskQueue: "queue1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing host port",
			config: &Config{
				Client: ClientConfig{
					Namespace: "default",
				},
				Workers: []WorkerConfig{
					{
						Name:      "worker1",
						TaskQueue: "queue1",
					},
				},
			},
			wantErr: true,
			errMsg:  "client.hostPort is required",
		},
		{
			name: "missing namespace",
			config: &Config{
				Client: ClientConfig{
					HostPort: "localhost:7233",
				},
				Workers: []WorkerConfig{
					{
						Name:      "worker1",
						TaskQueue: "queue1",
					},
				},
			},
			wantErr: true,
			errMsg:  "client.namespace is required",
		},
		{
			name: "no workers",
			config: &Config{
				Client: ClientConfig{
					HostPort:  "localhost:7233",
					Namespace: "default",
				},
				Workers: []WorkerConfig{},
			},
			wantErr: true,
			errMsg:  "at least one worker must be configured",
		},
		{
			name: "worker missing task queue",
			config: &Config{
				Client: ClientConfig{
					HostPort:  "localhost:7233",
					Namespace: "default",
				},
				Workers: []WorkerConfig{
					{
						Name: "worker1",
					},
				},
			},
			wantErr: true,
			errMsg:  "worker[0].taskQueue is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientConnectionManagement(t *testing.T) {
	// Use test suite for mocking
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	defer env.AssertExpectations(t)

	// Test client creation with mock
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}

	// Create client (will fail to connect in test, which is expected)
	c, err := NewClient(cfg)
	if err == nil {
		defer c.Close()
		
		// Test connection status
		assert.False(t, c.IsConnected())
		assert.NotNil(t, c.GetLastError())
	}
}

func TestWorkflowService(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	defer env.AssertExpectations(t)

	// Mock workflow function
	mockWorkflow := func(ctx context.Context, input string) (string, error) {
		return "result", nil
	}

	env.RegisterWorkflow(mockWorkflow)

	// Test workflow execution
	env.ExecuteWorkflow(mockWorkflow, "test-input")

	var result string
	err := env.GetWorkflowResult(&result)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
}

func TestMetricsReporter(t *testing.T) {
	reporter := NewMetricsReporter()

	// Record some metrics
	reporter.RecordWorkflowStarted("TestWorkflow")
	reporter.RecordWorkflowCompleted("TestWorkflow", 1*time.Second)
	reporter.RecordWorkflowFailed("TestWorkflow", "timeout")

	reporter.RecordActivityStarted("TestActivity")
	reporter.RecordActivityCompleted("TestActivity", 500*time.Millisecond)
	reporter.RecordActivityFailed("TestActivity", "error")

	// Get metrics
	metrics := reporter.GetMetrics()
	assert.NotNil(t, metrics)

	// Check workflow metrics
	workflowsStarted := metrics["workflows_started"].(map[string]int64)
	assert.Equal(t, int64(1), workflowsStarted["TestWorkflow"])

	workflowsCompleted := metrics["workflows_completed"].(map[string]int64)
	assert.Equal(t, int64(1), workflowsCompleted["TestWorkflow"])

	workflowsFailed := metrics["workflows_failed"].(map[string]int64)
	assert.Equal(t, int64(1), workflowsFailed["TestWorkflow:timeout"])

	// Check activity metrics
	activitiesStarted := metrics["activities_started"].(map[string]int64)
	assert.Equal(t, int64(1), activitiesStarted["TestActivity"])

	activitiesCompleted := metrics["activities_completed"].(map[string]int64)
	assert.Equal(t, int64(1), activitiesCompleted["TestActivity"])

	activitiesFailed := metrics["activities_failed"].(map[string]int64)
	assert.Equal(t, int64(1), activitiesFailed["TestActivity:error"])
}

func TestClientRetryPolicy(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}

	// Create client (will fail to connect in test)
	c, _ := NewClient(cfg)
	if c != nil {
		defer c.Close()

		// Check retry policy is set
		assert.NotNil(t, c.retryPolicy)
		assert.Equal(t, 1*time.Second, c.retryPolicy.InitialInterval)
		assert.Equal(t, 2.0, c.retryPolicy.BackoffCoefficient)
		assert.Equal(t, int32(10), c.retryPolicy.MaximumAttempts)
	}
}

func TestExecuteWorkflowOptions(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	defer env.AssertExpectations(t)

	// Mock workflow
	mockWorkflow := func(ctx context.Context) error {
		return nil
	}

	env.RegisterWorkflow(mockWorkflow)

	// Test with custom options
	env.RegisterDelayedCallback(func() {
		// Simulate workflow execution
	}, 0)

	env.ExecuteWorkflow(mockWorkflow)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestClientHealthCheck(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
		Connection: ConnectionConfig{
			ConnectionTimeout: 1 * time.Second,
			EnableRetry:       true,
			MaxRetryAttempts:  1,
			RetryBackoff:      100 * time.Millisecond,
		},
	}

	// Create client
	c, _ := NewClient(cfg)
	if c != nil {
		defer c.Close()

		// Trigger health check
		c.checkConnection()

		// Should not be connected (no server running)
		assert.False(t, c.IsConnected())
	}
}

func TestActivityService(t *testing.T) {
	cfg := &ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "test",
		Identity:  "test-worker",
	}

	c, _ := NewClient(cfg)
	if c != nil {
		defer c.Close()

		service := NewActivityService(c)
		assert.NotNil(t, service)

		// Test methods (will fail without connection, but should not panic)
		ctx := context.Background()
		_ = service.CompleteAsyncActivity(ctx, []byte("token"), "result", nil)
		_ = service.SendHeartbeat(ctx, []byte("token"), "details")
	}
}

func TestMinDuration(t *testing.T) {
	tests := []struct {
		a, b     time.Duration
		expected time.Duration
	}{
		{1 * time.Second, 2 * time.Second, 1 * time.Second},
		{3 * time.Second, 2 * time.Second, 2 * time.Second},
		{1 * time.Second, 1 * time.Second, 1 * time.Second},
	}

	for _, tt := range tests {
		result := minDuration(tt.a, tt.b)
		assert.Equal(t, tt.expected, result)
	}
}