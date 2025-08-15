// Package temporal provides Temporal client and worker infrastructure
package temporal

import (
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Config holds configuration for Temporal client and workers
type Config struct {
	// Client configuration
	Client ClientConfig `json:"client" yaml:"client"`

	// Worker pools configuration
	Workers []WorkerConfig `json:"workers" yaml:"workers"`

	// Global settings
	Global GlobalConfig `json:"global" yaml:"global"`

	// Metrics configuration
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`

	// Development mode settings
	Development DevelopmentConfig `json:"development" yaml:"development"`
}

// ClientConfig contains Temporal client configuration
type ClientConfig struct {
	// Temporal server host:port
	HostPort string `json:"hostPort" yaml:"hostPort"`

	// Namespace to connect to
	Namespace string `json:"namespace" yaml:"namespace"`

	// Identity for this client
	Identity string `json:"identity" yaml:"identity"`

	// Data converter settings
	DataConverter DataConverterConfig `json:"dataConverter" yaml:"dataConverter"`

	// Connection options
	Connection ConnectionConfig `json:"connection" yaml:"connection"`

	// TLS configuration
	TLS *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`

	// Auth configuration
	Auth *AuthConfig `json:"auth,omitempty" yaml:"auth,omitempty"`
}

// WorkerConfig defines configuration for a worker pool
type WorkerConfig struct {
	// Worker name/identifier
	Name string `json:"name" yaml:"name"`

	// Task queue this worker listens on
	TaskQueue string `json:"taskQueue" yaml:"taskQueue"`

	// Worker options
	Options WorkerOptions `json:"options" yaml:"options"`

	// Enable this worker
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Auto-start this worker
	AutoStart bool `json:"autoStart" yaml:"autoStart"`
}

// WorkerOptions contains Temporal worker options
type WorkerOptions struct {
	// Maximum concurrent workflow task executions
	MaxConcurrentWorkflowTaskExecutionSize int `json:"maxConcurrentWorkflowTaskExecutionSize" yaml:"maxConcurrentWorkflowTaskExecutionSize"`

	// Maximum concurrent activity executions
	MaxConcurrentActivityExecutionSize int `json:"maxConcurrentActivityExecutionSize" yaml:"maxConcurrentActivityExecutionSize"`

	// Maximum concurrent local activity executions
	MaxConcurrentLocalActivityExecutionSize int `json:"maxConcurrentLocalActivityExecutionSize" yaml:"maxConcurrentLocalActivityExecutionSize"`

	// Worker rate limits
	WorkerRateLimit float64 `json:"workerRateLimit" yaml:"workerRateLimit"`

	// Task queue rate limits
	TaskQueueActivitiesPerSecond float64 `json:"taskQueueActivitiesPerSecond" yaml:"taskQueueActivitiesPerSecond"`

	// Maximum activities per second for this worker
	MaxWorkerActivitiesPerSecond float64 `json:"maxWorkerActivitiesPerSecond" yaml:"maxWorkerActivitiesPerSecond"`

	// Enable session worker
	EnableSessionWorker bool `json:"enableSessionWorker" yaml:"enableSessionWorker"`

	// Maximum concurrent session execution size
	MaxConcurrentSessionExecutionSize int `json:"maxConcurrentSessionExecutionSize" yaml:"maxConcurrentSessionExecutionSize"`

	// Workflow panic policy
	WorkflowPanicPolicy string `json:"workflowPanicPolicy" yaml:"workflowPanicPolicy"`

	// Worker stop timeout
	WorkerStopTimeout time.Duration `json:"workerStopTimeout" yaml:"workerStopTimeout"`

	// Enable logging in replay
	EnableLoggingInReplay bool `json:"enableLoggingInReplay" yaml:"enableLoggingInReplay"`

	// Sticky cache size
	StickyScheduleToStartTimeout time.Duration `json:"stickyScheduleToStartTimeout" yaml:"stickyScheduleToStartTimeout"`

	// Disable eager activities
	DisableEagerActivities bool `json:"disableEagerActivities" yaml:"disableEagerActivities"`

	// Max heartbeat throttle interval
	MaxHeartbeatThrottleInterval time.Duration `json:"maxHeartbeatThrottleInterval" yaml:"maxHeartbeatThrottleInterval"`

	// Default heartbeat throttle interval
	DefaultHeartbeatThrottleInterval time.Duration `json:"defaultHeartbeatThrottleInterval" yaml:"defaultHeartbeatThrottleInterval"`
}

// GlobalConfig contains global settings
type GlobalConfig struct {
	// Default task queue
	DefaultTaskQueue string `json:"defaultTaskQueue" yaml:"defaultTaskQueue"`

	// Default workflow execution timeout
	DefaultWorkflowExecutionTimeout time.Duration `json:"defaultWorkflowExecutionTimeout" yaml:"defaultWorkflowExecutionTimeout"`

	// Default workflow run timeout
	DefaultWorkflowRunTimeout time.Duration `json:"defaultWorkflowRunTimeout" yaml:"defaultWorkflowRunTimeout"`

	// Default workflow task timeout
	DefaultWorkflowTaskTimeout time.Duration `json:"defaultWorkflowTaskTimeout" yaml:"defaultWorkflowTaskTimeout"`

	// Default activity options
	DefaultActivityOptions ActivityOptionsConfig `json:"defaultActivityOptions" yaml:"defaultActivityOptions"`

	// Default retry policy
	DefaultRetryPolicy RetryPolicyConfig `json:"defaultRetryPolicy" yaml:"defaultRetryPolicy"`

	// Enable debug mode
	Debug bool `json:"debug" yaml:"debug"`

	// Log level
	LogLevel string `json:"logLevel" yaml:"logLevel"`
}

// ActivityOptionsConfig defines default activity options
type ActivityOptionsConfig struct {
	// Schedule to close timeout
	ScheduleToCloseTimeout time.Duration `json:"scheduleToCloseTimeout" yaml:"scheduleToCloseTimeout"`

	// Schedule to start timeout
	ScheduleToStartTimeout time.Duration `json:"scheduleToStartTimeout" yaml:"scheduleToStartTimeout"`

	// Start to close timeout
	StartToCloseTimeout time.Duration `json:"startToCloseTimeout" yaml:"startToCloseTimeout"`

	// Heartbeat timeout
	HeartbeatTimeout time.Duration `json:"heartbeatTimeout" yaml:"heartbeatTimeout"`

	// Retry policy
	RetryPolicy *RetryPolicyConfig `json:"retryPolicy,omitempty" yaml:"retryPolicy,omitempty"`
}

// RetryPolicyConfig defines retry policy settings
type RetryPolicyConfig struct {
	// Initial interval
	InitialInterval time.Duration `json:"initialInterval" yaml:"initialInterval"`

	// Backoff coefficient
	BackoffCoefficient float64 `json:"backoffCoefficient" yaml:"backoffCoefficient"`

	// Maximum interval
	MaximumInterval time.Duration `json:"maximumInterval" yaml:"maximumInterval"`

	// Maximum attempts
	MaximumAttempts int32 `json:"maximumAttempts" yaml:"maximumAttempts"`

	// Non-retryable error types
	NonRetryableErrorTypes []string `json:"nonRetryableErrorTypes" yaml:"nonRetryableErrorTypes"`
}

// ConnectionConfig contains connection settings
type ConnectionConfig struct {
	// Max connection age
	MaxConnectionAge time.Duration `json:"maxConnectionAge" yaml:"maxConnectionAge"`

	// Max connection age grace
	MaxConnectionAgeGrace time.Duration `json:"maxConnectionAgeGrace" yaml:"maxConnectionAgeGrace"`

	// Connection timeout
	ConnectionTimeout time.Duration `json:"connectionTimeout" yaml:"connectionTimeout"`

	// Keep alive time
	KeepAliveTime time.Duration `json:"keepAliveTime" yaml:"keepAliveTime"`

	// Keep alive timeout
	KeepAliveTimeout time.Duration `json:"keepAliveTimeout" yaml:"keepAliveTimeout"`

	// Keep alive permit without stream
	KeepAlivePermitWithoutStream bool `json:"keepAlivePermitWithoutStream" yaml:"keepAlivePermitWithoutStream"`

	// Max message size
	MaxMessageSize int `json:"maxMessageSize" yaml:"maxMessageSize"`

	// Enable retry
	EnableRetry bool `json:"enableRetry" yaml:"enableRetry"`

	// Max retry attempts
	MaxRetryAttempts int `json:"maxRetryAttempts" yaml:"maxRetryAttempts"`

	// Retry backoff
	RetryBackoff time.Duration `json:"retryBackoff" yaml:"retryBackoff"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	// Enable TLS
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Server name
	ServerName string `json:"serverName" yaml:"serverName"`

	// Root CA certificate path
	RootCAFile string `json:"rootCAFile" yaml:"rootCAFile"`

	// Client certificate path
	ClientCertFile string `json:"clientCertFile" yaml:"clientCertFile"`

	// Client key path
	ClientKeyFile string `json:"clientKeyFile" yaml:"clientKeyFile"`

	// Disable host verification
	DisableHostVerification bool `json:"disableHostVerification" yaml:"disableHostVerification"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	// Auth type (e.g., "bearer", "mtls", "api-key")
	Type string `json:"type" yaml:"type"`

	// Bearer token
	BearerToken string `json:"bearerToken,omitempty" yaml:"bearerToken,omitempty"`

	// API key
	APIKey string `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`

	// OAuth2 configuration
	OAuth2 *OAuth2Config `json:"oauth2,omitempty" yaml:"oauth2,omitempty"`
}

// OAuth2Config contains OAuth2 settings
type OAuth2Config struct {
	// Client ID
	ClientID string `json:"clientId" yaml:"clientId"`

	// Client secret
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`

	// Token URL
	TokenURL string `json:"tokenUrl" yaml:"tokenUrl"`

	// Scopes
	Scopes []string `json:"scopes" yaml:"scopes"`
}

// DataConverterConfig contains data converter settings
type DataConverterConfig struct {
	// Enable compression
	EnableCompression bool `json:"enableCompression" yaml:"enableCompression"`

	// Compression type
	CompressionType string `json:"compressionType" yaml:"compressionType"`

	// Enable encryption
	EnableEncryption bool `json:"enableEncryption" yaml:"enableEncryption"`

	// Encryption key ID
	EncryptionKeyID string `json:"encryptionKeyId" yaml:"encryptionKeyId"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	// Enable metrics
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Metrics endpoint
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Prometheus configuration
	Prometheus *PrometheusConfig `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`

	// OpenTelemetry configuration
	OpenTelemetry *OpenTelemetryConfig `json:"openTelemetry,omitempty" yaml:"openTelemetry,omitempty"`

	// Statsd configuration
	Statsd *StatsdConfig `json:"statsd,omitempty" yaml:"statsd,omitempty"`
}

// PrometheusConfig contains Prometheus metrics settings
type PrometheusConfig struct {
	// Listen address
	ListenAddress string `json:"listenAddress" yaml:"listenAddress"`

	// Metrics path
	MetricsPath string `json:"metricsPath" yaml:"metricsPath"`

	// Enable default metrics
	EnableDefaultMetrics bool `json:"enableDefaultMetrics" yaml:"enableDefaultMetrics"`
}

// OpenTelemetryConfig contains OpenTelemetry settings
type OpenTelemetryConfig struct {
	// Endpoint
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Headers
	Headers map[string]string `json:"headers" yaml:"headers"`

	// Use TLS
	UseTLS bool `json:"useTls" yaml:"useTls"`
}

// StatsdConfig contains Statsd settings
type StatsdConfig struct {
	// Host:port
	Address string `json:"address" yaml:"address"`

	// Prefix
	Prefix string `json:"prefix" yaml:"prefix"`

	// Flush interval
	FlushInterval time.Duration `json:"flushInterval" yaml:"flushInterval"`

	// Flush bytes
	FlushBytes int `json:"flushBytes" yaml:"flushBytes"`
}

// DevelopmentConfig contains development mode settings
type DevelopmentConfig struct {
	// Enable development mode
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Use local server
	UseLocalServer bool `json:"useLocalServer" yaml:"useLocalServer"`

	// Local server address
	LocalServerAddress string `json:"localServerAddress" yaml:"localServerAddress"`

	// Auto-reload on code changes
	AutoReload bool `json:"autoReload" yaml:"autoReload"`

	// Enable verbose logging
	VerboseLogging bool `json:"verboseLogging" yaml:"verboseLogging"`

	// Pretty print logs
	PrettyPrintLogs bool `json:"prettyPrintLogs" yaml:"prettyPrintLogs"`
}

// LoadConfig loads configuration from environment and optional file
func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{
		Client: ClientConfig{
			HostPort:  getEnvOrDefault("TEMPORAL_HOST", "localhost:7233"),
			Namespace: getEnvOrDefault("TEMPORAL_NAMESPACE", "default"),
			Identity:  getEnvOrDefault("TEMPORAL_IDENTITY", fmt.Sprintf("worker-%d", os.Getpid())),
			Connection: ConnectionConfig{
				ConnectionTimeout:    30 * time.Second,
				KeepAliveTime:        30 * time.Second,
				KeepAliveTimeout:     10 * time.Second,
				MaxMessageSize:       4 * 1024 * 1024, // 4MB
				EnableRetry:          true,
				MaxRetryAttempts:     3,
				RetryBackoff:         100 * time.Millisecond,
			},
		},
		Global: GlobalConfig{
			DefaultTaskQueue:                getEnvOrDefault("TEMPORAL_TASK_QUEUE", "default"),
			DefaultWorkflowExecutionTimeout: 24 * time.Hour,
			DefaultWorkflowRunTimeout:       24 * time.Hour,
			DefaultWorkflowTaskTimeout:      10 * time.Second,
			DefaultActivityOptions: ActivityOptionsConfig{
				ScheduleToCloseTimeout: 30 * time.Minute,
				ScheduleToStartTimeout: 10 * time.Minute,
				StartToCloseTimeout:    10 * time.Minute,
				HeartbeatTimeout:       30 * time.Second,
			},
			DefaultRetryPolicy: RetryPolicyConfig{
				InitialInterval:    1 * time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    100 * time.Second,
				MaximumAttempts:    3,
			},
			Debug:    getEnvOrDefault("TEMPORAL_DEBUG", "false") == "true",
			LogLevel: getEnvOrDefault("TEMPORAL_LOG_LEVEL", "info"),
		},
		Workers: []WorkerConfig{
			{
				Name:      "default",
				TaskQueue: getEnvOrDefault("TEMPORAL_TASK_QUEUE", "default"),
				Options: WorkerOptions{
					MaxConcurrentWorkflowTaskExecutionSize:  100,
					MaxConcurrentActivityExecutionSize:      100,
					MaxConcurrentLocalActivityExecutionSize: 100,
					WorkerRateLimit:                         0, // No limit
					TaskQueueActivitiesPerSecond:            0, // No limit
					MaxWorkerActivitiesPerSecond:            0, // No limit
					EnableSessionWorker:                     false,
					MaxConcurrentSessionExecutionSize:       1000,
					WorkflowPanicPolicy:                     "FailWorkflow",
					WorkerStopTimeout:                       30 * time.Second,
					EnableLoggingInReplay:                   false,
					StickyScheduleToStartTimeout:            5 * time.Second,
					DisableEagerActivities:                  false,
					MaxHeartbeatThrottleInterval:            60 * time.Second,
					DefaultHeartbeatThrottleInterval:        30 * time.Second,
				},
				Enabled:   true,
				AutoStart: true,
			},
		},
		Metrics: MetricsConfig{
			Enabled:  getEnvOrDefault("TEMPORAL_METRICS_ENABLED", "false") == "true",
			Endpoint: getEnvOrDefault("TEMPORAL_METRICS_ENDPOINT", ""),
		},
		Development: DevelopmentConfig{
			Enabled:            getEnvOrDefault("TEMPORAL_DEV_MODE", "false") == "true",
			UseLocalServer:     false,
			LocalServerAddress: "localhost:7233",
			AutoReload:         false,
			VerboseLogging:     false,
			PrettyPrintLogs:    true,
		},
	}

	// Load from file if provided
	if configPath != "" {
		// TODO: Implement YAML/JSON config file loading
		// This would override the defaults set above
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Client.HostPort == "" {
		return fmt.Errorf("client.hostPort is required")
	}
	if c.Client.Namespace == "" {
		return fmt.Errorf("client.namespace is required")
	}
	if len(c.Workers) == 0 {
		return fmt.Errorf("at least one worker must be configured")
	}
	for i, w := range c.Workers {
		if w.TaskQueue == "" {
			return fmt.Errorf("worker[%d].taskQueue is required", i)
		}
		if w.Name == "" {
			return fmt.Errorf("worker[%d].name is required", i)
		}
	}
	return nil
}

// ToClientOptions converts config to Temporal client options
func (c *ClientConfig) ToClientOptions() client.Options {
	opts := client.Options{
		HostPort:  c.HostPort,
		Namespace: c.Namespace,
		Identity:  c.Identity,
	}

	// TODO: Add TLS configuration if enabled
	// TODO: Add auth configuration if enabled
	// TODO: Add data converter configuration

	return opts
}

// ToWorkerOptions converts config to Temporal worker options
func (w *WorkerOptions) ToWorkerOptions() worker.Options {
	opts := worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize:   w.MaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentActivityExecutionSize:       w.MaxConcurrentActivityExecutionSize,
		MaxConcurrentLocalActivityExecutionSize:  w.MaxConcurrentLocalActivityExecutionSize,
		TaskQueueActivitiesPerSecond:             w.TaskQueueActivitiesPerSecond,
		WorkerActivitiesPerSecond:                w.MaxWorkerActivitiesPerSecond,
		EnableSessionWorker:                      w.EnableSessionWorker,
		MaxConcurrentSessionExecutionSize:        w.MaxConcurrentSessionExecutionSize,
		WorkerStopTimeout:                        w.WorkerStopTimeout,
		EnableLoggingInReplay:                    w.EnableLoggingInReplay,
		StickyScheduleToStartTimeout:             w.StickyScheduleToStartTimeout,
		DisableEagerActivities:                   w.DisableEagerActivities,
		MaxHeartbeatThrottleInterval:             w.MaxHeartbeatThrottleInterval,
		DefaultHeartbeatThrottleInterval:         w.DefaultHeartbeatThrottleInterval,
	}

	// Set panic policy
	switch w.WorkflowPanicPolicy {
	case "BlockWorkflow":
		opts.WorkflowPanicPolicy = worker.BlockWorkflow
	case "FailWorkflow":
		opts.WorkflowPanicPolicy = worker.FailWorkflow
	default:
		opts.WorkflowPanicPolicy = worker.FailWorkflow
	}

	return opts
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}