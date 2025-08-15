// Package temporal provides Temporal client and worker infrastructure
package temporal

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Registry manages workflow and activity registrations
type Registry struct {
	workflows  map[string]WorkflowRegistration
	activities map[string]ActivityRegistration
	mu         sync.RWMutex
}

// WorkflowRegistration represents a registered workflow
type WorkflowRegistration struct {
	Name          string
	Function      interface{}
	TaskQueues    []string
	Options       workflow.RegisterOptions
	Description   string
	InputType     reflect.Type
	OutputType    reflect.Type
	RegisteredAt  time.Time
}

// ActivityRegistration represents a registered activity
type ActivityRegistration struct {
	Name          string
	Function      interface{}
	TaskQueues    []string
	Options       activity.RegisterOptions
	Description   string
	InputType     reflect.Type
	OutputType    reflect.Type
	RegisteredAt  time.Time
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		workflows:  make(map[string]WorkflowRegistration),
		activities: make(map[string]ActivityRegistration),
	}
}

// RegisterWorkflow registers a workflow with the registry
func (r *Registry) RegisterWorkflow(wf interface{}, opts ...WorkflowOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Apply options
	config := &workflowConfig{
		taskQueues: []string{"default"},
	}
	for _, opt := range opts {
		opt(config)
	}

	// Get workflow name
	name := config.name
	if name == "" {
		name = getWorkflowName(wf)
	}

	// Check if already registered
	if _, exists := r.workflows[name]; exists {
		return fmt.Errorf("workflow %s already registered", name)
	}

	// Extract type information
	fnType := reflect.TypeOf(wf)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("workflow must be a function")
	}

	var inputType, outputType reflect.Type
	if fnType.NumIn() > 1 {
		inputType = fnType.In(1) // First param is workflow.Context
	}
	if fnType.NumOut() > 0 {
		outputType = fnType.Out(0)
	}

	// Create registration
	reg := WorkflowRegistration{
		Name:         name,
		Function:     wf,
		TaskQueues:   config.taskQueues,
		Options:      config.registerOptions,
		Description:  config.description,
		InputType:    inputType,
		OutputType:   outputType,
		RegisteredAt: time.Now(),
	}

	r.workflows[name] = reg
	return nil
}

// RegisterActivity registers an activity with the registry
func (r *Registry) RegisterActivity(act interface{}, opts ...ActivityOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Apply options
	config := &activityConfig{
		taskQueues: []string{"default"},
	}
	for _, opt := range opts {
		opt(config)
	}

	// Get activity name
	name := config.name
	if name == "" {
		name = getActivityName(act)
	}

	// Check if already registered
	if _, exists := r.activities[name]; exists {
		return fmt.Errorf("activity %s already registered", name)
	}

	// Extract type information
	fnType := reflect.TypeOf(act)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("activity must be a function")
	}

	var inputType, outputType reflect.Type
	if fnType.NumIn() > 1 {
		inputType = fnType.In(1) // First param is context.Context
	}
	if fnType.NumOut() > 0 {
		outputType = fnType.Out(0)
	}

	// Create registration
	reg := ActivityRegistration{
		Name:         name,
		Function:     act,
		TaskQueues:   config.taskQueues,
		Options:      config.registerOptions,
		Description:  config.description,
		InputType:    inputType,
		OutputType:   outputType,
		RegisteredAt: time.Now(),
	}

	r.activities[name] = reg
	return nil
}

// GetWorkflow returns a registered workflow by name
func (r *Registry) GetWorkflow(name string) (WorkflowRegistration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wf, exists := r.workflows[name]
	return wf, exists
}

// GetActivity returns a registered activity by name
func (r *Registry) GetActivity(name string) (ActivityRegistration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	act, exists := r.activities[name]
	return act, exists
}

// ListWorkflows returns all registered workflows
func (r *Registry) ListWorkflows() []WorkflowRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]WorkflowRegistration, 0, len(r.workflows))
	for _, wf := range r.workflows {
		result = append(result, wf)
	}
	return result
}

// ListActivities returns all registered activities
func (r *Registry) ListActivities() []ActivityRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]ActivityRegistration, 0, len(r.activities))
	for _, act := range r.activities {
		result = append(result, act)
	}
	return result
}

// GetWorkflowsForTaskQueue returns workflows registered for a specific task queue
func (r *Registry) GetWorkflowsForTaskQueue(taskQueue string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]interface{})
	for name, wf := range r.workflows {
		for _, tq := range wf.TaskQueues {
			if tq == taskQueue {
				result[name] = wf.Function
				break
			}
		}
	}
	return result
}

// GetActivitiesForTaskQueue returns activities registered for a specific task queue
func (r *Registry) GetActivitiesForTaskQueue(taskQueue string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]interface{})
	for name, act := range r.activities {
		for _, tq := range act.TaskQueues {
			if tq == taskQueue {
				result[name] = act.Function
				break
			}
		}
	}
	return result
}

