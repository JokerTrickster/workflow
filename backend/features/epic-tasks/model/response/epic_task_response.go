package response

import "main/features/epic-tasks/model/request"

type EpicTaskResponse struct {
	Task request.EpicTaskFile `json:"task"`
}

type EpicTasksListResponse struct {
	Tasks []request.EpicTaskFile `json:"tasks"`
}

type DeleteEpicTaskResponse struct {
	Success bool `json:"success"`
}
