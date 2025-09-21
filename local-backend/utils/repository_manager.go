package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RepositoryManager struct {
	BaseDir string // 모든 레포지토리가 저장된 기본 디렉토리
}

type TaskRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	Task        string    `json:"task"`
	Status      string    `json:"status"` // started, completed, failed
	WorkingDir  string    `json:"working_dir"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type RepositoryInfo struct {
	Name          string       `json:"name"`
	Path          string       `json:"path"`
	IsGitRepo     bool         `json:"is_git_repo"`
	LastTaskTime  time.Time    `json:"last_task_time"`
	TaskHistory   []TaskRecord `json:"task_history"`
	ClaudeTaskDir string       `json:"claude_task_dir"`
}

func NewRepositoryManager(baseDir string) *RepositoryManager {
	if baseDir == "" {
		// 기본 디렉토리 설정 (환경변수 또는 홈 디렉토리)
		if envDir := os.Getenv("REPOSITORIES_BASE_DIR"); envDir != "" {
			baseDir = envDir
		} else {
			homeDir, _ := os.UserHomeDir()
			baseDir = filepath.Join(homeDir, "repositories")
		}
	}
	return &RepositoryManager{BaseDir: baseDir}
}

// FindRepository 레포지토리 이름으로 경로 찾기
func (rm *RepositoryManager) FindRepository(name string) (*RepositoryInfo, error) {
	// 1. 직접 경로로 찾기
	directPath := filepath.Join(rm.BaseDir, name)
	if info, err := rm.analyzeRepository(directPath, name); err == nil {
		return info, nil
	}

	// 2. 하위 디렉토리에서 찾기
	var foundRepo *RepositoryInfo
	filepath.Walk(rm.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(name)) {
			if repoInfo, err := rm.analyzeRepository(path, info.Name()); err == nil {
				foundRepo = repoInfo
				return filepath.SkipDir
			}
		}
		return nil
	})

	if foundRepo != nil {
		return foundRepo, nil
	}

	return nil, fmt.Errorf("repository '%s' not found in %s", name, rm.BaseDir)
}

// analyzeRepository 레포지토리 정보 분석
func (rm *RepositoryManager) analyzeRepository(path, name string) (*RepositoryInfo, error) {
	// 디렉토리 존재 확인
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	// Git 레포지토리 확인
	gitDir := filepath.Join(path, ".git")
	isGitRepo := false
	if _, err := os.Stat(gitDir); err == nil {
		isGitRepo = true
	}

	// .claude/tasks 디렉토리 경로
	claudeTaskDir := filepath.Join(path, ".claude", "tasks")

	repo := &RepositoryInfo{
		Name:          name,
		Path:          path,
		IsGitRepo:     isGitRepo,
		ClaudeTaskDir: claudeTaskDir,
		TaskHistory:   []TaskRecord{},
	}

	// 작업 히스토리 로드
	if err := rm.loadTaskHistory(repo); err != nil {
		// 히스토리 로드 실패는 무시 (처음 사용하는 레포지토리일 수 있음)
	}

	return repo, nil
}

// EnsureClaudeTaskDir .claude/tasks 디렉토리 생성
func (rm *RepositoryManager) EnsureClaudeTaskDir(repo *RepositoryInfo) error {
	return os.MkdirAll(repo.ClaudeTaskDir, 0755)
}

// GetTaskFilePath 작업 파일 경로 생성
func (rm *RepositoryManager) GetTaskFilePath(repo *RepositoryInfo, timestamp time.Time) string {
	filename := fmt.Sprintf("task_%s.md", timestamp.Format("20060102_150405"))
	return filepath.Join(repo.ClaudeTaskDir, filename)
}

// GetLatestTaskFile 가장 최근 작업 파일 찾기
func (rm *RepositoryManager) GetLatestTaskFile(repo *RepositoryInfo) (string, error) {
	files, err := filepath.Glob(filepath.Join(repo.ClaudeTaskDir, "task_*.md"))
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no task files found")
	}

	// 가장 최근 파일 찾기 (파일명 기준 정렬)
	var latestFile string
	for _, file := range files {
		if latestFile == "" || file > latestFile {
			latestFile = file
		}
	}

	return latestFile, nil
}

// CreateTaskFile 새 작업 파일 생성
func (rm *RepositoryManager) CreateTaskFile(repo *RepositoryInfo, task string, continueTask bool) (string, error) {
	if err := rm.EnsureClaudeTaskDir(repo); err != nil {
		return "", err
	}

	timestamp := time.Now()
	var taskFilePath string
	var content string

	if continueTask {
		// 기존 작업 이어하기
		latestFile, err := rm.GetLatestTaskFile(repo)
		if err != nil {
			// 기존 파일이 없으면 새로 생성
			taskFilePath = rm.GetTaskFilePath(repo, timestamp)
			content = rm.generateNewTaskContent(repo, task, timestamp)
		} else {
			taskFilePath = latestFile
			// 기존 파일에 추가
			existingContent, err := os.ReadFile(latestFile)
			if err != nil {
				return "", err
			}
			content = string(existingContent) + "\n\n" + rm.generateContinueTaskContent(task, timestamp)
		}
	} else {
		// 새 작업 파일 생성
		taskFilePath = rm.GetTaskFilePath(repo, timestamp)
		content = rm.generateNewTaskContent(repo, task, timestamp)
	}

	if err := os.WriteFile(taskFilePath, []byte(content), 0644); err != nil {
		return "", err
	}

	return taskFilePath, nil
}

// generateNewTaskContent 새 작업 파일 내용 생성
func (rm *RepositoryManager) generateNewTaskContent(repo *RepositoryInfo, task string, timestamp time.Time) string {
	return fmt.Sprintf(`# Claude Task - %s