// UnregisterWorkflow removes a workflow from the registry
func (r *Registry) UnregisterWorkflow(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.workflows[name]; !exists {
		return fmt.Errorf("workflow %s not found", name)
	}
	
	delete(r.workflows, name)
	return nil
}

// UnregisterActivity removes an activity from the registry
func (r *Registry) UnregisterActivity(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.activities[name]; !exists {
		return fmt.Errorf("activity %s not found", name)
	}
	
	delete(r.activities, name)
	return nil
}

// Clear removes all registrations
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.workflows = make(map[string]WorkflowRegistration)
	r.activities = make(map[string]ActivityRegistration)
}

// WorkflowOption configures workflow registration
type WorkflowOption func(*workflowConfig)

type workflowConfig struct {
	name            string
	taskQueues      []string
	description     string
	registerOptions workflow.RegisterOptions
}

// WithWorkflowName sets the workflow name
func WithWorkflowName(name string) WorkflowOption {
	return func(c *workflowConfig) {
		c.name = name
	}
}

// WithWorkflowTaskQueues sets the task queues for the workflow
func WithWorkflowTaskQueues(queues ...string) WorkflowOption {
	return func(c *workflowConfig) {
		c.taskQueues = queues
	}
}

// WithWorkflowDescription sets the workflow description
func WithWorkflowDescription(desc string) WorkflowOption {
	return func(c *workflowConfig) {
		c.description = desc
	}
}

// WithWorkflowOptions sets Temporal workflow registration options
func WithWorkflowOptions(opts workflow.RegisterOptions) WorkflowOption {
	return func(c *workflowConfig) {
		c.registerOptions = opts
	}
}

// ActivityOption configures activity registration
type ActivityOption func(*activityConfig)

type activityConfig struct {
	name            string
	taskQueues      []string
	description     string
	registerOptions activity.RegisterOptions
}

// WithActivityName sets the activity name
func WithActivityName(name string) ActivityOption {
	return func(c *activityConfig) {
		c.name = name
	}
}

// WithActivityTaskQueues sets the task queues for the activity
func WithActivityTaskQueues(queues ...string) ActivityOption {
	return func(c *activityConfig) {
		c.taskQueues = queues
	}
}

// WithActivityDescription sets the activity description
func WithActivityDescription(desc string) ActivityOption {
	return func(c *activityConfig) {
		c.description = desc
	}
}

// WithActivityOptions sets Temporal activity registration options
func WithActivityOptions(opts activity.RegisterOptions) ActivityOption {
	return func(c *activityConfig) {
		c.registerOptions = opts
	}
}

// Helper functions

func getWorkflowName(wf interface{}) string {
	t := reflect.TypeOf(wf)
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(wf).Pointer()).Name()
	}
	return fmt.Sprintf("%T", wf)
}

func getActivityName(act interface{}) string {
	t := reflect.TypeOf(act)
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(act).Pointer()).Name()
	}
	return fmt.Sprintf("%T", act)
}

// RegistryBuilder provides a fluent interface for building registrations
type RegistryBuilder struct {
	registry   *Registry
	errors     []error
}

// NewRegistryBuilder creates a new registry builder
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		registry: NewRegistry(),
		errors:   []error{},
	}
}

// AddWorkflow adds a workflow to the registry
func (b *RegistryBuilder) AddWorkflow(wf interface{}, opts ...WorkflowOption) *RegistryBuilder {
	if err := b.registry.RegisterWorkflow(wf, opts...); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// AddActivity adds an activity to the registry
func (b *RegistryBuilder) AddActivity(act interface{}, opts ...ActivityOption) *RegistryBuilder {
	if err := b.registry.RegisterActivity(act, opts...); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// Build returns the built registry and any errors
func (b *RegistryBuilder) Build() (*Registry, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("registry build errors: %v", b.errors)
	}
	return b.registry, nil
}

// RegistrySnapshot represents a snapshot of the registry state
type RegistrySnapshot struct {
	Workflows  []WorkflowInfo  `json:"workflows"`
	Activities []ActivityInfo  `json:"activities"`
	Timestamp  time.Time       `json:"timestamp"`
}

// WorkflowInfo contains workflow information
type WorkflowInfo struct {
	Name         string    `json:"name"`
	TaskQueues   []string  `json:"taskQueues"`
	Description  string    `json:"description"`
	InputType    string    `json:"inputType"`
	OutputType   string    `json:"outputType"`
	RegisteredAt time.Time `json:"registeredAt"`
}

// ActivityInfo contains activity information
type ActivityInfo struct {
	Name         string    `json:"name"`
	TaskQueues   []string  `json:"taskQueues"`
	Description  string    `json:"description"`
	InputType    string    `json:"inputType"`
	OutputType   string    `json:"outputType"`
	RegisteredAt time.Time `json:"registeredAt"`
}

// GetSnapshot returns a snapshot of the registry
func (r *Registry) GetSnapshot() RegistrySnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := RegistrySnapshot{
		Workflows:  make([]WorkflowInfo, 0, len(r.workflows)),
		Activities: make([]ActivityInfo, 0, len(r.activities)),
		Timestamp:  time.Now(),
	}

	for _, wf := range r.workflows {
		info := WorkflowInfo{
			Name:         wf.Name,
			TaskQueues:   wf.TaskQueues,
			Description:  wf.Description,
			RegisteredAt: wf.RegisteredAt,
		}
		if wf.InputType != nil {
			info.InputType = wf.InputType.String()
		}
		if wf.OutputType != nil {
			info.OutputType = wf.OutputType.String()
		}
		snapshot.Workflows = append(snapshot.Workflows, info)
	}

	for _, act := range r.activities {
		info := ActivityInfo{
			Name:         act.Name,
			TaskQueues:   act.TaskQueues,
			Description:  act.Description,
			RegisteredAt: act.RegisteredAt,
		}
		if act.InputType != nil {
			info.InputType = act.InputType.String()
		}
		if act.OutputType != nil {
			info.OutputType = act.OutputType.String()
		}
		snapshot.Activities = append(snapshot.Activities, info)
	}

	return snapshot
}

