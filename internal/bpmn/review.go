package bpmn

import (
	"fmt"
	"math/rand"
	"time"
)

// ReviewType defines different types of review workflows
type ReviewType string

const (
	// ReviewTypePeer requires peer review from same level
	ReviewTypePeer ReviewType = "peer_review"

	// ReviewTypeHierarchical requires review from higher level
	ReviewTypeHierarchical ReviewType = "hierarchical"

	// ReviewTypeSpecialist requires review from domain expert
	ReviewTypeSpecialist ReviewType = "specialist"

	// ReviewTypeCollaborative requires multiple reviewers
	ReviewTypeCollaborative ReviewType = "collaborative"

	// ReviewTypeAutomated uses automated rules
	ReviewTypeAutomated ReviewType = "automated"
)

// ReviewPattern defines the review workflow pattern
type ReviewPattern string

const (
	// PatternSequential reviews happen one after another
	PatternSequential ReviewPattern = "sequential"

	// PatternParallel reviews happen simultaneously
	PatternParallel ReviewPattern = "parallel"

	// PatternIterative allows for revision cycles
	PatternIterative ReviewPattern = "iterative"
)

// ReviewStatus tracks the status of a review
type ReviewStatus string

const (
	// StatusPending review not yet started
	StatusPending ReviewStatus = "pending"

	// StatusInProgress review underway
	StatusInProgress ReviewStatus = "in_progress"

	// StatusApproved review approved
	StatusApproved ReviewStatus = "approved"

	// StatusRejected review rejected
	StatusRejected ReviewStatus = "rejected"

	// StatusRevisionRequired needs changes
	StatusRevisionRequired ReviewStatus = "revision_required"
)

// ReviewRequest represents a request for review
type ReviewRequest struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	ActivityID    string                 `json:"activity_id"`
	SubmitterID   string                 `json:"submitter_id"`
	ReviewerID    string                 `json:"reviewer_id"`
	Status        ReviewStatus           `json:"status"`
	SubmittedAt   time.Time              `json:"submitted_at"`
	ReviewedAt    *time.Time             `json:"reviewed_at,omitempty"`
	Data          map[string]interface{} `json:"data"`
	Comments      []ReviewComment        `json:"comments"`
	RevisionCount int                    `json:"revision_count"`
}

// ReviewComment represents a comment in the review process
type ReviewComment struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "comment", "approval", "rejection", "suggestion"
}

// ReviewQueue manages pending reviews
type ReviewQueue struct {
	requests   map[string]*ReviewRequest
	byReviewer map[string][]*ReviewRequest
	byStatus   map[ReviewStatus][]*ReviewRequest
}

// NewReviewQueue creates a new review queue
func NewReviewQueue() *ReviewQueue {
	return &ReviewQueue{
		requests:   make(map[string]*ReviewRequest),
		byReviewer: make(map[string][]*ReviewRequest),
		byStatus:   make(map[ReviewStatus][]*ReviewRequest),
	}
}

// AddRequest adds a review request to the queue
func (rq *ReviewQueue) AddRequest(request *ReviewRequest) {
	rq.requests[request.ID] = request

	// Update indices
	rq.byReviewer[request.ReviewerID] = append(rq.byReviewer[request.ReviewerID], request)
	rq.byStatus[request.Status] = append(rq.byStatus[request.Status], request)
}

// GetRequestsByReviewer returns all requests for a reviewer
func (rq *ReviewQueue) GetRequestsByReviewer(reviewerID string) []*ReviewRequest {
	return rq.byReviewer[reviewerID]
}

// GetRequestsByStatus returns all requests with a status
func (rq *ReviewQueue) GetRequestsByStatus(status ReviewStatus) []*ReviewRequest {
	return rq.byStatus[status]
}

// UpdateRequestStatus updates the status of a request
func (rq *ReviewQueue) UpdateRequestStatus(requestID string, newStatus ReviewStatus) error {
	request, exists := rq.requests[requestID]
	if !exists {
		return fmt.Errorf("review request %s not found", requestID)
	}

	// Remove from old status index
	oldRequests := rq.byStatus[request.Status]
	for i, r := range oldRequests {
		if r.ID == requestID {
			rq.byStatus[request.Status] = append(oldRequests[:i], oldRequests[i+1:]...)
			break
		}
	}

	// Update status and add to new index
	request.Status = newStatus
	rq.byStatus[newStatus] = append(rq.byStatus[newStatus], request)

	// Update timestamp if completed
	if newStatus == StatusApproved || newStatus == StatusRejected {
		now := time.Now()
		request.ReviewedAt = &now
	}

	return nil
}

// ReviewEngine processes review workflows
type ReviewEngine struct {
	queue     *ReviewQueue
	workflows map[string]*ReviewWorkflow
	handlers  map[ReviewType]ReviewHandler
}

// ReviewHandler handles specific review types
type ReviewHandler interface {
	ProcessReview(request *ReviewRequest, workflow *ReviewWorkflow) (*ReviewResult, error)
	ValidateReviewer(reviewerID string, workflow *ReviewWorkflow) bool
}

// NewReviewEngine creates a new review engine
func NewReviewEngine() *ReviewEngine {
	return &ReviewEngine{
		queue:     NewReviewQueue(),
		workflows: make(map[string]*ReviewWorkflow),
		handlers:  make(map[ReviewType]ReviewHandler),
	}
}

// RegisterWorkflow registers a review workflow
func (re *ReviewEngine) RegisterWorkflow(workflow *ReviewWorkflow) {
	re.workflows[workflow.ID] = workflow
}

// RegisterHandler registers a review handler
func (re *ReviewEngine) RegisterHandler(reviewType ReviewType, handler ReviewHandler) {
	re.handlers[reviewType] = handler
}

