package usecase

import (
	"context"
	"fmt"
	"log"
	_interface "main/features/claude/model/interface"
	"main/features/claude/model/request"
	"main/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunTasksClaudeUseCase struct {
	Repository        _interface.IRunTasksClaudeRepository
	ContextTimeout    time.Duration
	RepositoryManager *utils.RepositoryManager
}

func NewRunTasksClaudeUseCase(repo _interface.IRunTasksClaudeRepository, timeout time.Duration) _interface.IRunTasksClaudeUseCase {
	return &RunTasksClaudeUseCase{
		Repository:        repo,
		ContextTimeout:    timeout,
		RepositoryManager: utils.NewRepositoryManager(""),
	}
}

func (d *RunTasksClaudeUseCase) RunTasks(c context.Context, req *request.ReqRunTasksClaude) error {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	log.Printf("Starting Claude CLI task execution: %s", req.Tasks)
	log.Printf("Repository: %s, Working directory: %s, Continue: %v", req.RepositoryName, req.WorkingDir, req.ContinueTask)

	// repository_name 필수 검증
	if req.RepositoryName == "" {
		log.Printf("Repository name is required")
		return fmt.Errorf("repository_name is required")
	}

	// 레포지토리 관리 처리
	var repo *utils.RepositoryInfo
	var taskFilePath string

	// 고정된 경로로 레포지토리 경로 설정 (JokerTrickster 하위)
	repoPath := fmt.Sprintf("/Users/mac/project/git-repository/JokerTrickster/%s", req.RepositoryName)

	// 레포지토리 디렉토리 존재 확인
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Printf("Repository directory not found: %s", repoPath)
		return fmt.Errorf("repository '%s' not found at %s", req.RepositoryName, repoPath)
	}

	log.Printf("Using repository: %s at %s", req.RepositoryName, repoPath)

	// 작업 디렉토리 설정 (레포지토리 경로 우선)
	if req.WorkingDir == "" {
		req.WorkingDir = repoPath
	}

	// 간단한 RepositoryInfo 구조체 생성 (로깅용)
	repo = &utils.RepositoryInfo{
		Name:      req.RepositoryName,
		Path:      repoPath,
		IsGitRepo: true,
	}

	// Task file 생성은 건너뛰고 단순히 로깅만 수행
	log.Printf("Repository configured for task execution: %s", req.RepositoryName)

	// .claude/tasks 폴더에 작업 로그 생성
	taskLogPath, err := d.createTaskLog(repoPath, req)
	if err != nil {
		log.Printf("Failed to create task log: %v", err)
		// 로그 생성 실패해도 작업은 계속 진행
	}

	// 작업 기록 시작
	record := utils.TaskRecord{
		Timestamp:  time.Now(),
		Task:       req.Tasks,
		Status:     "started",
		WorkingDir: req.WorkingDir,
	}
	d.RepositoryManager.SaveTaskRecord(repo, record)

	// 작업 실행
	startTime := time.Now()
	var executionError error

	if req.Interactive {
		tasks := d.ParseTasks(req.Tasks)
		executionError = d.executeClaudeInteractive(ctx, tasks)
	} else {
		if req.WorkingDir != "" {
			executionError = d.executeClaudeWithWorkingDir(ctx, req.Tasks, req.WorkingDir)
		} else {
			executionError = d.executeClaudeAdvanced(ctx, req)
		}
	}

	// 작업 완료 기록
	if repo != nil {
		status := "completed"
		errorMsg := ""
		if executionError != nil {
			status = "failed"
			errorMsg = executionError.Error()
		}

		record := utils.TaskRecord{
			Timestamp:  startTime,
			Task:       req.Tasks,
			Status:     status,
			WorkingDir: req.WorkingDir,
			Error:      errorMsg,
		}
		d.RepositoryManager.SaveTaskRecord(repo, record)

		// 작업 파일 업데이트는 taskFilePath가 있을 때만 수행
		if taskFilePath != "" {
			d.updateTaskFile(taskFilePath, status, errorMsg)
		}

		// .claude/tasks 로그 업데이트
		if taskLogPath != "" {
			d.updateTaskLog(taskLogPath, status, errorMsg, executionError)
		}

		// 작업 완료 후 Git 커밋 및 푸시 (성공한 경우에만)
		if status == "completed" {
			d.autoCommitAndPush(repoPath, req.Tasks)
		}
	}

	if executionError != nil {
		log.Printf("Claude CLI execution failed: %v", executionError)
		return fmt.Errorf("failed to execute Claude CLI: %w", executionError)
	}

	log.Println("Claude CLI task execution completed successfully")
	return nil
}

