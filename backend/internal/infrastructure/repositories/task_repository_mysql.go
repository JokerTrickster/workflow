package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"ai-git-workbench/internal/domain/entities"
	"ai-git-workbench/internal/domain/repositories"
	"ai-git-workbench/internal/infrastructure/database"
)

// MySQLTaskRepository implements TaskRepository interface for MySQL
type MySQLTaskRepository struct {
	db *database.DB
}

// NewMySQLTaskRepository creates a new MySQL task repository
func NewMySQLTaskRepository(db *database.DB) repositories.TaskRepository {
	return &MySQLTaskRepository{
		db: db,
	}
}

// Create creates a new task in the database
func (r *MySQLTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert task
	query := `
		INSERT INTO tasks (id, branch_name, title, content, repository, user_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		task.ID,
		task.BranchName,
		task.Title,
		task.Content,
		task.Repository,
		task.UserID,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error inserting task: %w", err)
	}

	// Insert metadata if present
	if task.Metadata != nil && len(task.Metadata) > 0 {
		if err := r.insertMetadata(ctx, tx, task.ID, task.Metadata); err != nil {
			return fmt.Errorf("error inserting metadata: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a task by its ID
func (r *MySQLTaskRepository) GetByID(ctx context.Context, id string) (*entities.Task, error) {
	query := `
		SELECT id, branch_name, title, content, repository, user_id, status, created_at, updated_at
		FROM tasks
		WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	task := &entities.Task{}
	err := row.Scan(
		&task.ID,
		&task.BranchName,
		&task.Title,
		&task.Content,
		&task.Repository,
		&task.UserID,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("error scanning task: %w", err)
	}

	// Load metadata
	metadata, err := r.getMetadata(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error loading metadata: %w", err)
	}
	task.Metadata = metadata

	return task, nil
}

// GetByUserID retrieves all tasks for a specific user
func (r *MySQLTaskRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Task, error) {
	filter := repositories.TaskFilter{
		UserID: userID,
	}
	return r.List(ctx, filter)
}

// Update updates an existing task
func (r *MySQLTaskRepository) Update(ctx context.Context, task *entities.Task) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Update task
	query := `
		UPDATE tasks
		SET branch_name = ?, title = ?, content = ?, repository = ?, user_id = ?, status = ?, updated_at = ?
		WHERE id = ?`

	result, err := tx.ExecContext(ctx, query,
		task.BranchName,
		task.Title,
		task.Content,
		task.Repository,
		task.UserID,
		task.Status,
		time.Now(),
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating task: %w", err)
	}

	// Check if task was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Delete existing metadata
	if err := r.deleteMetadata(ctx, tx, task.ID); err != nil {
		return fmt.Errorf("error deleting old metadata: %w", err)
	}

	// Insert new metadata if present
	if task.Metadata != nil && len(task.Metadata) > 0 {
		if err := r.insertMetadata(ctx, tx, task.ID, task.Metadata); err != nil {
			return fmt.Errorf("error inserting new metadata: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// Delete deletes a task by its ID
func (r *MySQLTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return nil
}

// List retrieves tasks with optional filtering
func (r *MySQLTaskRepository) List(ctx context.Context, filter repositories.TaskFilter) ([]*entities.Task, error) {
	query := `
		SELECT id, branch_name, title, content, repository, user_id, status, created_at, updated_at
		FROM tasks`

	var conditions []string
	var args []interface{}

	// Build WHERE clause
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Repository != "" {
		conditions = append(conditions, "repository = ?")
		args = append(args, filter.Repository)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*entities.Task
	for rows.Next() {
		task := &entities.Task{}
		err := rows.Scan(
			&task.ID,
			&task.BranchName,
			&task.Title,
			&task.Content,
			&task.Repository,
			&task.UserID,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning task: %w", err)
		}

		// Load metadata for each task
		metadata, err := r.getMetadata(ctx, task.ID)
		if err != nil {
			return nil, fmt.Errorf("error loading metadata for task %s: %w", task.ID, err)
		}
		task.Metadata = metadata

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tasks, nil
}

// UpdateStatus updates only the status of a task
func (r *MySQLTaskRepository) UpdateStatus(ctx context.Context, id string, status entities.TaskStatus) error {
	query := `
		UPDATE tasks
		SET status = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating task status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return nil
}

// GetByStatus retrieves all tasks with a specific status
func (r *MySQLTaskRepository) GetByStatus(ctx context.Context, status entities.TaskStatus) ([]*entities.Task, error) {
	filter := repositories.TaskFilter{
		Status: status,
	}
	return r.List(ctx, filter)
}

// Count returns the total number of tasks matching the filter
func (r *MySQLTaskRepository) Count(ctx context.Context, filter repositories.TaskFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM tasks`

	var conditions []string
	var args []interface{}

	// Build WHERE clause (same logic as List)
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Repository != "" {
		conditions = append(conditions, "repository = ?")
		args = append(args, filter.Repository)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting tasks: %w", err)
	}

	return count, nil
}

// Helper functions for metadata management

// insertMetadata inserts metadata for a task
func (r *MySQLTaskRepository) insertMetadata(ctx context.Context, tx *sql.Tx, taskID string, metadata map[string]string) error {
	query := `INSERT INTO task_metadata (task_id, meta_key, meta_value) VALUES (?, ?, ?)`

	for key, value := range metadata {
		_, err := tx.ExecContext(ctx, query, taskID, key, value)
		if err != nil {
			return fmt.Errorf("error inserting metadata %s: %w", key, err)
		}
	}

	return nil
}

// getMetadata retrieves metadata for a task
func (r *MySQLTaskRepository) getMetadata(ctx context.Context, taskID string) (map[string]string, error) {
	query := `SELECT meta_key, meta_value FROM task_metadata WHERE task_id = ?`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("error querying metadata: %w", err)
	}
	defer rows.Close()

	metadata := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("error scanning metadata: %w", err)
		}
		metadata[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata rows: %w", err)
	}

	return metadata, nil
}

// deleteMetadata deletes all metadata for a task
func (r *MySQLTaskRepository) deleteMetadata(ctx context.Context, tx *sql.Tx, taskID string) error {
	query := `DELETE FROM task_metadata WHERE task_id = ?`

	_, err := tx.ExecContext(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("error deleting metadata: %w", err)
	}

	return nil
}