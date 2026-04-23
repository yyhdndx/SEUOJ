package model

import "time"

type Submission struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      uint64     `gorm:"column:user_id;index:idx_submissions_user_id;index:idx_submissions_user_status,priority:1;index:idx_submissions_user_created,priority:1" json:"user_id"`
	ProblemID   uint64     `gorm:"column:problem_id;index:idx_submissions_problem_status,priority:1" json:"problem_id"`
	ContestID   *uint64    `gorm:"column:contest_id;index:idx_submissions_contest_created,priority:1" json:"contest_id"`
	IsPractice  bool       `gorm:"column:is_practice;not null;default:false" json:"is_practice"`
	Language    string     `gorm:"column:language" json:"language"`
	Code        string     `gorm:"column:code" json:"code"`
	Status      string     `gorm:"column:status;index:idx_submissions_status;index:idx_submissions_user_status,priority:2;index:idx_submissions_problem_status,priority:2" json:"status"`
	PassedCount int        `gorm:"column:passed_count" json:"passed_count"`
	TotalCount  int        `gorm:"column:total_count" json:"total_count"`
	RuntimeMS   *int       `gorm:"column:runtime_ms" json:"runtime_ms"`
	MemoryKB    *int       `gorm:"column:memory_kb" json:"memory_kb"`
	CompileInfo string     `gorm:"column:compile_info" json:"compile_info"`
	ErrorMsg    string     `gorm:"column:error_message" json:"error_message"`
	CreatedAt   time.Time  `gorm:"column:created_at;index:idx_submissions_contest_created,priority:2;index:idx_submissions_user_created,priority:2" json:"created_at"`
	JudgedAt    *time.Time `gorm:"column:judged_at" json:"judged_at"`
}

func (Submission) TableName() string {
	return "submissions"
}
