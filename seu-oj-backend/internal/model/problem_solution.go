package model

import "time"

type ProblemSolution struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProblemID  uint64    `gorm:"column:problem_id;not null;index:idx_problem_solutions_problem_visibility,priority:1" json:"problem_id"`
	Title      string    `gorm:"column:title;size:255;not null" json:"title"`
	Content    string    `gorm:"column:content;type:longtext;not null" json:"content"`
	Visibility string    `gorm:"column:visibility;size:32;not null;default:public;index:idx_problem_solutions_problem_visibility,priority:2" json:"visibility"`
	AuthorID   uint64    `gorm:"column:author_id;not null" json:"author_id"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ProblemSolution) TableName() string {
	return "problem_solutions"
}
