package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"main/features/epic-tasks/model/request"

	"gopkg.in/yaml.v3"
)

type EpicTaskRepository struct{}

func NewEpicTaskRepository() *EpicTaskRepository {
	return &EpicTaskRepository{}
}

// ValidateRepository checks if repository name is safe
func (r *EpicTaskRepository) ValidateRepository(repository string) error {
	if repository == "" {
		return fmt.Errorf("repository parameter is required")
	}

	// Prevent path traversal attacks
	if strings.Contains(repository, "..") || strings.Contains(repository, "/") || strings.Contains(repository, "\\") {
		return fmt.Errorf("invalid repository name: contains illegal characters")
	}

	return nil
}

// GetTasksDir returns the tasks directory path for a repository
func (r *EpicTaskRepository) GetTasksDir(repository string) string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".claude", "epics", "repositories", repository, "tasks")
}

// EnsureTasksDir creates the tasks directory if it doesn't exist
func (r *EpicTaskRepository) EnsureTasksDir(repository string) error {
	tasksDir := r.GetTasksDir(repository)
	return os.MkdirAll(tasksDir, 0755)
}

// ParseTaskFile parses a markdown file with YAML frontmatter
func (r *EpicTaskRepository) ParseTaskFile(filePath string) (*request.EpicTaskFile, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Check if file has frontmatter
	if !strings.HasPrefix(contentStr, "---\n") {
		return nil, fmt.Errorf("file does not have YAML frontmatter")
	}

	// Find the end of frontmatter
	parts := strings.SplitN(contentStr[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid frontmatter format")
	}

	frontmatterStr := parts[0]
	markdownContent := strings.TrimSpace(parts[1])

	// Parse YAML frontmatter
	var metadata request.TaskFileMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &request.EpicTaskFile{
		Metadata: metadata,
		Content:  markdownContent,
	}, nil
}

// FormatTaskFile formats a task file with YAML frontmatter
func (r *EpicTaskRepository) FormatTaskFile(task *request.EpicTaskFile) (string, error) {
	// Marshal metadata to YAML
	metadataBytes, err := yaml.Marshal(&task.Metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Combine frontmatter and content
	return fmt.Sprintf("---\n%s---\n\n%s", string(metadataBytes), task.Content), nil
}

// GetAllTasks retrieves all epic tasks for a repository
func (r *EpicTaskRepository) GetAllTasks(repository string) ([]request.EpicTaskFile, error) {
	tasksDir := r.GetTasksDir(repository)

	// Ensure directory exists
	if err := r.EnsureTasksDir(repository); err != nil {
		return nil, err
	}

	// Check if directory exists
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return []request.EpicTaskFile{}, nil
	}

	files, err := ioutil.ReadDir(tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	var tasks []request.EpicTaskFile

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") || file.Name() == "task-template.md" {
			continue
		}

		filePath := filepath.Join(tasksDir, file.Name())
		task, err := r.ParseTaskFile(filePath)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Failed to parse task file %s: %v\n", file.Name(), err)
			continue
		}

		// Ensure repository field is set
		if task.Metadata.Repository == "" {
			task.Metadata.Repository = repository
		}

		tasks = append(tasks, *task)
	}

	return tasks, nil
}

// GetTask retrieves a specific epic task
func (r *EpicTaskRepository) GetTask(repository, taskID string) (*request.EpicTaskFile, error) {
	tasksDir := r.GetTasksDir(repository)
	filePath := filepath.Join(tasksDir, fmt.Sprintf("%s.md", taskID))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("task not found")
	}

	task, err := r.ParseTaskFile(filePath)
	if err != nil {
		return nil, err
	}

	// Ensure repository field is set
	if task.Metadata.Repository == "" {
		task.Metadata.Repository = repository
	}

	return task, nil
}

// CreateTask creates a new epic task
func (r *EpicTaskRepository) CreateTask(repository string, task *request.EpicTaskFile) error {
	tasksDir := r.GetTasksDir(repository)

	// Ensure directory exists
	if err := r.EnsureTasksDir(repository); err != nil {
		return err
	}

	// Ensure repository field is set
	if task.Metadata.Repository == "" {
		task.Metadata.Repository = repository
	}

	filename := fmt.Sprintf("%s.md", task.Metadata.ID)
	filePath := filepath.Join(tasksDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("task file already exists")
	}

	// Format task file
	content, err := r.FormatTaskFile(task)
	if err != nil {
		return err
	}

	// Write file
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// UpdateTask updates an existing epic task
func (r *EpicTaskRepository) UpdateTask(repository string, task *request.EpicTaskFile) error {
	tasksDir := r.GetTasksDir(repository)
	filename := fmt.Sprintf("%s.md", task.Metadata.ID)
	filePath := filepath.Join(tasksDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("task not found")
	}

	// Ensure repository field is set
	if task.Metadata.Repository == "" {
		task.Metadata.Repository = repository
	}

	// Format task file
	content, err := r.FormatTaskFile(task)
	if err != nil {
		return err
	}

	// Write file
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// DeleteTask deletes an epic task
func (r *EpicTaskRepository) DeleteTask(repository, taskID string) error {
	tasksDir := r.GetTasksDir(repository)
	filename := fmt.Sprintf("%s.md", taskID)
	filePath := filepath.Join(tasksDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("task not found")
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	return nil
}
