package handler

import (
	"context"
	_interface "main/features/claude/model/interface"
	"main/features/claude/model/request"
	"main/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CloneRepositoriesHandler struct {
	UseCase _interface.ICloneRepositoriesUseCase
}

func NewCloneRepositoriesHandler(c *echo.Echo, useCase _interface.ICloneRepositoriesUseCase) _interface.ICloneRepositoriesHandler {
	handler := &CloneRepositoriesHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/claude/repositories/clone", handler.CloneRepositories)
	return handler
}

// GitHub 저장소 일괄 복제
// @Router /v0.1/claude/repositories/clone [post]
// @Summary GitHub 저장소 일괄 복제
// @Description GitHub 사용자/조직의 모든 저장소를 로컬 디렉토리에 복제합니다.
// @Description
// @Description **기능:**
// @Description - GitHub API를 통해 지정된 사용자/조직의 모든 저장소 목록을 가져옵니다
// @Description - 각 저장소를 지정된 로컬 디렉토리에 Git clone으로 복제합니다
// @Description - 이미 존재하는 저장소는 건너뛰어 중복 다운로드를 방지합니다
// @Description - 기본 사용자: JokerTrickster, 기본 경로: /Users/mac/project/git-repository/JokerTrickster
// @Description
// @Description **요청 예시:**
// @Description ```json
// @Description {
// @Description   "github_username": "JokerTrickster",
// @Description   "github_token": "ghp_xxxxxxxxxxxx"
// @Description }
// @Description ```
// @Description
// @Description **응답 예시:**
// @Description ```json
// @Description {
// @Description   "status": "success",
// @Description   "total_repositories": 15,
// @Description   "cloned_count": 12,
// @Description   "skipped_count": 3,
// @Description   "failed_count": 0,
// @Description   "details": [
// @Description     {
// @Description       "name": "workflow",
// @Description       "clone_url": "https://github.com/JokerTrickster/workflow.git",
// @Description       "status": "cloned",
// @Description       "local_path": "/Users/mac/project/git-repository/JokerTrickster/workflow"
// @Description     }
// @Description   ]
// @Description }
// @Description ```
// @Description
// @Description ■ **에러 코드 (HTTP 400)**
// @Description - PARAM_BAD : 파라미터 오류 또는 유효성 검증 실패
// @Description - GITHUB_USER_NOT_FOUND : GitHub 사용자/조직이 존재하지 않음
// @Description - GITHUB_USERNAME_INVALID : GitHub 사용자명 형식이 올바르지 않음
// @Description - GITHUB_API_RATE_LIMIT : GitHub API 호출 한도 초과
// @Description - GITHUB_API_AUTH_FAILED : GitHub API 인증 실패
// @Description
// @Description ■ **에러 코드 (HTTP 500)**
// @Description - INTERNAL_SERVER : 내부 로직 처리 실패
// @Description - DIRECTORY_CREATE_FAILED : 대상 디렉토리 생성 실패
// @Description - GIT_CLONE_FAILED : Git 복제 작업 실패
// @Description - GITHUB_API_ERROR : GitHub API 서버 오류
// @Param json body request.ReqCloneRepositories true "저장소 복제 요청"
// @Produce json
// @Success 200 {object} request.ResCloneRepositories "복제 작업 완료"
// @Failure 400 {object} error "잘못된 요청"
// @Failure 500 {object} error "서버 내부 오류"
// @Tags claude
func (d *CloneRepositoriesHandler) CloneRepositories(c echo.Context) error {
	ctx := context.Background()
	req := &request.ReqCloneRepositories{}
	if err := utils.ValidateReq(c, req); err != nil {
		return err
	}
	response, err := d.UseCase.CloneRepositories(ctx, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response)
}