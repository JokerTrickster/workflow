package usecase

import (
	"context"
	"fmt"
	_interface "main/features/claude/model/interface"
	"main/features/claude/model/request"

	"time"
)

type RunTasksClaudeUseCase struct {
	Repository     _interface.IRunTasksClaudeRepository
	ContextTimeout time.Duration
}

func NewRunTasksClaudeUseCase(repo _interface.IRunTasksClaudeRepository, timeout time.Duration) _interface.IRunTasksClaudeUseCase {
	return &RunTasksClaudeUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *RunTasksClaudeUseCase) RunTasks(c context.Context, req *request.ReqRunTasksClaude) error {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()
	fmt.Println(ctx)

	return nil
}