// executeClaude executes Claude CLI with the given task
func (d *RunTasksClaudeUseCase) executeClaude(ctx context.Context, task string) error {
	// Claude CLI 명령어 구성
	// claude 명령어에 task를 전달 (승인 요청 없이 실행)
	args := []string{"--dangerously-skip-permissions", task}

	// 환경변수에서 Claude CLI 경로 확인 (기본값: "claude")
	claudeCmd := os.Getenv("CLAUDE_CLI_PATH")
	if claudeCmd == "" {
		claudeCmd = "claude"
	}

	// 명령어 실행
	cmd := exec.CommandContext(ctx, claudeCmd, args...)

	// 현재 디렉토리에서 실행
	cmd.Dir = "."

	// 환경변수 상속
	cmd.Env = os.Environ()

	// 출력 캡처
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 에러 로그에 출력 내용 포함
		return fmt.Errorf("claude command failed: %w\nOutput: %s", err, string(output))
	}

	// 성공시 출력 로그
	log.Printf("Claude CLI output:\n%s", string(output))

	return nil
}

// executeClaudeWithWorkingDir executes Claude CLI in a specific working directory
func (d *RunTasksClaudeUseCase) executeClaudeWithWorkingDir(ctx context.Context, task, workingDir string) error {
	// Claude CLI 명령어 구성 (승인 요청 없이 실행)
	args := []string{"--dangerously-skip-permissions", task}

	// 환경변수에서 Claude CLI 경로 확인 (기본값: "claude")
	claudeCmd := os.Getenv("CLAUDE_CLI_PATH")
	if claudeCmd == "" {
		claudeCmd = "claude"
	}

	// 명령어 실행
	cmd := exec.CommandContext(ctx, claudeCmd, args...)

	// 지정된 디렉토리에서 실행
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// 환경변수 상속
	cmd.Env = os.Environ()

	// 실시간 출력을 위해 stdout, stderr 설정
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 명령어 실행
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude command failed in directory %s: %w", workingDir, err)
	}

	log.Printf("Claude CLI executed successfully in directory: %s", workingDir)
	return nil
}

// executeClaudeInteractive executes Claude CLI in interactive mode with multiple tasks
func (d *RunTasksClaudeUseCase) executeClaudeInteractive(ctx context.Context, tasks []string) error {
	if len(tasks) == 0 {
		return fmt.Errorf("no tasks provided")
	}

	// 첫 번째 task로 Claude CLI 시작
	firstTask := tasks[0]
	remainingTasks := tasks[1:]

	log.Printf("Starting Claude CLI with task: %s", firstTask)

	// 첫 번째 task 실행
	if err := d.executeClaude(ctx, firstTask); err != nil {
		return fmt.Errorf("failed to execute first task: %w", err)
	}

	// 나머지 tasks 순차 실행
	for i, task := range remainingTasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("Executing additional task %d: %s", i+1, task)
			if err := d.executeClaude(ctx, task); err != nil {
				log.Printf("Failed to execute task %d: %v", i+1, err)
				// 하나의 task가 실패해도 계속 진행할지 결정
				continue
			}
		}
	}

	return nil
}

