package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"main/features/work-logs/model/request"
	"main/features/work-logs/model/response"
)

type WorkLogRepository struct{}

func NewWorkLogRepository() *WorkLogRepository {
	return &WorkLogRepository{}
}

// ValidateRepository checks if repository name is safe
func (r *WorkLogRepository) ValidateRepository(repository string) error {
	if repository == "" {
		return fmt.Errorf("repository parameter is required")
	}

	// Prevent path traversal attacks
	if strings.Contains(repository, "..") || strings.Contains(repository, "/") || strings.Contains(repository, "\\") {
		return fmt.Errorf("invalid repository name: contains illegal characters")
	}

	return nil
}

// ValidateDate checks if date format is correct
func (r *WorkLogRepository) ValidateDate(date string) error {
	if date == "" {
		return nil // Optional parameter
	}

	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}

	return nil
}

// GetLogsDir returns the logs directory path for a repository
func (r *WorkLogRepository) GetLogsDir(repository string) string {
	// Use current working directory + .claude/logs/repository
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".claude", "logs", repository)
}

// GetLogFilePath returns the full path to a log file
func (r *WorkLogRepository) GetLogFilePath(repository, date string) string {
	logsDir := r.GetLogsDir(repository)
	return filepath.Join(logsDir, fmt.Sprintf("%s.md", date))
}

// EnsureDirectoryExists creates the directory if it doesn't exist
func (r *WorkLogRepository) EnsureDirectoryExists(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// GetWorkLogs retrieves work logs for a repository within date range
func (r *WorkLogRepository) GetWorkLogs(repository, startDate, endDate string) ([]response.DailyWorkLog, error) {
	logsDir := r.GetLogsDir(repository)

	// Check if directory exists
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		// Directory doesn't exist, return empty array
		return []response.DailyWorkLog{}, nil
	}

	files, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs directory: %w", err)
	}

	var workLogs []response.DailyWorkLog

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		date := strings.TrimSuffix(file.Name(), ".md")

		// Filter by date range if provided
		if startDate != "" && date < startDate {
			continue
		}
		if endDate != "" && date > endDate {
			continue
		}

		workLogs = append(workLogs, response.DailyWorkLog{
			Date:       date,
			Repository: repository,
			Entries:    []request.WorkLogEntry{}, // Simplified - not parsing markdown back
		})
	}

	return workLogs, nil
}

// FormatLogEntry formats a work log entry as markdown
func (r *WorkLogRepository) FormatLogEntry(entry request.WorkLogEntry) string {
	time := entry.Timestamp.Format("15:04:05")
	markdown := fmt.Sprintf("### %s - %s (%s)\n\n", time, entry.TaskTitle, entry.Status)

	if entry.ProgressUpdate != "" {
		markdown += fmt.Sprintf("**Progress**: %s\n\n", entry.ProgressUpdate)
	}

	if len(entry.IssuesDiscovered) > 0 {
		markdown += "**Issues Discovered**:\n"
		for _, issue := range entry.IssuesDiscovered {
			markdown += fmt.Sprintf("- %s\n", issue)
		}
		markdown += "\n"
	}

	if len(entry.ImprovementsMade) > 0 {
		markdown += "**Improvements Made**:\n"
		for _, improvement := range entry.ImprovementsMade {
			markdown += fmt.Sprintf("- %s\n", improvement)
		}
		markdown += "\n"
	}

	if entry.Metadata != nil {
		markdown += "**Metadata**:\n"
		if entry.Metadata.Branch != "" {
			markdown += fmt.Sprintf("- Branch: %s\n", entry.Metadata.Branch)
		}
		if entry.Metadata.GithubIssue != 0 {
			markdown += fmt.Sprintf("- GitHub Issue: #%d\n", entry.Metadata.GithubIssue)
		}
		if entry.Metadata.PrUrl != "" {
			markdown += fmt.Sprintf("- PR URL: %s\n", entry.Metadata.PrUrl)
		}
		if entry.Metadata.TokensUsed != 0 {
			markdown += fmt.Sprintf("- Tokens Used: %d\n", entry.Metadata.TokensUsed)
		}
		markdown += "\n"
	}

	markdown += "---\n\n"
	return markdown
}

// GetDailyLogHeader returns the header for a daily log file
func (r *WorkLogRepository) GetDailyLogHeader(repository, date string) string {
	return fmt.Sprintf(`# Work Log - %s - %s

Generated automatically by Claude Code workflow system.

---

`, repository, date)
}

// CreateWorkLogEntry creates or appends a work log entry
func (r *WorkLogRepository) CreateWorkLogEntry(repository string, entry request.WorkLogEntry) error {
	today := time.Now().Format("2006-01-02")
	logsDir := r.GetLogsDir(repository)
	logFilePath := r.GetLogFilePath(repository, today)

	// Ensure directory exists
	if err := r.EnsureDirectoryExists(logsDir); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Check if log file exists for today
	var existingContent string
	if data, err := ioutil.ReadFile(logFilePath); err == nil {
		existingContent = string(data)
	} else {
		// File doesn't exist, create with header
		existingContent = r.GetDailyLogHeader(repository, today)
	}

	// Append new entry
	entryContent := r.FormatLogEntry(entry)
	updatedContent := existingContent + entryContent

	// Write updated content
	if err := ioutil.WriteFile(logFilePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}
