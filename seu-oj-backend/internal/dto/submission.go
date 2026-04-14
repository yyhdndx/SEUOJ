package dto

import "time"

type CreateSubmissionRequest struct {
	ProblemID uint64  `json:"problem_id" binding:"required,min=1"`
	ContestID *uint64 `json:"contest_id"`
	Language  string  `json:"language" binding:"required,oneof=cpp c java python3 go rust"`
	Code      string  `json:"code" binding:"required"`
}

type SubmissionListQuery struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	UserID    uint64 `form:"user_id" binding:"omitempty,min=1"`
	ProblemID uint64 `form:"problem_id" binding:"omitempty,min=1"`
	ContestID uint64 `form:"contest_id" binding:"omitempty,min=1"`
	Status    string `form:"status"`
}

type CreateSubmissionResponse struct {
	SubmissionID uint64 `json:"submission_id"`
	Status       string `json:"status"`
}

type RejudgeSubmissionResponse struct {
	SubmissionID uint64 `json:"submission_id"`
	Status       string `json:"status"`
}

type RunSubmissionRequest struct {
	ProblemID uint64  `json:"problem_id" binding:"required,min=1"`
	ContestID *uint64 `json:"contest_id"`
	Language  string  `json:"language" binding:"required,oneof=cpp c java python3 go rust"`
	Code      string  `json:"code" binding:"required"`
}

type RunSubmissionCaseResult struct {
	CaseType       string `json:"case_type"`
	SortOrder      int    `json:"sort_order"`
	Status         string `json:"status"`
	RuntimeMS      *int   `json:"runtime_ms"`
	InputData      string `json:"input_data"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
	ErrorMessage   string `json:"error_message"`
}

type RunSubmissionResponse struct {
	Status      string                    `json:"status"`
	CompileInfo string                    `json:"compile_info"`
	ErrorMsg    string                    `json:"error_message"`
	Results     []RunSubmissionCaseResult `json:"results"`
}

type SubmissionListItem struct {
	ID          uint64     `json:"id"`
	UserID      uint64     `json:"user_id"`
	ProblemID   uint64     `json:"problem_id"`
	ContestID   *uint64    `json:"contest_id"`
	IsPractice  bool       `json:"is_practice"`
	Language    string     `json:"language"`
	Status      string     `json:"status"`
	PassedCount int        `json:"passed_count"`
	TotalCount  int        `json:"total_count"`
	RuntimeMS   *int       `json:"runtime_ms"`
	CreatedAt   time.Time  `json:"created_at"`
	JudgedAt    *time.Time `json:"judged_at"`
}

type SubmissionListResponse struct {
	List     []SubmissionListItem `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type SubmissionDetailResponse struct {
	ID          uint64                     `json:"id"`
	UserID      uint64                     `json:"user_id"`
	ProblemID   uint64                     `json:"problem_id"`
	ContestID   *uint64                    `json:"contest_id"`
	IsPractice  bool                       `json:"is_practice"`
	Language    string                     `json:"language"`
	Code        string                     `json:"code"`
	Status      string                     `json:"status"`
	PassedCount int                        `json:"passed_count"`
	TotalCount  int                        `json:"total_count"`
	RuntimeMS   *int                       `json:"runtime_ms"`
	MemoryKB    *int                       `json:"memory_kb"`
	CompileInfo string                     `json:"compile_info"`
	ErrorMsg    string                     `json:"error_message"`
	CreatedAt   time.Time                  `json:"created_at"`
	JudgedAt    *time.Time                 `json:"judged_at"`
	Results     []SubmissionResultResponse `json:"results"`
}

type SubmissionResultResponse struct {
	ID         uint64    `json:"id"`
	TestcaseID uint64    `json:"testcase_id"`
	Status     string    `json:"status"`
	RuntimeMS  *int      `json:"runtime_ms"`
	MemoryKB   *int      `json:"memory_kb"`
	ErrorMsg   string    `json:"error_message"`
	CreatedAt  time.Time `json:"created_at"`
}
