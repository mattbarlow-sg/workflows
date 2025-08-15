// Package workflows provides educational workflow implementation
package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// Checkpoint represents a workflow checkpoint
type Checkpoint struct {
	ID        string                 `json:"id"`
	Phase     string                 `json:"phase"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// EducationalWorkflow implements a long-running educational workflow with checkpoints
type EducationalWorkflow struct {
	*BaseWorkflowImpl
	currentModule     int
	checkpoints       []Checkpoint
	studentProgress   map[string]ModuleProgress
	assessmentResults []AssessmentResult
	startTime         time.Time
}

// EducationalInput defines the input for the educational workflow
type EducationalInput struct {
	CourseID         string                 `json:"course_id"`
	StudentID        string                 `json:"student_id"`
	CourseName       string                 `json:"course_name"`
	Modules          []Module               `json:"modules"`
	Settings         CourseSettings         `json:"settings"`
	Prerequisites    []string               `json:"prerequisites,omitempty"`
	Instructor       string                 `json:"instructor"`
	Deadline         *time.Time             `json:"deadline,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
}

// EducationalOutput defines the output of the educational workflow
type EducationalOutput struct {
	CourseID           string                       `json:"course_id"`
	StudentID          string                       `json:"student_id"`
	Status             EducationalWorkflowStatus    `json:"status"`
	CompletedModules   []string                     `json:"completed_modules"`
	CurrentModule      string                       `json:"current_module,omitempty"`
	Progress           CourseProgress               `json:"progress"`
	Assessments        []AssessmentResult           `json:"assessments"`
	Checkpoints        []Checkpoint                 `json:"checkpoints"`
	StartTime          time.Time                    `json:"start_time"`
	CompletionTime     *time.Time                   `json:"completion_time,omitempty"`
	TotalDuration      time.Duration                `json:"total_duration"`
	Grade              *Grade                       `json:"grade,omitempty"`
	Certificate        *Certificate                 `json:"certificate,omitempty"`
}

// Module represents a learning module
type Module struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Type             ModuleType             `json:"type"`
	EstimatedDuration time.Duration         `json:"estimated_duration"`
	Prerequisites    []string               `json:"prerequisites,omitempty"`
	Content          ModuleContent          `json:"content"`
	Assessment       *Assessment            `json:"assessment,omitempty"`
	Required         bool                   `json:"required"`
	Weight           float64                `json:"weight"` // For final grade calculation
}

// ModuleType represents the type of learning module
type ModuleType string

const (
	ModuleTypeLesson     ModuleType = "lesson"
	ModuleTypeVideo      ModuleType = "video"
	ModuleTypeLab        ModuleType = "lab"
	ModuleTypeQuiz       ModuleType = "quiz"
	ModuleTypeAssignment ModuleType = "assignment"
	ModuleTypeProject    ModuleType = "project"
	ModuleTypeDiscussion ModuleType = "discussion"
)

// ModuleContent represents the content of a module
type ModuleContent struct {
	Materials    []Material             `json:"materials"`
	Instructions string                 `json:"instructions"`
	Resources    []Resource             `json:"resources,omitempty"`
	Activities   []Activity             `json:"activities,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// Material represents learning material
type Material struct {
	ID       string       `json:"id"`
	Type     MaterialType `json:"type"`
	Title    string       `json:"title"`
	URL      string       `json:"url"`
	Duration *time.Duration `json:"duration,omitempty"`
	Required bool         `json:"required"`
}

// MaterialType represents the type of learning material
type MaterialType string

const (
	MaterialTypeDocument MaterialType = "document"
	MaterialTypeVideo    MaterialType = "video"
	MaterialTypeAudio    MaterialType = "audio"
	MaterialTypeLink     MaterialType = "link"
	MaterialTypeFile     MaterialType = "file"
)

// Resource represents additional learning resources
type Resource struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Activity represents a learning activity
type Activity struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         ActivityType           `json:"type"`
	Instructions string                 `json:"instructions"`
	TimeLimit    *time.Duration         `json:"time_limit,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeReading     ActivityType = "reading"
	ActivityTypeExercise    ActivityType = "exercise"
	ActivityTypePractice    ActivityType = "practice"
	ActivityTypeSubmission  ActivityType = "submission"
	ActivityTypeInteractive ActivityType = "interactive"
)