// ParseTasks parses task string into individual tasks
func (d *RunTasksClaudeUseCase) ParseTasks(taskString string) []string {
	// 줄바꿈이나 세미콜론으로 구분된 여러 task 처리
	tasks := strings.Split(taskString, "\n")
	var result []string

	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if task != "" {
			// 세미콜론으로도 구분 가능
			subTasks := strings.Split(task, ";")
			for _, subTask := range subTasks {
				subTask = strings.TrimSpace(subTask)
				if subTask != "" {
					result = append(result, subTask)
				}
			}
		}
	}

	return result
}

// executeClaudeAdvanced executes Claude CLI with advanced options from request
func (d *RunTasksClaudeUseCase) executeClaudeAdvanced(ctx context.Context, req *request.ReqRunTasksClaude) error {
	// Claude CLI 명령어 구성 (승인 요청 없이 실행)
	args := []string{"--dangerously-skip-permissions", req.Tasks}

	// 환경변수 또는 요청에서 Claude CLI 경로 확인
	claudeCmd := req.ClaudeCmd
	if claudeCmd == "" {
		claudeCmd = os.Getenv("CLAUDE_CLI_PATH")
		if claudeCmd == "" {
			claudeCmd = "claude"
		}
	}

	// 명령어 실행
	cmd := exec.CommandContext(ctx, claudeCmd, args...)

	// 작업 디렉토리 설정
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	} else {
		cmd.Dir = "."
	}

	// 환경변수 상속
	cmd.Env = os.Environ()

	// 출력 캡처
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 에러 로그에 출력 내용 포함
		return fmt.Errorf("claude command failed: %w\nOutput: %s", err, string(output))
	}

	// 성공시 출력 로그
	log.Printf("Claude CLI output:\n%s", string(output))

	return nil
}

// updateTaskFile 작업 파일에 결과 업데이트
func (d *RunTasksClaudeUseCase) updateTaskFile(taskFilePath, status, errorMsg string) error {
	content, err := os.ReadFile(taskFilePath)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var updateContent string

	if status == "completed" {
		updateContent = fmt.Sprintf(`

## Task Completed - %s

✅ **Status:** Completed successfully

### Result Notes

<!-- Task completed successfully -->

`, timestamp)
	} else if status == "failed" {
		updateContent = fmt.Sprintf(`

## Task Failed - %s

❌ **Status:** Failed
**Error:** %s

### Error Details

<!-- Review and fix the error -->

`, timestamp, errorMsg)
	}

	updatedContent := string(content) + updateContent
	return os.WriteFile(taskFilePath, []byte(updatedContent), 0644)
}

// createTaskLog .claude/tasks 폴더에 작업 로그 파일을 생성합니다
func (d *RunTasksClaudeUseCase) createTaskLog(repoPath string, req *request.ReqRunTasksClaude) (string, error) {
	claudeDir := filepath.Join(repoPath, ".claude")
	tasksDir := filepath.Join(claudeDir, "tasks")

	// .claude/tasks 디렉토리 생성
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// 타임스탬프를 포함한 파일명 생성
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("task_%s_%s.md", timestamp, req.RepositoryName)
	taskLogPath := filepath.Join(tasksDir, filename)

	// 작업 로그 초기 내용 생성
	content := fmt.Sprintf(`# Task Log - %s

## Task Information
- **Repository**: %s
- **Started**: %s
- **Working Directory**: %s
- **Interactive Mode**: %t
- **Status**: In Progress

## Task Description
%s

## Execution Log
- Task started at %s
- Waiting for execution to complete...

`, req.RepositoryName, req.RepositoryName, time.Now().Format("2006-01-02 15:04:05"),
		req.WorkingDir, req.Interactive, req.Tasks, time.Now().Format("2006-01-02 15:04:05"))

	// 파일 생성
	if err := os.WriteFile(taskLogPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to create task log file: %w", err)
	}

	log.Printf("Task log created: %s", taskLogPath)
	return taskLogPath, nil
}

