package model

type CleanupResult struct {
	Total           int `json:"total"`
	Completed       int `json:"completed"`
	SuccessfulCount int `json:"successful-count"`
	FailedCount     int `json:"failed-count"`

	Deleted int `json:"delete-count"`
	Error   int `json:"error-count"`

	Status string `json:"status"`
}

func (jr *CleanupResult) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"total":            jr.Total,
		"completed":        jr.Completed,
		"successful_count": jr.SuccessfulCount,
		"failed_count":     jr.FailedCount,
		"deleted":          jr.Deleted,
		"error":            jr.Error,
		"status":           jr.Status,
	}
}

type TaskInput struct {
	Namespaces       []string
	DelayFrequency   int
	SuccessThreshold int
	FailureThreshold int
	ProcessJobs      bool
	ProcessPods      bool
}
