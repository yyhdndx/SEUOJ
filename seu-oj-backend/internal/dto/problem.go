package dto

import "time"

type CreateProblemRequest struct {
	DisplayID     string                       `json:"display_id" binding:"required,max=32"`
	Title         string                       `json:"title" binding:"required,max=255"`
	Description   string                       `json:"description" binding:"required"`
	InputDesc     string                       `json:"input_desc"`
	OutputDesc    string                       `json:"output_desc"`
	SampleInput   string                       `json:"sample_input"`
	SampleOutput  string                       `json:"sample_output"`
	Hint          string                       `json:"hint"`
	Source        string                       `json:"source" binding:"max=255"`
	JudgeMode     string                       `json:"judge_mode" binding:"required,oneof=standard"`
	TimeLimitMS   int                          `json:"time_limit_ms" binding:"required,min=1"`
	MemoryLimitMB int                          `json:"memory_limit_mb" binding:"required,min=1"`
	Visible       bool                         `json:"visible"`
	Testcases     []CreateProblemTestcaseInput `json:"testcases" binding:"required,min=1,dive"`
}

type CreateProblemTestcaseInput struct {
	CaseType   string `json:"case_type" binding:"required,oneof=sample hidden"`
	InputData  string `json:"input_data" binding:"required"`
	OutputData string `json:"output_data" binding:"required"`
	Score      int    `json:"score" binding:"min=0"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
}

type ProblemListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
}

type AdminProblemListQuery struct {
	Page          int    `form:"page" binding:"omitempty,min=1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword       string `form:"keyword"`
	IncludeHidden bool   `form:"include_hidden"`
}

type ProblemListItem struct {
	ID            uint64 `json:"id"`
	DisplayID     string `json:"display_id"`
	Title         string `json:"title"`
	JudgeMode     string `json:"judge_mode"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	Visible       bool   `json:"visible"`
	CreatedAt     string `json:"created_at"`
}

type ProblemListResponse struct {
	List     []ProblemListItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type CreateProblemResponse struct {
	ProblemID uint64 `json:"problem_id"`
}

type ProblemTestcaseResponse struct {
	ID         uint64 `json:"id"`
	CaseType   string `json:"case_type"`
	InputData  string `json:"input_data"`
	OutputData string `json:"output_data,omitempty"`
	Score      int    `json:"score"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
}

type ProblemDetailResponse struct {
	ID            uint64                    `json:"id"`
	DisplayID     string                    `json:"display_id"`
	Title         string                    `json:"title"`
	Description   string                    `json:"description"`
	InputDesc     string                    `json:"input_desc"`
	OutputDesc    string                    `json:"output_desc"`
	SampleInput   string                    `json:"sample_input"`
	SampleOutput  string                    `json:"sample_output"`
	Hint          string                    `json:"hint"`
	Source        string                    `json:"source"`
	JudgeMode     string                    `json:"judge_mode"`
	TimeLimitMS   int                       `json:"time_limit_ms"`
	MemoryLimitMB int                       `json:"memory_limit_mb"`
	Visible       bool                      `json:"visible"`
	CreatedBy     uint64                    `json:"created_by"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
	Testcases     []ProblemTestcaseResponse `json:"testcases"`
	Solutions     []ProblemSolutionResponse `json:"solutions"`
}

type ProblemStatsResponse struct {
	ProblemID            uint64      `json:"problem_id"`
	SubmissionsTotal     int64       `json:"submissions_total"`
	AcceptedSubmissions  int64       `json:"accepted_submissions"`
	AcceptedUsers        int64       `json:"accepted_users"`
	AcceptedRate         float64     `json:"accepted_rate"`
	LanguageBreakdown    []CountItem `json:"language_breakdown"`
}