// AutoRegistry provides automatic registration based on tags or conventions
type AutoRegistry struct {
	registry *Registry
	scanner  *TypeScanner
}

// NewAutoRegistry creates a new auto-registry
func NewAutoRegistry(registry *Registry) *AutoRegistry {
	return &AutoRegistry{
		registry: registry,
		scanner:  NewTypeScanner(),
	}
}

// TypeScanner scans for types to register
type TypeScanner struct{}

// NewTypeScanner creates a new type scanner
func NewTypeScanner() *TypeScanner {
	return &TypeScanner{}
}

// ScanPackage scans a package for workflows and activities
func (s *TypeScanner) ScanPackage(packagePath string) ([]interface{}, []interface{}, error) {
	// This would use reflection or AST parsing to find tagged functions
	// For now, return empty slices
	return []interface{}{}, []interface{}{}, nil
}

// RegisterPackage registers all workflows and activities from a package
func (a *AutoRegistry) RegisterPackage(packagePath string, taskQueue string) error {
	workflows, activities, err := a.scanner.ScanPackage(packagePath)
	if err != nil {
		return fmt.Errorf("failed to scan package: %w", err)
	}

	// Register workflows
	for _, wf := range workflows {
		if err := a.registry.RegisterWorkflow(wf, WithWorkflowTaskQueues(taskQueue)); err != nil {
			return fmt.Errorf("failed to register workflow: %w", err)
		}
	}

	// Register activities
	for _, act := range activities {
		if err := a.registry.RegisterActivity(act, WithActivityTaskQueues(taskQueue)); err != nil {
			return fmt.Errorf("failed to register activity: %w", err)
		}
	}

	return nil
}

// DynamicRegistry allows runtime registration and unregistration
type DynamicRegistry struct {
	*Registry
	onChange func(RegistrySnapshot)
}

// NewDynamicRegistry creates a new dynamic registry
func NewDynamicRegistry(onChange func(RegistrySnapshot)) *DynamicRegistry {
	return &DynamicRegistry{
		Registry: NewRegistry(),
		onChange: onChange,
	}
}

// RegisterWorkflow registers a workflow and triggers onChange
func (d *DynamicRegistry) RegisterWorkflow(wf interface{}, opts ...WorkflowOption) error {
	if err := d.Registry.RegisterWorkflow(wf, opts...); err != nil {
		return err
	}
	if d.onChange != nil {
		d.onChange(d.GetSnapshot())
	}
	return nil
}

// RegisterActivity registers an activity and triggers onChange
func (d *DynamicRegistry) RegisterActivity(act interface{}, opts ...ActivityOption) error {
	if err := d.Registry.RegisterActivity(act, opts...); err != nil {
		return err
	}
	if d.onChange != nil {
		d.onChange(d.GetSnapshot())
	}
	return nil
}

// UnregisterWorkflow unregisters a workflow and triggers onChange
func (d *DynamicRegistry) UnregisterWorkflow(name string) error {
	if err := d.Registry.UnregisterWorkflow(name); err != nil {
		return err
	}
	if d.onChange != nil {
		d.onChange(d.GetSnapshot())
	}
	return nil
}

// UnregisterActivity unregisters an activity and triggers onChange
func (d *DynamicRegistry) UnregisterActivity(name string) error {
	if err := d.Registry.UnregisterActivity(name); err != nil {
		return err
	}
	if d.onChange != nil {
		d.onChange(d.GetSnapshot())
	}
	return nil
}