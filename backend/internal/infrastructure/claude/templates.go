package claude

import (
	"fmt"
	"strings"
	"text/template"
)

// TemplateManager manages prompt templates for different workflow types
type TemplateManager struct {
	templates map[string]*PromptTemplate
}

// NewTemplateManager creates a new template manager with predefined templates
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*PromptTemplate),
	}

	tm.loadDefaultTemplates()
	return tm
}

// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*PromptTemplate, error) {
	template, exists := tm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return template, nil
}

// RenderTemplate renders a template with the given variables
func (tm *TemplateManager) RenderTemplate(name string, variables map[string]interface{}) (string, string, error) {
	promptTemplate, err := tm.GetTemplate(name)
	if err != nil {
		return "", "", err
	}

	// Render system prompt
	systemPrompt, err := tm.renderString(promptTemplate.System, variables)
	if err != nil {
		return "", "", fmt.Errorf("failed to render system prompt: %w", err)
	}

	// Render user prompt
	userPrompt, err := tm.renderString(promptTemplate.Template, variables)
	if err != nil {
		return "", "", fmt.Errorf("failed to render user prompt: %w", err)
	}

	return systemPrompt, userPrompt, nil
}

// renderString renders a template string with variables
func (tm *TemplateManager) renderString(templateStr string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// AddTemplate adds a new template
func (tm *TemplateManager) AddTemplate(template *PromptTemplate) {
	tm.templates[template.Name] = template
}

// ListTemplates returns all available template names
func (tm *TemplateManager) ListTemplates() []string {
	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// loadDefaultTemplates loads the predefined templates
func (tm *TemplateManager) loadDefaultTemplates() {
	// Code Review Template
	tm.AddTemplate(&PromptTemplate{
		Name:        "code_review",
		Description: "Comprehensive code review analysis",
		System: `You are an expert code reviewer with extensive experience in software engineering best practices. 
Analyze code thoroughly and provide constructive, actionable feedback. Focus on:
- Code quality and adherence to best practices
- Potential bugs and security vulnerabilities
- Performance considerations and optimizations
- Maintainability and readability improvements
- Architecture and design patterns

Provide your analysis in JSON format with structured feedback.`,
		Template: `Please review the following code:

**Context:** {{.context}}

**Code:**
` + "```{{.language}}\n{{.code}}\n```" + `

Provide a comprehensive code review including:
1. Overall assessment and score (1-10)
2. Specific issues found with severity levels
3. Positive aspects worth highlighting
4. Recommendations for improvement
5. Security considerations if applicable

Format your response as JSON with the structure:
{
  "overall_score": 1-10,
  "summary": "Brief overall assessment",
  "issues": [...],
  "positive_aspects": [...],
  "recommendations": [...]
}`,
		Variables: map[string]string{
			"context":  "Additional context about the code",
			"code":     "The code to review",
			"language": "Programming language",
		},
	})

	// Issue Analysis Template
	tm.AddTemplate(&PromptTemplate{
		Name:        "issue_analysis",
		Description: "GitHub issue analysis and solution planning",
		System: `You are an experienced software engineer and project manager. Analyze GitHub issues thoroughly to understand:
- The core problem and its root causes
- Technical complexity and implementation challenges
- Resource requirements and time estimation
- Risk assessment and mitigation strategies
- Dependencies and prerequisites

Provide structured analysis to help with project planning and implementation.`,
		Template: `Analyze the following GitHub issue:

**Repository Context:** {{.repo_context}}

**Issue Title:** {{.title}}

**Issue Description:**
{{.description}}

**Labels:** {{.labels}}

Please provide a comprehensive analysis including:
1. Problem categorization and complexity assessment
2. Root cause analysis (if it's a bug)
3. Suggested implementation approach
4. Time and effort estimation
5. Risk assessment and dependencies
6. Testing strategy recommendations

Format your response as JSON:
{
  "complexity": "low|medium|high",
  "category": "bug|feature|enhancement|documentation",
  "estimated_hours": number,
  "summary": "Brief analysis summary",
  "implementation_approach": "Step-by-step approach",
  "risks": [...],
  "dependencies": [...]
}`,
		Variables: map[string]string{
			"repo_context": "Repository and project context",
			"title":        "Issue title",
			"description":  "Issue description",
			"labels":       "Issue labels",
		},
	})

	// Bug Fix Template
	tm.AddTemplate(&PromptTemplate{
		Name:        "bug_fix",
		Description: "Bug analysis and fix suggestions",
		System: `You are a debugging expert with deep knowledge of software engineering principles. 
When analyzing bugs:
- Identify the root cause systematically
- Consider edge cases and error conditions
- Suggest robust, well-tested solutions
- Think about prevention strategies
- Consider performance and security implications

Provide clear, actionable solutions with detailed explanations.`,
		Template: `Help fix the following bug:

**Bug Description:** {{.bug_description}}

**Environment:** {{.environment}}

**Steps to Reproduce:**
{{.steps_to_reproduce}}

**Expected Behavior:** {{.expected_behavior}}

**Actual Behavior:** {{.actual_behavior}}

**Relevant Code:**
` + "```{{.language}}\n{{.code}}\n```" + `

**Error Messages/Logs:**
{{.error_logs}}

Please provide:
1. Root cause analysis
2. Specific fix recommendations with code examples
3. Testing strategy to verify the fix
4. Prevention measures for similar bugs
5. Any performance or security considerations

Format as JSON:
{
  "root_cause": "Detailed explanation of the root cause",
  "fix_approach": "High-level approach to fix",
  "code_changes": [...],
  "testing_strategy": [...],
  "prevention_measures": [...]
}`,
		Variables: map[string]string{
			"bug_description":     "Description of the bug",
			"environment":         "Environment details",
			"steps_to_reproduce":  "Steps to reproduce the bug",
			"expected_behavior":   "What should happen",
			"actual_behavior":     "What actually happens",
			"code":                "Relevant code snippet",
			"language":            "Programming language",
			"error_logs":          "Error messages or logs",
		},
	})

	// Feature Implementation Template
	tm.AddTemplate(&PromptTemplate{
		Name:        "feature_implementation",
		Description: "Feature design and implementation planning",
		System: `You are a senior software architect and developer. When designing features:
- Consider scalability and maintainability
- Follow established architectural patterns
- Think about user experience and edge cases
- Plan for testing and quality assurance
- Consider security and performance implications
- Provide detailed implementation guidance

Create comprehensive implementation plans with clear steps and best practices.`,
		Template: `Design and plan implementation for the following feature:

**Feature Description:** {{.feature_description}}

**Requirements:** {{.requirements}}

**Existing System Context:** {{.system_context}}

**Constraints:** {{.constraints}}

**Target Users:** {{.target_users}}

Please provide:
1. Feature analysis and design approach
2. Technical architecture and component design
3. Implementation steps and timeline
4. API design (if applicable)
5. Database schema changes (if needed)
6. Testing strategy
7. Deployment considerations

Format as JSON:
{
  "design_approach": "High-level design strategy",
  "architecture": {
    "components": [...],
    "interactions": [...]
  },
  "implementation_steps": [...],
  "api_design": {...},
  "database_changes": [...],
  "testing_strategy": [...],
  "deployment_plan": [...]
}`,
		Variables: map[string]string{
			"feature_description": "Description of the feature to implement",
			"requirements":        "Detailed requirements",
			"system_context":      "Existing system architecture and context",
			"constraints":         "Technical or business constraints",
			"target_users":        "Target user personas and use cases",
		},
	})

	// General Workflow Template
	tm.AddTemplate(&PromptTemplate{
		Name:        "general_workflow",
		Description: "General purpose workflow processing",
		System: `You are an AI assistant helping with software development workflows. 
Provide helpful, accurate, and actionable responses based on the context provided.
Always consider best practices, security, and maintainability in your recommendations.`,
		Template: `{{.request}}

{{if .context}}
**Context:** {{.context}}
{{end}}

{{if .additional_info}}
**Additional Information:** {{.additional_info}}
{{end}}

Please provide a comprehensive response addressing the request with practical recommendations and implementation guidance.`,
		Variables: map[string]string{
			"request":         "The main request or question",
			"context":         "Additional context information",
			"additional_info": "Any additional relevant information",
		},
	})
}