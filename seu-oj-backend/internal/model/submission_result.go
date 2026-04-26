package model

import "time"

type SubmissionResult struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SubmissionID uint64    `gorm:"column:submission_id;index:idx_submission_results_submission_id" json:"submission_id"`
	TestcaseID   uint64    `gorm:"column:testcase_id" json:"testcase_id"`
	Status       string    `gorm:"column:status" json:"status"`
	RuntimeMS    *int      `gorm:"column:runtime_ms" json:"runtime_ms"`
	MemoryKB     *int      `gorm:"column:memory_kb" json:"memory_kb"`
	ErrorMsg     string    `gorm:"column:error_message" json:"error_message"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SubmissionResult) TableName() string {
	return "submission_results"
}
