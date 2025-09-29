package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	_interface "main/features/claude/model/interface"
	"main/features/claude/model/request"
	"main/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CloneRepositoriesUseCase struct {
	Repository        _interface.ICloneRepositoriesRepository
	ContextTimeout    time.Duration
	RepositoryManager *utils.RepositoryManager
}

// GitHub API 응답 구조체
type GitHubRepository struct {
	Name        string `json:"name"`
	CloneURL    string `json:"clone_url"`
	HTMLURL     string `json:"html_url"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
	Archived    bool   `json:"archived"`
	Disabled    bool   `json:"disabled"`
}

func NewCloneRepositoriesUseCase(repo _interface.ICloneRepositoriesRepository, timeout time.Duration) _interface.ICloneRepositoriesUseCase {
	return &CloneRepositoriesUseCase{
		Repository:        repo,
		ContextTimeout:    timeout,
		RepositoryManager: utils.NewRepositoryManager(""),
	}
}

func (d *CloneRepositoriesUseCase) CloneRepositories(c context.Context, req *request.ReqCloneRepositories) (*request.ResCloneRepositories, error) {
	_, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	// 기본값 설정
	username := req.GitHubUsername
	if username == "" {
		username = "JokerTrickster"
	}

	// GitHub 사용자명 유효성 검증
	if err := d.validateGitHubUsername(username); err != nil {
		log.Printf("Invalid GitHub username: %v", err)
		return nil, fmt.Errorf("invalid GitHub username: %w", err)
	}

	// 타겟 디렉토리 설정 (환경변수 우선, 요청 파라미터 다음, 기본값 마지막)
	targetDir := os.Getenv("GITHUB_DEFAULT_TARGET_DIR")
	if targetDir == "" {
		targetDir = utils.GetJokerTricksterPath()
	}
	if req.TargetDirectory != "" {
		targetDir = req.TargetDirectory
	}

	log.Printf("Starting GitHub repository cloning for user: %s", username)
	log.Printf("Target directory: %s", targetDir)

	// GitHub API를 통한 저장소 목록 가져오기
	repos, err := d.fetchGitHubRepositories(username, req.GitHubToken)
	if err != nil {
		log.Printf("Failed to fetch repositories: %v", err)
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	log.Printf("Found %d repositories for user %s", len(repos), username)

	// 타겟 디렉토리 생성
	if err := d.ensureDirectoryExists(targetDir); err != nil {
		log.Printf("Failed to create target directory: %v", err)
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Git clone 작업 수행
	details := make([]request.RepositoryResult, len(repos))
	clonedCount := 0
	skippedCount := 0
	failedCount := 0

	for i, repo := range repos {
		result := d.cloneRepository(repo, targetDir)
		details[i] = result

		switch result.Status {
		case "cloned":
			clonedCount++
		case "skipped":
			skippedCount++
		case "failed":
			failedCount++
		}
	}

	// 전체 작업 상태 결정
	overallStatus := "success"
	if failedCount > 0 && clonedCount == 0 && skippedCount == 0 {
		overallStatus = "failed"
	} else if failedCount > 0 {
		overallStatus = "partial"
	}

	response := &request.ResCloneRepositories{
		Status:            overallStatus,
		TotalRepositories: len(repos),
		ClonedCount:       clonedCount,
		SkippedCount:      skippedCount,
		FailedCount:       failedCount,
		Details:           details,
	}

	log.Println("GitHub repository cloning completed successfully")
	return response, nil
}

// fetchGitHubRepositories GitHub API를 통해 사용자/조직의 저장소 목록을 가져옵니다
func (d *CloneRepositoriesUseCase) fetchGitHubRepositories(username, token string) ([]GitHubRepository, error) {
	var allRepos []GitHubRepository
	page := 1
	perPage := 100

	for {
		repos, hasNext, err := d.fetchRepositoriesPage(username, token, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		allRepos = append(allRepos, repos...)

		if !hasNext {
			break
		}
		page++
	}

	return allRepos, nil
}

// fetchRepositoriesPage GitHub API에서 특정 페이지의 저장소를 가져옵니다
func (d *CloneRepositoriesUseCase) fetchRepositoriesPage(username, token string, page, perPage int) ([]GitHubRepository, bool, error) {
	baseURL := os.Getenv("GITHUB_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/users/%s/repos?page=%d&per_page=%d&type=all", baseURL, username, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	// GitHub 토큰이 있으면 Authorization 헤더 추가
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	// User-Agent 헤더 추가 (GitHub API 권장사항)
	req.Header.Set("User-Agent", "workflow-local-backend/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, d.handleGitHubAPIError(resp.StatusCode, body, username)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response body: %w", err)
	}

	var repos []GitHubRepository
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, false, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Link 헤더를 확인하여 다음 페이지가 있는지 판단
	linkHeader := resp.Header.Get("Link")
	hasNext := d.hasNextPage(linkHeader)

	log.Printf("Fetched %d repositories from page %d", len(repos), page)
	return repos, hasNext, nil
}

// hasNextPage Link 헤더를 파싱하여 다음 페이지가 있는지 확인합니다
func (d *CloneRepositoriesUseCase) hasNextPage(linkHeader string) bool {
	// Link 헤더 예시: <https://api.github.com/user/repos?page=2>; rel="next", <https://api.github.com/user/repos?page=5>; rel="last"
	return linkHeader != "" && strings.Contains(linkHeader, `rel="next"`)
}

// ensureDirectoryExists 디렉토리가 존재하지 않으면 생성합니다
func (d *CloneRepositoriesUseCase) ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("Created directory: %s", dir)
	}
	return nil
}

// cloneRepository 개별 저장소를 복제합니다
func (d *CloneRepositoriesUseCase) cloneRepository(repo GitHubRepository, targetDir string) request.RepositoryResult {
	repoPath := filepath.Join(targetDir, repo.Name)

	// 중복 체크: 이미 존재하는지 확인
	if d.repositoryExists(repoPath) {
		log.Printf("Repository %s already exists, skipping", repo.Name)
		return request.RepositoryResult{
			Name:      repo.Name,
			CloneURL:  repo.CloneURL,
			Status:    "skipped",
			LocalPath: repoPath,
		}
	}

	// Git clone 실행
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", repo.CloneURL, repoPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorMsg := d.parseGitCloneError(err, string(output))
		log.Printf("Failed to clone repository %s: %s", repo.Name, errorMsg)
		return request.RepositoryResult{
			Name:         repo.Name,
			CloneURL:     repo.CloneURL,
			Status:       "failed",
			ErrorMessage: errorMsg,
		}
	}

	log.Printf("Successfully cloned repository: %s", repo.Name)
	return request.RepositoryResult{
		Name:      repo.Name,
		CloneURL:  repo.CloneURL,
		Status:    "cloned",
		LocalPath: repoPath,
	}
}

// repositoryExists 저장소가 이미 존재하는지 확인합니다
func (d *CloneRepositoriesUseCase) repositoryExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		// .git 디렉토리가 있는지 확인하여 유효한 git 저장소인지 판단
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return true
		}

		// .git 파일이 있는지 확인 (submodule이나 worktree의 경우)
		if gitFile, err := os.Stat(gitDir); err == nil && !gitFile.IsDir() {
			return true
		}
	}
	return false
}

// handleGitHubAPIError GitHub API 에러를 처리하고 의미있는 에러 메시지를 반환합니다
func (d *CloneRepositoriesUseCase) handleGitHubAPIError(statusCode int, body []byte, username string) error {
	switch statusCode {
	case http.StatusNotFound:
		return fmt.Errorf("GitHub user '%s' not found or has no public repositories", username)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitHub API authentication failed - check your token")
	case http.StatusForbidden:
		// GitHub API rate limit 확인
		var githubError struct {
			Message          string `json:"message"`
			DocumentationURL string `json:"documentation_url"`
		}
		if err := json.Unmarshal(body, &githubError); err == nil {
			if strings.Contains(githubError.Message, "rate limit") {
				return fmt.Errorf("GitHub API rate limit exceeded - try again later or provide authentication token")
			}
			return fmt.Errorf("GitHub API access forbidden: %s", githubError.Message)
		}
		return fmt.Errorf("GitHub API access forbidden")
	case http.StatusTooManyRequests:
		return fmt.Errorf("GitHub API rate limit exceeded - try again later")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("GitHub API is temporarily unavailable")
	default:
		return fmt.Errorf("GitHub API returned status %d: %s", statusCode, string(body))
	}
}

// validateGitHubUsername GitHub 사용자명의 유효성을 검증합니다
func (d *CloneRepositoriesUseCase) validateGitHubUsername(username string) error {
	if username == "" {
		return fmt.Errorf("GitHub username cannot be empty")
	}

	// GitHub 사용자명 규칙 검증
	if len(username) > 39 {
		return fmt.Errorf("GitHub username too long (max 39 characters)")
	}

	// 기본적인 문자 검증 (GitHub 사용자명은 alphanumeric + hyphens)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			 (char >= '0' && char <= '9') || char == '-') {
			return fmt.Errorf("GitHub username contains invalid characters")
		}
	}

	// 하이픈으로 시작하거나 끝날 수 없음
	if strings.HasPrefix(username, "-") || strings.HasSuffix(username, "-") {
		return fmt.Errorf("GitHub username cannot start or end with hyphen")
	}

	return nil
}

// parseGitCloneError Git clone 에러를 분석하여 사용자 친화적인 메시지를 반환합니다
func (d *CloneRepositoriesUseCase) parseGitCloneError(err error, output string) string {
	outputLower := strings.ToLower(output)

	// 일반적인 Git clone 에러 패턴들
	if strings.Contains(outputLower, "repository not found") {
		return "Repository not found or access denied"
	}

	if strings.Contains(outputLower, "permission denied") {
		return "Permission denied - repository may be private"
	}

	if strings.Contains(outputLower, "authentication failed") {
		return "Authentication failed - check your credentials"
	}

	if strings.Contains(outputLower, "timeout") || strings.Contains(err.Error(), "timeout") {
		return "Clone operation timed out - repository may be too large"
	}

	if strings.Contains(outputLower, "network") || strings.Contains(outputLower, "connection") {
		return "Network connection error"
	}

	if strings.Contains(outputLower, "destination path") && strings.Contains(outputLower, "already exists") {
		return "Destination directory already exists"
	}

	if strings.Contains(outputLower, "disk") || strings.Contains(outputLower, "space") {
		return "Insufficient disk space"
	}

	// 기본 에러 메시지
	if output != "" {
		// 에러 출력에서 첫 번째 줄만 반환 (보통 가장 유용한 정보)
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) > 0 && lines[0] != "" {
			return fmt.Sprintf("Git error: %s", lines[0])
		}
	}

	return fmt.Sprintf("Git clone failed: %v", err)
}