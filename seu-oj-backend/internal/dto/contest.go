package dto

import "time"

type ContestListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status" binding:"omitempty,oneof=upcoming running ended"`
}

type CreateContestRequest struct {
	Title            string                    `json:"title" binding:"required,max=255"`
	Description      string                    `json:"description"`
	RuleType         string                    `json:"rule_type" binding:"required,oneof=acm"`
	StartTime        time.Time                 `json:"start_time" binding:"required"`
	EndTime          time.Time                 `json:"end_time" binding:"required"`
	IsPublic         bool                      `json:"is_public"`
	AllowPractice    bool                      `json:"allow_practice"`
	RanklistFreezeAt *time.Time                `json:"ranklist_freeze_at"`
	Problems         []CreateContestProblemDTO `json:"problems" binding:"required,min=1,dive"`
}

type CreateContestProblemDTO struct {
	ProblemID    uint64 `json:"problem_id" binding:"required,min=1"`
	ProblemCode  string `json:"problem_code" binding:"omitempty,max=16"`
	DisplayOrder int    `json:"display_order" binding:"omitempty,min=1"`
}

type ContestListItem struct {
	ID               uint64     `json:"id"`
	Title            string     `json:"title"`
	RuleType         string     `json:"rule_type"`
	Status           string     `json:"status"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          time.Time  `json:"end_time"`
	IsPublic         bool       `json:"is_public"`
	AllowPractice    bool       `json:"allow_practice"`
	RanklistFreezeAt *time.Time `json:"ranklist_freeze_at"`
	RanklistFrozen   bool       `json:"ranklist_frozen"`
	RegisteredCount  int64      `json:"registered_count"`
	ProblemCount     int64      `json:"problem_count"`
}

type ContestListResponse struct {
	List     []ContestListItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type ContestProblemItem struct {
	ID            uint64 `json:"id"`
	ProblemID     uint64 `json:"problem_id"`
	ProblemCode   string `json:"problem_code"`
	DisplayOrder  int    `json:"display_order"`
	Title         string `json:"title"`
	JudgeMode     string `json:"judge_mode"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
}

type ContestDetailResponse struct {
	ID               uint64               `json:"id"`
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	RuleType         string               `json:"rule_type"`
	Status           string               `json:"status"`
	StartTime        time.Time            `json:"start_time"`
	EndTime          time.Time            `json:"end_time"`
	IsPublic         bool                 `json:"is_public"`
	AllowPractice    bool                 `json:"allow_practice"`
	RanklistFreezeAt *time.Time           `json:"ranklist_freeze_at"`
	RanklistFrozen   bool                 `json:"ranklist_frozen"`
	CreatedBy        uint64               `json:"created_by"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	RegisteredCount  int64                `json:"registered_count"`
	ProblemCount     int64                `json:"problem_count"`
	Problems         []ContestProblemItem `json:"problems,omitempty"`
}

type ContestMeResponse struct {
	ContestID       uint64 `json:"contest_id"`
	Registered      bool   `json:"registered"`
	ContestStatus   string `json:"contest_status"`
	CanRegister     bool   `json:"can_register"`
	CanViewProblems bool   `json:"can_view_problems"`
	CanSubmit       bool   `json:"can_submit"`
	PracticeEnabled bool   `json:"practice_enabled"`
}

type ContestRegisterResponse struct {
	ContestID  uint64 `json:"contest_id"`
	Registered bool   `json:"registered"`
}

type ContestProblemListResponse struct {
	ContestID     uint64               `json:"contest_id"`
	ContestTitle  string               `json:"contest_title"`
	ContestStatus string               `json:"contest_status"`
	List          []ContestProblemItem `json:"list"`
}

type ContestRanklistProblem struct {
	ProblemID    uint64 `json:"problem_id"`
	ProblemCode  string `json:"problem_code"`
	DisplayOrder int    `json:"display_order"`
}

type ContestRanklistCell struct {
	ProblemID         uint64     `json:"problem_id"`
	ProblemCode       string     `json:"problem_code"`
	Solved            bool       `json:"solved"`
	WrongAttempts     int        `json:"wrong_attempts"`
	AcceptedAt        *time.Time `json:"accepted_at"`
	PenaltyMinutes    int        `json:"penalty_minutes"`
	LatestStatus      string     `json:"latest_status"`
	SubmissionCount   int        `json:"submission_count"`
	FrozenSubmissions int        `json:"frozen_submissions"`
}

type ContestRanklistItem struct {
	Rank            int                   `json:"rank"`
	UserID          uint64                `json:"user_id"`
	Username        string                `json:"username"`
	StudentID       string                `json:"userid"`
	SolvedCount     int                   `json:"solved_count"`
	PenaltyMinutes  int                   `json:"penalty_minutes"`
	LastAcceptedAt  *time.Time            `json:"last_accepted_at"`
	SubmissionCount int                   `json:"submission_count"`
	Cells           []ContestRanklistCell `json:"cells"`
}

type ContestRanklistResponse struct {
	ContestID        uint64                   `json:"contest_id"`
	ContestTitle     string                   `json:"contest_title"`
	ContestStatus    string                   `json:"contest_status"`
	RanklistFrozen   bool                     `json:"ranklist_frozen"`
	RanklistFreezeAt *time.Time               `json:"ranklist_freeze_at"`
	Problems         []ContestRanklistProblem `json:"problems"`
	List             []ContestRanklistItem    `json:"list"`
}

type CreateContestResponse struct {
	ContestID uint64 `json:"contest_id"`
}