// SubmitForReview submits an item for review
func (re *ReviewEngine) SubmitForReview(workflowID, activityID, submitterID string, data map[string]interface{}) (*ReviewRequest, error) {
	workflow, exists := re.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	// Create review request
	request := &ReviewRequest{
		ID:          generateID(),
		WorkflowID:  workflowID,
		ActivityID:  activityID,
		SubmitterID: submitterID,
		Status:      StatusPending,
		SubmittedAt: time.Now(),
		Data:        data,
		Comments:    []ReviewComment{},
	}

	// Assign reviewer based on workflow rules
	reviewerID, err := re.assignReviewer(workflow, data)
	if err != nil {
		return nil, err
	}
	request.ReviewerID = reviewerID

	// Add to queue
	re.queue.AddRequest(request)

	return request, nil
}

// ProcessReview processes a review request
func (re *ReviewEngine) ProcessReview(requestID string) (*ReviewResult, error) {
	request, exists := re.queue.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("request %s not found", requestID)
	}

	workflow, exists := re.workflows[request.WorkflowID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", request.WorkflowID)
	}

	// Get appropriate handler
	handler, exists := re.handlers[ReviewType(workflow.Type)]
	if !exists {
		return nil, fmt.Errorf("no handler for review type %s", workflow.Type)
	}

	// Update status
	re.queue.UpdateRequestStatus(requestID, StatusInProgress)

	// Process review
	result, err := handler.ProcessReview(request, workflow)
	if err != nil {
		return nil, err
	}

	// Update final status
	if result.Approved {
		re.queue.UpdateRequestStatus(requestID, StatusApproved)
	} else {
		re.queue.UpdateRequestStatus(requestID, StatusRejected)
	}

	return result, nil
}

// AddComment adds a comment to a review request
func (re *ReviewEngine) AddComment(requestID, authorID, text, commentType string) error {
	request, exists := re.queue.requests[requestID]
	if !exists {
		return fmt.Errorf("request %s not found", requestID)
	}

	comment := ReviewComment{
		ID:        generateID(),
		AuthorID:  authorID,
		Text:      text,
		Timestamp: time.Now(),
		Type:      commentType,
	}

	request.Comments = append(request.Comments, comment)
	return nil
}

// assignReviewer assigns a reviewer based on workflow rules
func (re *ReviewEngine) assignReviewer(workflow *ReviewWorkflow, data map[string]interface{}) (string, error) {
	// In a real implementation, this would use the agent manager
	// to select an appropriate reviewer
	if workflow.ReviewerID != "" {
		return workflow.ReviewerID, nil
	}

	// For now, return a placeholder
	return "default-reviewer", nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int31())
}

// Built-in Review Handlers

// PeerReviewHandler handles peer reviews
type PeerReviewHandler struct {
	agentManager *AgentManager
}

func NewPeerReviewHandler(am *AgentManager) *PeerReviewHandler {
	return &PeerReviewHandler{agentManager: am}
}

func (h *PeerReviewHandler) ProcessReview(request *ReviewRequest, workflow *ReviewWorkflow) (*ReviewResult, error) {
	// Implement peer review logic
	result := &ReviewResult{
		ReviewerID: request.ReviewerID,
		Timestamp:  time.Now(),
		ReviewData: request.Data,
	}

	// Apply review rules
	if workflow.Rules != nil {
		// Check required fields
		for _, field := range workflow.Rules.RequiredFields {
			if _, exists := request.Data[field]; !exists {
				result.Approved = false
				result.Comments = fmt.Sprintf("Missing required field: %s", field)
				return result, nil
			}
		}

		// Check approval threshold
		if workflow.Rules.ApprovalThreshold > 0 {
			if score, ok := request.Data["score"].(float64); ok {
				result.Approved = score >= workflow.Rules.ApprovalThreshold
				if !result.Approved {
					result.Comments = fmt.Sprintf("Score %.2f below threshold %.2f",
						score, workflow.Rules.ApprovalThreshold)
				}
			}
		}
	}

	// Default to approved if no rules failed
	if result.Comments == "" {
		result.Approved = true
		result.Comments = "Peer review completed successfully"
	}

	return result, nil
}

func (h *PeerReviewHandler) ValidateReviewer(reviewerID string, workflow *ReviewWorkflow) bool {
	// Check if reviewer has peer review capability
	// In real implementation, would check against agent manager
	return true
}

// AutomatedReviewHandler handles automated reviews
type AutomatedReviewHandler struct{}

func NewAutomatedReviewHandler() *AutomatedReviewHandler {
	return &AutomatedReviewHandler{}
}

func (h *AutomatedReviewHandler) ProcessReview(request *ReviewRequest, workflow *ReviewWorkflow) (*ReviewResult, error) {
	result := &ReviewResult{
		ReviewerID: "automated",
		Timestamp:  time.Now(),
		ReviewData: request.Data,
		Approved:   true,
	}

	// Apply automated rules
	if workflow.Rules != nil && workflow.Rules.CustomRules != nil {
		for key, value := range workflow.Rules.CustomRules {
			// Simple rule evaluation
			if dataValue, exists := request.Data[key]; exists {
				if dataValue != value {
					result.Approved = false
					result.Comments = fmt.Sprintf("Rule failed: %s expected %v, got %v",
						key, value, dataValue)
					break
				}
			}
		}
	}

	if result.Approved {
		result.Comments = "Automated review passed all rules"
	}

	return result, nil
}

func (h *AutomatedReviewHandler) ValidateReviewer(reviewerID string, workflow *ReviewWorkflow) bool {
	// Automated reviews don't need reviewer validation
	return true
}
