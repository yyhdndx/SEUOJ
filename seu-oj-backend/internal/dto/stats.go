package dto

type CountItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type RecentActivityItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type OverviewStatsResponse struct {
	ProblemsTotal       int64 `json:"problems_total"`
	VisibleProblems     int64 `json:"visible_problems"`
	UsersTotal          int64 `json:"users_total"`
	SubmissionsTotal    int64 `json:"submissions_total"`
	AcceptedSubmissions int64 `json:"accepted_submissions"`
}

type UserStatsResponse struct {
	UserID               uint64               `json:"user_id"`
	Username             string               `json:"username"`
	SubmissionsTotal     int64                `json:"submissions_total"`
	AcceptedSubmissions  int64                `json:"accepted_submissions"`
	AcceptedProblems     int64                `json:"accepted_problems"`
	PendingSubmissions   int64                `json:"pending_submissions"`
	RunningSubmissions   int64                `json:"running_submissions"`
	AverageRuntimeMS     *float64             `json:"average_runtime_ms"`
	StatusBreakdown      []CountItem          `json:"status_breakdown"`
	LanguageBreakdown    []CountItem          `json:"language_breakdown"`
	RecentActivity       []RecentActivityItem `json:"recent_activity"`
}

type AdminStatsResponse struct {
	QueueLength       int64       `json:"queue_length"`
	UsersTotal        int64       `json:"users_total"`
	ProblemsTotal     int64       `json:"problems_total"`
	HiddenProblems    int64       `json:"hidden_problems"`
	SubmissionsTotal  int64       `json:"submissions_total"`
	PendingSubmissions int64      `json:"pending_submissions"`
	RunningSubmissions int64      `json:"running_submissions"`
	SystemErrors      int64       `json:"system_errors"`
	StatusBreakdown   []CountItem `json:"status_breakdown"`
}
