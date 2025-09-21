package _interface

import (
	"context"

	"main/features/claude/model/request"
)

type IRunTasksClaudeUseCase interface {
	RunTasks(c context.Context, req *request.ReqRunTasksClaude) error
}

type ICloneRepositoriesUseCase interface {
	CloneRepositories(c context.Context, req *request.ReqCloneRepositories) (*request.ResCloneRepositories, error)
}
