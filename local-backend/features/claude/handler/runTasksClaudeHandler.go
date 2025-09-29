package handler

import (
	"context"
	_interface "main/features/claude/model/interface"
	"main/features/claude/model/request"
	"main/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RunTasksClaudeHandler struct {
	UseCase _interface.IRunTasksClaudeUseCase
}

func NewRunTasksClaudeHandler(c *echo.Echo, useCase _interface.IRunTasksClaudeUseCase) _interface.IRunTasksClaudeHandler {
	handler := &RunTasksClaudeHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/claude/tasks/run", handler.RunTasks)
	return handler
}

// 태스크 실행
// @Router /v0.1/claude/tasks/run [post]
// @Summary 태스크 실행 (Claude CLI)
// @Description 지정된 저장소에서 Claude CLI 명령어를 실행합니다
// @Description
// @Description **기능:**
// @Description - repository_name으로 지정된 저장소(~/project/git-repository/JokerTrickster/{repository_name})에서 작업 실행
// @Description - interactive 모드: 여러 작업을 순차적으로 실행 (줄바꿈 또는 세미콜론으로 구분)
// @Description - 일반 모드: 단일 작업을 한 번에 실행
// @Description - --dangerously-skip-permissions 플래그로 승인 요청 없이 자동 실행
// @Description
// @Description **필수 파라미터:**
// @Description - tasks: 실행할 작업 내용
// @Description - repository_name: 작업할 저장소 이름 (필수)
// @Description
// @Description **Interactive 모드 예시:**
// @Description ```json
// @Description {
// @Description   "tasks": "echo 첫번째 작업\necho 두번째 작업\necho 세번째 작업",
// @Description   "repository_name": "JokerTrickster",
// @Description   "interactive": true
// @Description }
// @Description ```
// @Description
// @Description ■ **에러 코드 (HTTP 400)**
// @Description - PARAM_BAD : 파라미터 오류 또는 유효성 검증 실패
// @Description - REPOSITORY_NAME_REQUIRED : repository_name 필수 파라미터 누락
// @Description
// @Description ■ **에러 코드 (HTTP 500)**
// @Description - INTERNAL_SERVER : 내부 로직 처리 실패
// @Description - REPOSITORY_NOT_FOUND : 저장소 디렉토리를 찾을 수 없음
// @Description - CLAUDE_CLI_EXECUTION_FAILED : Claude CLI 실행 실패
// @Param json body request.ReqRunTasksClaude true "작업 실행 요청"
// @Produce json
// @Success 200 {object} bool "작업 실행 완료"
// @Failure 400 {object} error "잘못된 요청"
// @Failure 500 {object} error "서버 내부 오류"
// @Tags claude
func (d *RunTasksClaudeHandler) RunTasks(c echo.Context) error {
	ctx := context.Background()
	req := &request.ReqRunTasksClaude{}
	if err := utils.ValidateReq(c, req); err != nil {
		return err
	}
	err := d.UseCase.RunTasks(ctx, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, true)
}
