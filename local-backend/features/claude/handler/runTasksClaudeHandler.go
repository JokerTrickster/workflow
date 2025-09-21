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
// @Summary 태스크 실행
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description USER_NOT_EXIST : 유저가 존재하지 않음
// @Description USER_ALREADY_EXISTED : 유저가 이미 존재
// @Description USER_GOOGLE_ALREADY_EXISTED : 구글 계정이 이미 존재
// @Description PASSWORD_NOT_MATCH : 비밀번호가 일치하지 않음
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Description INTERNAL_DB : DB 처리 실패
// @Param json body request.ReqRunTasksClaude true "json body "
// @Produce json
// @Success 200 {object} bool
// @Failure 400 {object} error
// @Failure 500 {object} error
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
