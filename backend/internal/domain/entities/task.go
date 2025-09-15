package entities

import (
	"strings"
	"time"

	"ai-git-workbench/internal/domain/errors"
	"ai-git-workbench/internal/domain/valueobjects"
)

// Task represents a workflow task in the domain
type Task struct {
	// Identity
	id     valueobjects.TaskID
	userID valueobjects.UserID

	// Basic properties
	title       string
	description string
	status      valueobjects.TaskStatus

	// Repository context
	repository valueobjects.RepositoryPath
	epic       string
	branch     valueobjects.BranchName

	// Tracking
	tokensUsed int

	// Timestamps
	createdAt   time.Time
	updatedAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time

	// Metadata
	metadata map[string]string

	// Concurrency control
	version int64
}

// NewTask creates a new task with business validation
func NewTask(
	id valueobjects.TaskID,
	userID valueobjects.UserID,
	title string,
	description string,
	repository valueobjects.RepositoryPath,
	epic string,
	branch valueobjects.BranchName,
) (*Task, error) {
	// Validate title
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	// Validate description
	if err := validateDescription(description); err != nil {
		return nil, err
	}

	// Validate epic
	if err := validateEpic(epic); err != nil {
		return nil, err
	}

	now := time.Now()
	
	task := &Task{
		id:          id,
		userID:      userID,
		title:       strings.TrimSpace(title),
		description: strings.TrimSpace(description),
		status:      valueobjects.NewPendingStatus(),
		repository:  repository,
		epic:        strings.TrimSpace(epic),
		branch:      branch,
		tokensUsed:  0,
		createdAt:   now,
		updatedAt:   now,
		metadata:    make(map[string]string),
		version:     1,
	}

	return task, nil
}

// CreateTask is a factory method that generates a new TaskID
func CreateTask(
	userID valueobjects.UserID,
	title string,
	description string,
	repository valueobjects.RepositoryPath,
	epic string,
	branch valueobjects.BranchName,
) (*Task, error) {
	taskID := valueobjects.GenerateTaskID()
	return NewTask(taskID, userID, title, description, repository, epic, branch)
}

// Getters
func (t *Task) ID() valueobjects.TaskID {
	return t.id
}

func (t *Task) UserID() valueobjects.UserID {
	return t.userID
}

func (t *Task) Title() string {
	return t.title
}

func (t *Task) Description() string {
	return t.description
}

func (t *Task) Status() valueobjects.TaskStatus {
	return t.status
}

func (t *Task) Repository() valueobjects.RepositoryPath {
	return t.repository
}

func (t *Task) Epic() string {
	return t.epic
}

func (t *Task) Branch() valueobjects.BranchName {
	return t.branch
}

func (t *Task) TokensUsed() int {
	return t.tokensUsed
}

func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Task) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Task) StartedAt() *time.Time {
	return t.startedAt
}

func (t *Task) CompletedAt() *time.Time {
	return t.completedAt
}

func (t *Task) Version() int64 {
	return t.version
}

func (t *Task) Metadata() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range t.metadata {
		result[k] = v
	}
	return result
}

// Business methods

// StartProcessing transitions the task to processing status
func (t *Task) StartProcessing() error {
	newStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusProcessing)
	if err != nil {
		return err
	}

	if !t.status.CanTransitionTo(newStatus) {
		return errors.NewInvalidStatusTransitionError(t.status.Value(), newStatus.Value())
	}

	t.status = newStatus
	now := time.Now()
	t.startedAt = &now
	t.updatedAt = now
	t.version++

	return nil
}

// Complete marks the task as completed
func (t *Task) Complete() error {
	newStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusCompleted)
	if err != nil {
		return err
	}

	if !t.status.CanTransitionTo(newStatus) {
		return errors.NewInvalidStatusTransitionError(t.status.Value(), newStatus.Value())
	}

	t.status = newStatus
	now := time.Now()
	t.completedAt = &now
	t.updatedAt = now
	t.version++

	return nil
}