// updateTaskLog 작업 완료 후 로그를 업데이트합니다
func (d *RunTasksClaudeUseCase) updateTaskLog(taskLogPath, status, errorMsg string, err error) {
	content, readErr := os.ReadFile(taskLogPath)
	if readErr != nil {
		log.Printf("Failed to read task log for update: %v", readErr)
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var updateContent string

	if status == "completed" {
		updateContent = fmt.Sprintf(`

## Task Completed - %s

✅ **Status**: Completed Successfully

### Result Summary
- Task execution completed without errors
- All operations finished successfully

### Completion Details
- Completed at: %s
- Final status: %s

---
*Task log generated by local-backend API*
`, timestamp, timestamp, status)
	} else if status == "failed" {
		updateContent = fmt.Sprintf(`

## Task Failed - %s

❌ **Status**: Failed

### Error Information
- **Error Message**: %s
- **Failed at**: %s

### Error Details
%s

### Troubleshooting
- Review the error message above
- Check repository permissions and Claude CLI installation
- Verify the task syntax and requirements

---
*Task log generated by local-backend API*
`, timestamp, errorMsg, timestamp, errorMsg)
	}

	// 기존 "Waiting for execution to complete..." 부분 제거하고 업데이트 내용 추가
	existingContent := string(content)
	existingContent = strings.Replace(existingContent, "- Waiting for execution to complete...",
		fmt.Sprintf("- Task completed at %s", timestamp), 1)

	updatedContent := existingContent + updateContent

	if writeErr := os.WriteFile(taskLogPath, []byte(updatedContent), 0644); writeErr != nil {
		log.Printf("Failed to update task log: %v", writeErr)
	} else {
		log.Printf("Task log updated: %s", taskLogPath)
	}
}

// autoCommitAndPush 작업 완료 후 자동으로 Git 커밋과 푸시를 수행합니다
func (d *RunTasksClaudeUseCase) autoCommitAndPush(repoPath, taskDescription string) {
	log.Printf("Starting auto Git commit and push for repository: %s", repoPath)

	// Git status 확인
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoPath
	statusOutput, err := statusCmd.Output()
	if err != nil {
		log.Printf("Failed to check git status: %v", err)
		return
	}

	// 변경사항이 없으면 커밋하지 않음
	if len(strings.TrimSpace(string(statusOutput))) == 0 {
		log.Printf("No changes to commit in repository: %s", repoPath)
		return
	}

	log.Printf("Found changes to commit: %s", string(statusOutput))

	// Git add .
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoPath
	if err := addCmd.Run(); err != nil {
		log.Printf("Failed to add files to git: %v", err)
		return
	}

	// 커밋 메시지 생성
	commitMsg := d.generateCommitMessage(taskDescription)

	// Git commit
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = repoPath
	commitOutput, err := commitCmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to commit changes: %v, output: %s", err, string(commitOutput))
		return
	}

	log.Printf("Successfully committed changes: %s", string(commitOutput))

	// Git push
	pushCmd := exec.Command("git", "push")
	pushCmd.Dir = repoPath
	pushOutput, err := pushCmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to push changes: %v, output: %s", err, string(pushOutput))
		return
	}

	log.Printf("Successfully pushed changes: %s", string(pushOutput))
	log.Printf("Auto Git commit and push completed for repository: %s", repoPath)
}

// generateCommitMessage 작업 내용을 기반으로 커밋 메시지를 생성합니다
func (d *RunTasksClaudeUseCase) generateCommitMessage(taskDescription string) string {
	// 작업 내용을 요약하여 커밋 메시지 생성
	summary := taskDescription
	if len(summary) > 100 {
		summary = summary[:97] + "..."
	}

	return fmt.Sprintf(`%s

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>`, summary)
}
