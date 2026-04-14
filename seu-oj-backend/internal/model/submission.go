package model

import "time"

type Submission struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      uint64     `gorm:"column:user_id" json:"user_id"`
	ProblemID   uint64     `gorm:"column:problem_id" json:"problem_id"`
	ContestID   *uint64    `gorm:"column:contest_id;index:idx_submissions_contest_id" json:"contest_id"`
	IsPractice  bool       `gorm:"column:is_practice;not null;default:false" json:"is_practice"`
	Language    string     `gorm:"column:language" json:"language"`
	Code        string     `gorm:"column:code" json:"code"`
	Status      string     `gorm:"column:status" json:"status"`
	PassedCount int        `gorm:"column:passed_count" json:"passed_count"`
	TotalCount  int        `gorm:"column:total_count" json:"total_count"`
	RuntimeMS   *int       `gorm:"column:runtime_ms" json:"runtime_ms"`
	MemoryKB    *int       `gorm:"column:memory_kb" json:"memory_kb"`
	CompileInfo string     `gorm:"column:compile_info" json:"compile_info"`
	ErrorMsg    string     `gorm:"column:error_message" json:"error_message"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	JudgedAt    *time.Time `gorm:"column:judged_at" json:"judged_at"`
}

func (Submission) TableName() string {
	return "submissions"
}