// Assessment represents an assessment
type Assessment struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Type            AssessmentType     `json:"type"`
	Questions       []Question         `json:"questions,omitempty"`
	TimeLimit       *time.Duration     `json:"time_limit,omitempty"`
	MaxAttempts     int                `json:"max_attempts"`
	PassingScore    float64            `json:"passing_score"`
	Weight          float64            `json:"weight"`
	Instructions    string             `json:"instructions"`
	RequiresProctor bool               `json:"requires_proctor"`
}

// AssessmentType represents the type of assessment
type AssessmentType string

const (
	AssessmentTypeQuiz       AssessmentType = "quiz"
	AssessmentTypeExam       AssessmentType = "exam"
	AssessmentTypeAssignment AssessmentType = "assignment"
	AssessmentTypePeer       AssessmentType = "peer_review"
	AssessmentTypeProject    AssessmentType = "project"
)

// Question represents an assessment question
type Question struct {
	ID          string                 `json:"id"`
	Type        QuestionType           `json:"type"`
	Text        string                 `json:"text"`
	Options     []string               `json:"options,omitempty"`
	CorrectAnswer interface{}          `json:"correct_answer,omitempty"`
	Points      float64                `json:"points"`
	Explanation string                 `json:"explanation,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// QuestionType represents the type of question
type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeTrueFalse      QuestionType = "true_false"
	QuestionTypeShortAnswer    QuestionType = "short_answer"
	QuestionTypeEssay          QuestionType = "essay"
	QuestionTypeCode           QuestionType = "code"
)

// CourseSettings defines course-level settings
type CourseSettings struct {
	AllowRetake            bool          `json:"allow_retake"`
	MaxRetakeAttempts      int           `json:"max_retake_attempts"`
	MinPassingGrade        float64       `json:"min_passing_grade"`
	ModuleTimeout          time.Duration `json:"module_timeout"`
	AssessmentTimeout      time.Duration `json:"assessment_timeout"`
	RequireSequentialOrder bool          `json:"require_sequential_order"`
	AllowSkipOptional      bool          `json:"allow_skip_optional"`
	SendReminders          bool          `json:"send_reminders"`
	ReminderIntervals      []time.Duration `json:"reminder_intervals,omitempty"`
	ProctorRequired        bool          `json:"proctor_required"`
}

// ModuleProgress tracks progress for a module
type ModuleProgress struct {
	ModuleID      string                 `json:"module_id"`
	Status        ModuleStatus           `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	TimeSpent     time.Duration          `json:"time_spent"`
	Attempts      int                    `json:"attempts"`
	Score         *float64               `json:"score,omitempty"`
	CompletedItems []string              `json:"completed_items"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// ModuleStatus represents the status of a module
type ModuleStatus string

const (
	ModuleStatusNotStarted ModuleStatus = "not_started"
	ModuleStatusInProgress ModuleStatus = "in_progress"
	ModuleStatusCompleted  ModuleStatus = "completed"
	ModuleStatusFailed     ModuleStatus = "failed"
	ModuleStatusSkipped    ModuleStatus = "skipped"
)

// CourseProgress tracks overall course progress
type CourseProgress struct {
	OverallPercent    float64                      `json:"overall_percent"`
	ModulesCompleted  int                          `json:"modules_completed"`
	ModulesTotal      int                          `json:"modules_total"`
	AssessmentsTotal  int                          `json:"assessments_total"`
	AssessmentsPassed int                          `json:"assessments_passed"`
	CurrentGrade      *float64                     `json:"current_grade,omitempty"`
	ModuleProgress    map[string]ModuleProgress    `json:"module_progress"`
	Milestones        []Milestone                  `json:"milestones,omitempty"`
}

// Milestone represents a course milestone
type Milestone struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Percent     int       `json:"percent"`
}

// AssessmentResult represents the result of an assessment
type AssessmentResult struct {
	AssessmentID   string                 `json:"assessment_id"`
	ModuleID       string                 `json:"module_id"`
	Attempt        int                    `json:"attempt"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	Duration       time.Duration          `json:"duration"`
	Score          float64                `json:"score"`
	MaxScore       float64                `json:"max_score"`
	Percentage     float64                `json:"percentage"`
	Passed         bool                   `json:"passed"`
	Answers        []Answer               `json:"answers,omitempty"`
	Feedback       string                 `json:"feedback,omitempty"`
	GradedBy       string                 `json:"graded_by,omitempty"`
	GradedAt       *time.Time             `json:"graded_at,omitempty"`
}

