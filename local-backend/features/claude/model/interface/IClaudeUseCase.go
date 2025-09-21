package _interface

import (
	"context"

	"main/features/claude/model/request"
)

type IRunTasksClaudeUseCase interface {
	RunTasks(c context.Context, req *request.ReqRunTasksClaude) error
}
