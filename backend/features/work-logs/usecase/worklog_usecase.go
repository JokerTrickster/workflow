package usecase

import (
	"main/features/work-logs/model/request"
	"main/features/work-logs/model/response"
	"main/features/work-logs/repository"
)

type WorkLogUseCase struct {
	repo *repository.WorkLogRepository
}

func NewWorkLogUseCase() *WorkLogUseCase {
	return &WorkLogUseCase{
		repo: repository.NewWorkLogRepository(),
	}
}

func (uc *WorkLogUseCase) GetWorkLogs(req *request.GetWorkLogsRequest) ([]response.DailyWorkLog, error) {
	// Validate repository
	if err := uc.repo.ValidateRepository(req.Repository); err != nil {
		return nil, err
	}

	// Validate dates
	if err := uc.repo.ValidateDate(req.StartDate); err != nil {
		return nil, err
	}
	if err := uc.repo.ValidateDate(req.EndDate); err != nil {
		return nil, err
	}

	return uc.repo.GetWorkLogs(req.Repository, req.StartDate, req.EndDate)
}

func (uc *WorkLogUseCase) CreateWorkLogEntry(req *request.CreateWorkLogEntryRequest) (*response.CreateWorkLogEntryResponse, error) {
	// Validate repository
	if err := uc.repo.ValidateRepository(req.Repository); err != nil {
		return nil, err
	}

	// Create work log entry
	if err := uc.repo.CreateWorkLogEntry(req.Repository, req.Entry); err != nil {
		return nil, err
	}

	return &response.CreateWorkLogEntryResponse{Success: true}, nil
}