// Answer represents a student's answer to a question
type Answer struct {
	QuestionID string      `json:"question_id"`
	Answer     interface{} `json:"answer"`
	IsCorrect  *bool       `json:"is_correct,omitempty"`
	Points     *float64    `json:"points,omitempty"`
}

// Grade represents the final course grade
type Grade struct {
	Score      float64   `json:"score"`
	Letter     string    `json:"letter"`
	Passed     bool      `json:"passed"`
	GradedBy   string    `json:"graded_by"`
	GradedAt   time.Time `json:"graded_at"`
	Comments   string    `json:"comments,omitempty"`
}

// Certificate represents a course completion certificate
type Certificate struct {
	ID          string    `json:"id"`
	CourseID    string    `json:"course_id"`
	StudentID   string    `json:"student_id"`
	CourseName  string    `json:"course_name"`
	StudentName string    `json:"student_name"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Grade       Grade     `json:"grade"`
	URL         string    `json:"url"`
}

// EducationalWorkflowStatus represents the status of the educational workflow
type EducationalWorkflowStatus string

const (
	EducationalStatusEnrolled    EducationalWorkflowStatus = "enrolled"
	EducationalStatusInProgress  EducationalWorkflowStatus = "in_progress"
	EducationalStatusCompleted   EducationalWorkflowStatus = "completed"
	EducationalStatusFailed      EducationalWorkflowStatus = "failed"
	EducationalStatusWithdrawn   EducationalWorkflowStatus = "withdrawn"
	EducationalStatusExpired     EducationalWorkflowStatus = "expired"
)

// NewEducationalWorkflow creates a new educational workflow
func NewEducationalWorkflow() *EducationalWorkflow {
	metadata := WorkflowMetadata{
		Name:        "EducationalWorkflow",
		Version:     "1.0.0",
		Description: "Long-running educational workflow with checkpoints and assessments",
		Tags: map[string]string{
			"type":     "educational",
			"category": "long-running",
		},
		Author:    "Temporal Workflows",
		CreatedAt: time.Now(),
	}

	return &EducationalWorkflow{
		BaseWorkflowImpl:  NewBaseWorkflow(metadata),
		currentModule:     0,
		checkpoints:       []Checkpoint{},
		studentProgress:   make(map[string]ModuleProgress),
		assessmentResults: []AssessmentResult{},
	}
}

// Execute runs the educational workflow
func (ew *EducationalWorkflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	// Type assertion for input
	eduInput, ok := input.(EducationalInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for EducationalWorkflow")
	}

	// Validate input
	if err := ew.Validate(eduInput); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting educational workflow", 
		"courseID", eduInput.CourseID, 
		"studentID", eduInput.StudentID, 
		"moduleCount", len(eduInput.Modules))

	ew.startTime = workflow.Now(ctx)
	ew.UpdateState("enrollment", 0, WorkflowStatusRunning)

	// Setup base queries and signals
	if err := ew.SetupBaseQueries(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup base queries: %w", err)
	}
	ew.SetupBaseSignals(ctx)

	// Setup educational-specific queries and signals
	ew.setupEducationalQueries(ctx)
	moduleCompleteSignal, withdrawSignal := ew.setupEducationalSignals(ctx)

	// Configure activity options for long-running educational activities
	ao := LongRunningActivityOptions()
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Initialize student enrollment
	if err := ew.enrollStudent(ctx, eduInput); err != nil {
		ew.SetError(err)
		return ew.createOutput(eduInput, EducationalStatusFailed), err
	}

	// Create initial checkpoint
	ew.createCheckpoint("enrollment", "Student enrolled in course", nil)

	// Process prerequisites if any
	if len(eduInput.Prerequisites) > 0 {
		if err := ew.checkPrerequisites(ctx, eduInput); err != nil {
			ew.SetError(err)
			return ew.createOutput(eduInput, EducationalStatusFailed), err
		}
	}

	// Main learning loop - process each module
	for moduleIndex, module := range eduInput.Modules {
		ew.currentModule = moduleIndex
		progress := (moduleIndex * 100) / len(eduInput.Modules)
		ew.UpdateState(fmt.Sprintf("module_%s", module.ID), progress, WorkflowStatusRunning)

		logger.Info("Starting module", "moduleID", module.ID, "moduleName", module.Name)

		// Check module prerequisites
		if !ew.checkModulePrerequisites(module) {
			if module.Required {
				err := fmt.Errorf("module prerequisites not met: %s", module.ID)
				ew.SetError(err)
				return ew.createOutput(eduInput, EducationalStatusFailed), err
			}
			// Skip optional module
			ew.skipModule(module.ID, "Prerequisites not met")
			continue
		}

		// Process the module
		moduleResult, err := ew.processModule(ctx, eduInput, module, moduleCompleteSignal, withdrawSignal)
		if err != nil {
			logger.Error("Module processing failed", "moduleID", module.ID, "error", err)
			if module.Required {
				ew.SetError(err)
				return ew.createOutput(eduInput, EducationalStatusFailed), err
			}
			// Mark as failed but continue with non-required modules
			ew.studentProgress[module.ID] = ModuleProgress{
				ModuleID:  module.ID,
				Status:    ModuleStatusFailed,
				StartTime: workflow.Now(ctx),
				EndTime:   &[]time.Time{workflow.Now(ctx)}[0],
				Attempts:  1,
			}
			continue
		}

		// Check if student withdrew
		if moduleResult.Status == EducationalStatusWithdrawn {
			logger.Info("Student withdrew from course")
			return ew.createOutput(eduInput, EducationalStatusWithdrawn), nil
		}

		// Create checkpoint after each module
		ew.createCheckpoint(fmt.Sprintf("module_%s_completed", module.ID), 
			fmt.Sprintf("Completed module: %s", module.Name), 
			map[string]interface{}{
				"moduleID": module.ID,
				"score":    moduleResult.Score,
			})

		// Check deadline if set
		if eduInput.Deadline != nil && workflow.Now(ctx).After(*eduInput.Deadline) {
			logger.Warn("Course deadline exceeded")
			return ew.createOutput(eduInput, EducationalStatusExpired), nil
		}

		// Send progress update
		ew.sendProgressUpdate(ctx, eduInput, progress)
	}

	// All modules completed - calculate final grade
	finalGrade, err := ew.calculateFinalGrade(ctx, eduInput)
	if err != nil {
		logger.Error("Failed to calculate final grade", "error", err)
		ew.SetError(err)
		return ew.createOutput(eduInput, EducationalStatusFailed), err
	}

	// Check if passed
	passed := finalGrade.Score >= eduInput.Settings.MinPassingGrade
	status := EducationalStatusCompleted
	if !passed {
		status = EducationalStatusFailed
	}

	// Generate certificate if passed
	var certificate *Certificate
	if passed {
		certificate, err = ew.generateCertificate(ctx, eduInput, *finalGrade)
		if err != nil {
			logger.Warn("Failed to generate certificate", "error", err)
		}
	}

	// Final checkpoint
	ew.createCheckpoint("course_completed", "Course completed", map[string]interface{}{
		"finalGrade": finalGrade,
		"passed":     passed,
		"certificate": certificate,
	})

	completionTime := workflow.Now(ctx)
	ew.UpdateState("completed", 100, WorkflowStatusCompleted)

	logger.Info("Educational workflow completed", 
		"status", status, 
		"finalGrade", finalGrade.Score, 
		"passed", passed,
		"duration", completionTime.Sub(ew.startTime))

	output := ew.createOutput(eduInput, status)
	output.Grade = finalGrade
	output.Certificate = certificate
	output.CompletionTime = &completionTime

	return output, nil
}

// Validate validates the educational input
func (ew *EducationalWorkflow) Validate(input interface{}) error {
	eduInput, ok := input.(EducationalInput)
	if !ok {
		return fmt.Errorf("input must be of type EducationalInput")
	}

	if eduInput.CourseID == "" {
		return fmt.Errorf("course ID is required")
	}

	if eduInput.StudentID == "" {
		return fmt.Errorf("student ID is required")
	}

	if eduInput.CourseName == "" {
		return fmt.Errorf("course name is required")
	}

	if len(eduInput.Modules) == 0 {
		return fmt.Errorf("at least one module is required")
	}

	if eduInput.Instructor == "" {
		return fmt.Errorf("instructor is required")
	}

	// Validate modules
	moduleIDs := make(map[string]bool)
	for i, module := range eduInput.Modules {
		if module.ID == "" {
			return fmt.Errorf("module %d ID is required", i)
		}

		if moduleIDs[module.ID] {
			return fmt.Errorf("duplicate module ID: %s", module.ID)
		}
		moduleIDs[module.ID] = true

		if module.Name == "" {
			return fmt.Errorf("module %s name is required", module.ID)
		}

		if module.EstimatedDuration <= 0 {
			return fmt.Errorf("module %s estimated duration must be positive", module.ID)
		}

		// Validate prerequisites exist
		for _, prereqID := range module.Prerequisites {
			if !moduleIDs[prereqID] && prereqID != module.ID {
				// Check if it's a course prerequisite
				found := false
				for _, coursePrereq := range eduInput.Prerequisites {
					if coursePrereq == prereqID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("module %s has invalid prerequisite: %s", module.ID, prereqID)
				}
			}
		}
	}

	// Validate settings
	if eduInput.Settings.MinPassingGrade < 0 || eduInput.Settings.MinPassingGrade > 100 {
		return fmt.Errorf("minimum passing grade must be between 0 and 100")
	}

	if eduInput.Settings.ModuleTimeout <= 0 {
		return fmt.Errorf("module timeout must be positive")
	}

	return nil
}

// enrollStudent enrolls the student in the course
func (ew *EducationalWorkflow) enrollStudent(ctx workflow.Context, input EducationalInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Enrolling student", "studentID", input.StudentID, "courseID", input.CourseID)

	enrollmentData := map[string]interface{}{
		"courseID":   input.CourseID,
		"studentID":  input.StudentID,
		"instructor": input.Instructor,
		"enrolledAt": workflow.Now(ctx),
	}

	return workflow.ExecuteActivity(ctx, EnrollStudentActivity, enrollmentData).Get(ctx, nil)
}

// checkPrerequisites checks if student has met course prerequisites
func (ew *EducationalWorkflow) checkPrerequisites(ctx workflow.Context, input EducationalInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Checking prerequisites", "prerequisites", input.Prerequisites)

	prereqCheck := map[string]interface{}{
		"studentID":     input.StudentID,
		"prerequisites": input.Prerequisites,
	}

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, CheckPrerequisitesActivity, prereqCheck).Get(ctx, &result)
	if err != nil {
		return fmt.Errorf("prerequisite check failed: %w", err)
	}

	if met, ok := result["met"].(bool); !ok || !met {
		return fmt.Errorf("prerequisites not met")
	}

	return nil
}

// checkModulePrerequisites checks if module prerequisites are met
func (ew *EducationalWorkflow) checkModulePrerequisites(module Module) bool {
	for _, prereqID := range module.Prerequisites {
		if progress, exists := ew.studentProgress[prereqID]; !exists || progress.Status != ModuleStatusCompleted {
			return false
		}
	}
	return true
}

// processModule processes a single module
func (ew *EducationalWorkflow) processModule(ctx workflow.Context, input EducationalInput, module Module,
	moduleCompleteSignal, withdrawSignal workflow.ReceiveChannel) (*ModuleResult, error) {

	logger := workflow.GetLogger(ctx)
	startTime := workflow.Now(ctx)

	// Initialize module progress
	progress := ModuleProgress{
		ModuleID:       module.ID,
		Status:         ModuleStatusInProgress,
		StartTime:      startTime,
		TimeSpent:      0,
		Attempts:       1,
		CompletedItems: []string{},
	}
	ew.studentProgress[module.ID] = progress

	// Start module activity
	moduleData := map[string]interface{}{
		"moduleID":   module.ID,
		"studentID":  input.StudentID,
		"courseID":   input.CourseID,
		"module":     module,
		"startTime":  startTime,
	}

	var moduleActivityResult map[string]interface{}
	err := workflow.ExecuteActivity(ctx, StartModuleActivity, moduleData).Get(ctx, &moduleActivityResult)
	if err != nil {
		return nil, fmt.Errorf("failed to start module: %w", err)
	}

	// Wait for module completion or withdrawal
	for {
		selector := workflow.NewSelector(ctx)
		
		// Module completion handler
		selector.AddReceive(moduleCompleteSignal, func(c workflow.ReceiveChannel, more bool) {
			var completion ModuleCompletionSignal
			c.Receive(ctx, &completion)
			
			if completion.ModuleID == module.ID {
				endTime := workflow.Now(ctx)
				
				// Update progress
				progress.Status = ModuleStatusCompleted
				progress.EndTime = &endTime
				progress.TimeSpent = endTime.Sub(startTime)
				progress.Score = &completion.Score
				progress.CompletedItems = completion.CompletedItems
				ew.studentProgress[module.ID] = progress
				
				logger.Info("Module completed", "moduleID", module.ID, "score", completion.Score)
			}
		})
		
		// Withdrawal handler
		selector.AddReceive(withdrawSignal, func(c workflow.ReceiveChannel, more bool) {
			var reason string
			c.Receive(ctx, &reason)
			logger.Info("Student withdrew", "reason", reason)
		})
		
		// Timeout handler
		timer := workflow.NewTimer(ctx, input.Settings.ModuleTimeout)
		selector.AddFuture(timer, func(f workflow.Future) {
			logger.Warn("Module timeout", "moduleID", module.ID)
			// Handle timeout - could escalate or mark as failed
		})
		
		selector.Select(ctx)
		
		// Check if we should break out of the loop
		currentProgress := ew.studentProgress[module.ID]
		if currentProgress.Status == ModuleStatusCompleted || currentProgress.Status == ModuleStatusFailed {
			break
		}
		
		// Check for withdrawal
		var withdrawn bool
		withdrawSignal.ReceiveAsync(&withdrawn)
		if withdrawn {
			return &ModuleResult{Status: EducationalStatusWithdrawn}, nil
		}
	}

	// Process assessment if module has one
	var assessmentScore *float64
	if module.Assessment != nil {
		score, err := ew.processAssessment(ctx, input, module, *module.Assessment)
		if err != nil {
			logger.Error("Assessment failed", "moduleID", module.ID, "error", err)
			if module.Required {
				return nil, err
			}
		} else {
			assessmentScore = &score
		}
	}

	result := &ModuleResult{
		Status:    EducationalStatusInProgress,
		ModuleID:  module.ID,
		Score:     assessmentScore,
		Completed: true,
	}

	return result, nil
}

// processAssessment processes a module assessment
func (ew *EducationalWorkflow) processAssessment(ctx workflow.Context, input EducationalInput, 
	module Module, assessment Assessment) (float64, error) {
	
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting assessment", "assessmentID", assessment.ID, "moduleID", module.ID)

	assessmentData := map[string]interface{}{
		"assessmentID": assessment.ID,
		"moduleID":     module.ID,
		"studentID":    input.StudentID,
		"assessment":   assessment,
		"attempt":      1,
	}

	var result AssessmentResult
	err := workflow.ExecuteActivity(ctx, ProcessAssessmentActivity, assessmentData).Get(ctx, &result)
	if err != nil {
		return 0, fmt.Errorf("assessment processing failed: %w", err)
	}

	ew.assessmentResults = append(ew.assessmentResults, result)
	
	if !result.Passed && assessment.MaxAttempts > 1 {
		// Allow retakes
		for attempt := 2; attempt <= assessment.MaxAttempts; attempt++ {
			logger.Info("Retaking assessment", "attempt", attempt, "assessmentID", assessment.ID)
			
			assessmentData["attempt"] = attempt
			err = workflow.ExecuteActivity(ctx, ProcessAssessmentActivity, assessmentData).Get(ctx, &result)
			if err != nil {
				continue
			}
			
			ew.assessmentResults = append(ew.assessmentResults, result)
			
			if result.Passed {
				break
			}
		}
	}

	logger.Info("Assessment completed", "score", result.Score, "passed", result.Passed)
	return result.Score, nil
}

// calculateFinalGrade calculates the final course grade
func (ew *EducationalWorkflow) calculateFinalGrade(ctx workflow.Context, input EducationalInput) (*Grade, error) {
	logger := workflow.GetLogger(ctx)
	
	var totalScore float64
	var totalWeight float64
	
	// Calculate weighted average based on module scores
	for _, module := range input.Modules {
		if progress, exists := ew.studentProgress[module.ID]; exists && progress.Score != nil {
			totalScore += *progress.Score * module.Weight
			totalWeight += module.Weight
		}
	}
	
	// Add assessment scores
	for _, result := range ew.assessmentResults {
		// Find the assessment weight
		for _, module := range input.Modules {
			if module.ID == result.ModuleID && module.Assessment != nil {
				totalScore += result.Score * module.Assessment.Weight
				totalWeight += module.Assessment.Weight
			}
		}
	}
	
	var finalScore float64
	if totalWeight > 0 {
		finalScore = totalScore / totalWeight
	}
	
	// Convert to letter grade
	letterGrade := ew.calculateLetterGrade(finalScore)
	passed := finalScore >= input.Settings.MinPassingGrade
	
	grade := &Grade{
		Score:    finalScore,
		Letter:   letterGrade,
		Passed:   passed,
		GradedBy: input.Instructor,
		GradedAt: workflow.Now(ctx),
	}
	
	logger.Info("Final grade calculated", "score", finalScore, "letter", letterGrade, "passed", passed)
	return grade, nil
}

// calculateLetterGrade converts numeric score to letter grade
func (ew *EducationalWorkflow) calculateLetterGrade(score float64) string {
	switch {
	case score >= 97:
		return "A+"
	case score >= 93:
		return "A"
	case score >= 90:
		return "A-"
	case score >= 87:
		return "B+"
	case score >= 83:
		return "B"
	case score >= 80:
		return "B-"
	case score >= 77:
		return "C+"
	case score >= 73:
		return "C"
	case score >= 70:
		return "C-"
	case score >= 67:
		return "D+"
	case score >= 63:
		return "D"
	case score >= 60:
		return "D-"
	default:
		return "F"
	}
}

// generateCertificate generates a completion certificate
func (ew *EducationalWorkflow) generateCertificate(ctx workflow.Context, input EducationalInput, grade Grade) (*Certificate, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Generating certificate", "studentID", input.StudentID, "courseID", input.CourseID)

	certData := map[string]interface{}{
		"courseID":   input.CourseID,
		"studentID":  input.StudentID,
		"courseName": input.CourseName,
		"grade":      grade,
		"issuedAt":   workflow.Now(ctx),
	}

	var certificate Certificate
	err := workflow.ExecuteActivity(ctx, GenerateCertificateActivity, certData).Get(ctx, &certificate)
	if err != nil {
		return nil, fmt.Errorf("certificate generation failed: %w", err)
	}

	return &certificate, nil
}

// Helper methods

func (ew *EducationalWorkflow) skipModule(moduleID, reason string) {
	progress := ModuleProgress{
		ModuleID:  moduleID,
		Status:    ModuleStatusSkipped,
		StartTime: time.Now(),
		EndTime:   &[]time.Time{time.Now()}[0],
		Data:      map[string]interface{}{"skipReason": reason},
	}
	ew.studentProgress[moduleID] = progress
}

func (ew *EducationalWorkflow) createCheckpoint(id, description string, data map[string]interface{}) {
	checkpoint := Checkpoint{
		ID:        id,
		Phase:     description,
		Timestamp: time.Now(),
		Data:      data,
	}
	ew.checkpoints = append(ew.checkpoints, checkpoint)
}

func (ew *EducationalWorkflow) sendProgressUpdate(ctx workflow.Context, input EducationalInput, progress int) {
	updateData := map[string]interface{}{
		"studentID": input.StudentID,
		"courseID":  input.CourseID,
		"progress":  progress,
		"timestamp": workflow.Now(ctx),
	}
	
	// Send progress update (fire-and-forget)
	workflow.ExecuteActivity(ctx, SendProgressUpdateActivity, updateData)
}

func (ew *EducationalWorkflow) createOutput(input EducationalInput, status EducationalWorkflowStatus) EducationalOutput {
	var completedModules []string
	for moduleID, progress := range ew.studentProgress {
		if progress.Status == ModuleStatusCompleted {
			completedModules = append(completedModules, moduleID)
		}
	}

	var currentModule string
	if ew.currentModule < len(input.Modules) {
		currentModule = input.Modules[ew.currentModule].ID
	}

	courseProgress := CourseProgress{
		ModulesCompleted: len(completedModules),
		ModulesTotal:     len(input.Modules),
		ModuleProgress:   ew.studentProgress,
	}

	if courseProgress.ModulesTotal > 0 {
		courseProgress.OverallPercent = float64(courseProgress.ModulesCompleted*100) / float64(courseProgress.ModulesTotal)
	}

	return EducationalOutput{
		CourseID:         input.CourseID,
		StudentID:        input.StudentID,
		Status:           status,
		CompletedModules: completedModules,
		CurrentModule:    currentModule,
		Progress:         courseProgress,
		Assessments:      ew.assessmentResults,
		Checkpoints:      ew.checkpoints,
		StartTime:        ew.startTime,
		TotalDuration:    time.Since(ew.startTime),
	}
}

// Query and signal setup

func (ew *EducationalWorkflow) setupEducationalQueries(ctx workflow.Context) {
	workflow.SetQueryHandler(ctx, "getProgress", func() (CourseProgress, error) {
		return ew.createOutput(EducationalInput{}, EducationalStatusInProgress).Progress, nil
	})
	
	workflow.SetQueryHandler(ctx, "getCurrentModule", func() (string, error) {
		if ew.currentModule < len(ew.studentProgress) {
			// This would need the actual input to get module ID
			return fmt.Sprintf("module_%d", ew.currentModule), nil
		}
		return "", nil
	})
	
	workflow.SetQueryHandler(ctx, "getAssessments", func() ([]AssessmentResult, error) {
		return ew.assessmentResults, nil
	})
	
	workflow.SetQueryHandler(ctx, "getCheckpoints", func() ([]Checkpoint, error) {
		return ew.checkpoints, nil
	})
}

func (ew *EducationalWorkflow) setupEducationalSignals(ctx workflow.Context) (workflow.ReceiveChannel, workflow.ReceiveChannel) {
	moduleCompleteSignal := workflow.GetSignalChannel(ctx, "moduleComplete")
	withdrawSignal := workflow.GetSignalChannel(ctx, "withdraw")
	
	return moduleCompleteSignal, withdrawSignal
}

// Support types

type ModuleResult struct {
	Status    EducationalWorkflowStatus
	ModuleID  string
	Score     *float64
	Completed bool
}

type ModuleCompletionSignal struct {
	ModuleID       string   `json:"module_id"`
	Score          float64  `json:"score"`
	CompletedItems []string `json:"completed_items"`
	Notes          string   `json:"notes,omitempty"`
}

// Activity stubs (would be implemented with real functionality)

func EnrollStudentActivity(ctx context.Context, data map[string]interface{}) error {
	// Implementation would enroll student in external system
	return nil
}

func CheckPrerequisitesActivity(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// Implementation would check prerequisites in external system
	return map[string]interface{}{"met": true}, nil
}

func StartModuleActivity(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// Implementation would start module in external system
	return map[string]interface{}{"started": true}, nil
}

func ProcessAssessmentActivity(ctx context.Context, data map[string]interface{}) (AssessmentResult, error) {
	// Implementation would process assessment in external system
	return AssessmentResult{
		Score:      85.0,
		MaxScore:   100.0,
		Percentage: 85.0,
		Passed:     true,
	}, nil
}

func GenerateCertificateActivity(ctx context.Context, data map[string]interface{}) (Certificate, error) {
	// Implementation would generate certificate in external system
	return Certificate{
		ID:        "cert-123",
		IssuedAt:  time.Now(),
		URL:       "https://example.com/certificate/cert-123",
	}, nil
}

func SendProgressUpdateActivity(ctx context.Context, data map[string]interface{}) error {
	// Implementation would send progress update to external system
	return nil
}