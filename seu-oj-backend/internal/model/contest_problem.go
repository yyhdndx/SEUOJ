package model

import "time"

type ContestProblem struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContestID    uint64    `gorm:"column:contest_id;not null;index;uniqueIndex:uk_contest_problem_code,priority:1;uniqueIndex:uk_contest_problem_id,priority:1" json:"contest_id"`
	ProblemID    uint64    `gorm:"column:problem_id;not null;index;uniqueIndex:uk_contest_problem_id,priority:2" json:"problem_id"`
	ProblemCode  string    `gorm:"column:problem_code;size:16;not null;uniqueIndex:uk_contest_problem_code,priority:2" json:"problem_code"`
	DisplayOrder int       `gorm:"column:display_order;not null;default:1;index" json:"display_order"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ContestProblem) TableName() string {
	return "contest_problems"
}