**Repository:** %s
**Path:** %s
**Started:** %s

## Task Description

%s

## Progress

- [ ] Task started

## Notes

<!-- Add your notes and progress here -->

---

`, repo.Name, repo.Name, repo.Path, timestamp.Format("2006-01-02 15:04:05"), task)
}

// generateContinueTaskContent 이어하기 작업 내용 생성
func (rm *RepositoryManager) generateContinueTaskContent(task string, timestamp time.Time) string {
	return fmt.Sprintf(`## Continue Task - %s

**New Task:** %s

### Progress Update

- [ ] Continued task

### Additional Notes

<!-- Add continuation notes here -->

`, timestamp.Format("2006-01-02 15:04:05"), task)
}

// SaveTaskRecord 작업 기록 저장
func (rm *RepositoryManager) SaveTaskRecord(repo *RepositoryInfo, record TaskRecord) error {
	repo.TaskHistory = append(repo.TaskHistory, record)
	repo.LastTaskTime = record.Timestamp
	return rm.saveTaskHistory(repo)
}

// loadTaskHistory 작업 히스토리 로드 (구현 필요)
func (rm *RepositoryManager) loadTaskHistory(repo *RepositoryInfo) error {
	// JSON 파일에서 히스토리 로드하는 로직 추가 예정
	return nil
}

// saveTaskHistory 작업 히스토리 저장 (구현 필요)
func (rm *RepositoryManager) saveTaskHistory(repo *RepositoryInfo) error {
	// JSON 파일로 히스토리 저장하는 로직 추가 예정
	return nil
}

// ListRepositories 모든 레포지토리 목록 조회
func (rm *RepositoryManager) ListRepositories() ([]*RepositoryInfo, error) {
	var repositories []*RepositoryInfo

	err := filepath.Walk(rm.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// .git 디렉토리가 있는 경우 Git 레포지토리로 인식
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repoName := filepath.Base(repoPath)

			if repoInfo, err := rm.analyzeRepository(repoPath, repoName); err == nil {
				repositories = append(repositories, repoInfo)
			}
			return filepath.SkipDir
		}
		return nil
	})

	return repositories, err
}