// Fail marks the task as failed
func (t *Task) Fail() error {
	newStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusFailed)
	if err != nil {
		return err
	}

	if !t.status.CanTransitionTo(newStatus) {
		return errors.NewInvalidStatusTransitionError(t.status.Value(), newStatus.Value())
	}

	t.status = newStatus
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// Cancel marks the task as cancelled
func (t *Task) Cancel() error {
	newStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusCancelled)
	if err != nil {
		return err
	}

	if !t.status.CanTransitionTo(newStatus) {
		return errors.NewInvalidStatusTransitionError(t.status.Value(), newStatus.Value())
	}

	t.status = newStatus
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// Restart resets a failed or cancelled task back to pending
func (t *Task) Restart() error {
	newStatus, err := valueobjects.NewTaskStatus(valueobjects.StatusPending)
	if err != nil {
		return err
	}

	if !t.status.CanTransitionTo(newStatus) {
		return errors.NewInvalidStatusTransitionError(t.status.Value(), newStatus.Value())
	}

	t.status = newStatus
	t.startedAt = nil
	t.completedAt = nil
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// UpdateTitle updates the task title with validation
func (t *Task) UpdateTitle(title string) error {
	if err := validateTitle(title); err != nil {
		return err
	}

	t.title = strings.TrimSpace(title)
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// UpdateDescription updates the task description with validation
func (t *Task) UpdateDescription(description string) error {
	if err := validateDescription(description); err != nil {
		return err
	}

	t.description = strings.TrimSpace(description)
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// AddTokensUsed adds to the tokens used count
func (t *Task) AddTokensUsed(tokens int) error {
	if tokens < 0 {
		return errors.NewTaskValidationFailedError("tokens used cannot be negative", nil)
	}

	t.tokensUsed += tokens
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// SetMetadata sets a metadata key-value pair
func (t *Task) SetMetadata(key, value string) error {
	if key == "" {
		return errors.NewTaskValidationFailedError("metadata key cannot be empty", nil)
	}

	if t.metadata == nil {
		t.metadata = make(map[string]string)
	}

	t.metadata[key] = value
	t.updatedAt = time.Now()
	t.version++

	return nil
}

// RemoveMetadata removes a metadata key
func (t *Task) RemoveMetadata(key string) {
	if t.metadata != nil {
		delete(t.metadata, key)
		t.updatedAt = time.Now()
		t.version++
	}
}

// GetMetadata retrieves a metadata value
func (t *Task) GetMetadata(key string) (string, bool) {
	if t.metadata == nil {
		return "", false
	}
	value, exists := t.metadata[key]
	return value, exists
}

// Business rule queries

// IsActive checks if the task is in an active state
func (t *Task) IsActive() bool {
	return t.status.IsActive()
}

// IsCompleted checks if the task is completed
func (t *Task) IsCompleted() bool {
	return t.status.IsCompleted()
}

// IsFailed checks if the task is failed
func (t *Task) IsFailed() bool {
	return t.status.IsFailed()
}

// IsCancelled checks if the task is cancelled
func (t *Task) IsCancelled() bool {
	return t.status.IsCancelled()
}

// CanBeModified checks if the task can be modified
func (t *Task) CanBeModified() bool {
	return !t.status.IsTerminal()
}

// GetDuration returns the duration of the task (if completed)
func (t *Task) GetDuration() *time.Duration {
	if t.startedAt == nil {
		return nil
	}

	endTime := time.Now()
	if t.completedAt != nil {
		endTime = *t.completedAt
	}

	duration := endTime.Sub(*t.startedAt)
	return &duration
}

// IsLongRunning checks if the task has been running for more than the specified duration
func (t *Task) IsLongRunning(threshold time.Duration) bool {
	if t.startedAt == nil || t.status.IsCompleted() {
		return false
	}

	return time.Since(*t.startedAt) > threshold
}

// Validation functions

func validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.NewInvalidTaskTitleError("title cannot be empty")
	}
	if len(title) > 500 {
		return errors.NewInvalidTaskTitleError("title cannot exceed 500 characters")
	}
	return nil
}

func validateDescription(description string) error {
	description = strings.TrimSpace(description)
	if len(description) > 5000 {
		return errors.NewInvalidTaskDescriptionError("description cannot exceed 5000 characters")
	}
	return nil
}

func validateEpic(epic string) error {
	epic = strings.TrimSpace(epic)
	if epic == "" {
		return errors.NewTaskValidationFailedError("epic cannot be empty", nil)
	}
	if len(epic) > 255 {
		return errors.NewTaskValidationFailedError("epic cannot exceed 255 characters", nil)
	}
	return nil
}

// Reconstruction method for loading from persistence
func ReconstructTask(
	id valueobjects.TaskID,
	userID valueobjects.UserID,
	title string,
	description string,
	status valueobjects.TaskStatus,
	repository valueobjects.RepositoryPath,
	epic string,
	branch valueobjects.BranchName,
	tokensUsed int,
	createdAt time.Time,
	updatedAt time.Time,
	startedAt *time.Time,
	completedAt *time.Time,
	metadata map[string]string,
	version int64,
) *Task {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	return &Task{
		id:          id,
		userID:      userID,
		title:       title,
		description: description,
		status:      status,
		repository:  repository,
		epic:        epic,
		branch:      branch,
		tokensUsed:  tokensUsed,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		startedAt:   startedAt,
		completedAt: completedAt,
		metadata:    metadata,
		version:     version,
	}
}