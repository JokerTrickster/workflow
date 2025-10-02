package response

import "main/features/work-logs/model/request"

type DailyWorkLog struct {
	Date       string                    `json:"date"`
	Repository string                    `json:"repository"`
	Entries    []request.WorkLogEntry    `json:"entries"`
}

type CreateWorkLogEntryResponse struct {
	Success bool `json:"success"`
}
