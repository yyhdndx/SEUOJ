package model

import "time"

type Assignment struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClassID     uint64     `gorm:"column:class_id;not null;index" json:"class_id"`
	PlaylistID  uint64     `gorm:"column:playlist_id;not null;index" json:"playlist_id"`
	Title       string     `gorm:"column:title;size:255;not null" json:"title"`
	Description string     `gorm:"column:description;type:mediumtext" json:"description"`
	Type        string     `gorm:"column:type;type:enum('homework','exam');default:'homework';not null" json:"type"`
	StartAt     *time.Time `gorm:"column:start_at" json:"start_at"`
	DueAt       *time.Time `gorm:"column:due_at" json:"due_at"`
	CreatedBy   uint64     `gorm:"column:created_by;not null;index" json:"created_by"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (Assignment) TableName() string { return "assignments" }
