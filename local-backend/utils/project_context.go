package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ClaudeDir           = ".claude"
	ProjectContextFile  = "PROJECT_CONTEXT.md"
	ProjectRulesFile    = "RULES.md"
)

// ProjectContextManager manages project-specific documentation
type ProjectContextManager struct {
	workingDir string
}

// NewProjectContextManager creates a new project context manager
func NewProjectContextManager(workingDir string) *ProjectContextManager {
	return &ProjectContextManager{
		workingDir: workingDir,
	}
}

// EnsureClaudeDirectory ensures .claude directory exists
func (p *ProjectContextManager) EnsureClaudeDirectory() error {
	claudeDir := filepath.Join(p.workingDir, ClaudeDir)

	// Check if directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		// Create directory with proper permissions
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			return fmt.Errorf("failed to create .claude directory: %w", err)
		}
	}

	return nil
}

// GetProjectContextPath returns the path to PROJECT_CONTEXT.md
func (p *ProjectContextManager) GetProjectContextPath() string {
	return filepath.Join(p.workingDir, ClaudeDir, ProjectContextFile)
}

// GetProjectRulesPath returns the path to RULES.md
func (p *ProjectContextManager) GetProjectRulesPath() string {
	return filepath.Join(p.workingDir, ClaudeDir, ProjectRulesFile)
}

// ProjectContextExists checks if PROJECT_CONTEXT.md exists
func (p *ProjectContextManager) ProjectContextExists() bool {
	path := p.GetProjectContextPath()
	_, err := os.Stat(path)
	return err == nil
}

// ProjectRulesExists checks if RULES.md exists
func (p *ProjectContextManager) ProjectRulesExists() bool {
	path := p.GetProjectRulesPath()
	_, err := os.Stat(path)
	return err == nil
}

// ReadProjectContext reads PROJECT_CONTEXT.md content
func (p *ProjectContextManager) ReadProjectContext() (string, error) {
	path := p.GetProjectContextPath()

	if !p.ProjectContextExists() {
		return "", fmt.Errorf("PROJECT_CONTEXT.md does not exist")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read PROJECT_CONTEXT.md: %w", err)
	}

	return string(content), nil
}

// ReadProjectRules reads RULES.md content
func (p *ProjectContextManager) ReadProjectRules() (string, error) {
	path := p.GetProjectRulesPath()

	if !p.ProjectRulesExists() {
		return "", fmt.Errorf("RULES.md does not exist")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read RULES.md: %w", err)
	}

	return string(content), nil
}

// CreateDefaultProjectContext creates a default PROJECT_CONTEXT.md if it doesn't exist
func (p *ProjectContextManager) CreateDefaultProjectContext(repositoryName string) error {
	if p.ProjectContextExists() {
		return nil // Already exists
	}

	// Ensure .claude directory exists
	if err := p.EnsureClaudeDirectory(); err != nil {
		return err
	}

	// Generate default content
	defaultContent := p.generateDefaultProjectContext(repositoryName)

	// Write to file
	path := p.GetProjectContextPath()
	if err := os.WriteFile(path, []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("failed to create PROJECT_CONTEXT.md: %w", err)
	}

	return nil
}

// generateDefaultProjectContext generates default PROJECT_CONTEXT.md content
func (p *ProjectContextManager) generateDefaultProjectContext(repositoryName string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s - Project Context\n\n", repositoryName))
	builder.WriteString("This document provides context about the project architecture, codebase structure, and important conventions.\n\n")

	builder.WriteString("## Project Overview\n\n")
	builder.WriteString("<!-- Describe the project's purpose, main features, and technology stack -->\n\n")

	builder.WriteString("## Architecture\n\n")
	builder.WriteString("<!-- Document the overall architecture, key components, and their relationships -->\n\n")

	builder.WriteString("## Directory Structure\n\n")
	builder.WriteString("```\n")
	builder.WriteString("<!-- Add your project's directory structure here -->\n")
	builder.WriteString("```\n\n")

	builder.WriteString("## Key Technologies\n\n")
	builder.WriteString("<!-- List main technologies, frameworks, and libraries -->\n\n")

	builder.WriteString("## Coding Conventions\n\n")
	builder.WriteString("<!-- Document naming conventions, code style, and best practices -->\n\n")

	builder.WriteString("## Important Files\n\n")
	builder.WriteString("<!-- List critical files and their purposes -->\n\n")

	builder.WriteString("## Dependencies\n\n")
	builder.WriteString("<!-- Document major dependencies and their roles -->\n\n")

	builder.WriteString("## Development Workflow\n\n")
	builder.WriteString("<!-- Describe the development process, testing approach, and deployment -->\n\n")

	builder.WriteString("## Notes\n\n")
	builder.WriteString("<!-- Add any additional important information -->\n\n")

	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("*This document was auto-generated. Please update it with accurate project information.*\n"))

	return builder.String()
}

// CreateDefaultProjectRules creates a default RULES.md if it doesn't exist
func (p *ProjectContextManager) CreateDefaultProjectRules(repositoryName string) error {
	if p.ProjectRulesExists() {
		return nil // Already exists
	}

	// Ensure .claude directory exists
	if err := p.EnsureClaudeDirectory(); err != nil {
		return err
	}

	// Generate default content
	defaultContent := p.generateDefaultProjectRules(repositoryName)

	// Write to file
	path := p.GetProjectRulesPath()
	if err := os.WriteFile(path, []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("failed to create RULES.md: %w", err)
	}

	return nil
}

// generateDefaultProjectRules generates default RULES.md content
func (p *ProjectContextManager) generateDefaultProjectRules(repositoryName string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s - Project Rules\n\n", repositoryName))
	builder.WriteString("Project-specific rules and conventions for AI-assisted development.\n\n")

	builder.WriteString("## Code Style\n\n")
	builder.WriteString("<!-- Define code style requirements specific to this project -->\n\n")

	builder.WriteString("## File Organization\n\n")
	builder.WriteString("<!-- Specify where different types of files should be created -->\n\n")

	builder.WriteString("## Naming Conventions\n\n")
	builder.WriteString("<!-- Document naming patterns for files, functions, variables, etc. -->\n\n")

	builder.WriteString("## Testing Requirements\n\n")
	builder.WriteString("<!-- Specify testing requirements and conventions -->\n\n")

	builder.WriteString("## Documentation Standards\n\n")
	builder.WriteString("<!-- Define documentation requirements -->\n\n")

	builder.WriteString("## Prohibited Actions\n\n")
	builder.WriteString("<!-- List things that should NOT be done in this project -->\n\n")

	builder.WriteString("## Required Actions\n\n")
	builder.WriteString("<!-- List things that MUST be done when making changes -->\n\n")

	builder.WriteString("---\n")
	builder.WriteString("*This document was auto-generated. Please customize it for your project.*\n")

	return builder.String()
}

// GetContextForPrompt reads and combines all available context
func (p *ProjectContextManager) GetContextForPrompt() string {
	var builder strings.Builder

	// Try to read project context
	if context, err := p.ReadProjectContext(); err == nil && context != "" {
		builder.WriteString("PROJECT CONTEXT:\n")
		builder.WriteString(context)
		builder.WriteString("\n\n")
	}

	// Try to read project rules
	if rules, err := p.ReadProjectRules(); err == nil && rules != "" {
		builder.WriteString("PROJECT RULES:\n")
		builder.WriteString(rules)
		builder.WriteString("\n\n")
	}

	return builder.String()
}
