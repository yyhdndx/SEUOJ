package model

import "time"

type PlaylistProblem struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PlaylistID   uint64    `gorm:"column:playlist_id;not null;index" json:"playlist_id"`
	ProblemID    uint64    `gorm:"column:problem_id;not null;index" json:"problem_id"`
	DisplayOrder int       `gorm:"column:display_order;not null;default:1" json:"display_order"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PlaylistProblem) TableName() string { return "playlist_problems" }
