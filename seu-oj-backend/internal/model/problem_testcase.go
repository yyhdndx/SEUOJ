package model

import "time"

type ProblemTestcase struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement;index:idx_problem_testcases_problem_sort,priority:3" json:"id"`
	ProblemID  uint64    `gorm:"column:problem_id;index:idx_problem_testcases_problem_sort,priority:1" json:"problem_id"`
	CaseType   string    `gorm:"column:case_type" json:"case_type"`
	InputData  string    `gorm:"column:input_data" json:"input_data"`
	OutputData string    `gorm:"column:output_data" json:"output_data"`
	Score      int       `gorm:"column:score" json:"score"`
	SortOrder  int       `gorm:"column:sort_order;index:idx_problem_testcases_problem_sort,priority:2" json:"sort_order"`
	IsActive   bool      `gorm:"column:is_active" json:"is_active"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProblemTestcase) TableName() string {
	return "problem_testcases"
}
