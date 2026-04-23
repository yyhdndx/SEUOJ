package model

import "time"

type Problem struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DisplayID     string    `gorm:"column:display_id" json:"display_id"`
	Title         string    `gorm:"column:title" json:"title"`
	Description   string    `gorm:"column:description" json:"description"`
	InputDesc     string    `gorm:"column:input_desc" json:"input_desc"`
	OutputDesc    string    `gorm:"column:output_desc" json:"output_desc"`
	SampleInput   string    `gorm:"column:sample_input" json:"sample_input"`
	SampleOutput  string    `gorm:"column:sample_output" json:"sample_output"`
	Hint          string    `gorm:"column:hint" json:"hint"`
	Source        string    `gorm:"column:source" json:"source"`
	JudgeMode     string    `gorm:"column:judge_mode" json:"judge_mode"`
	Difficulty    int       `gorm:"column:difficulty" json:"difficulty"`
	TimeLimitMS   int       `gorm:"column:time_limit_ms" json:"time_limit_ms"`
	MemoryLimitMB int       `gorm:"column:memory_limit_mb" json:"memory_limit_mb"`
	Visible       bool      `gorm:"column:visible" json:"visible"`
	CreatedBy     uint64    `gorm:"column:created_by" json:"created_by"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Problem) TableName() string {
	return "problems"
}
