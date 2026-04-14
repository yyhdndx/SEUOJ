package model

import "time"

type Playlist struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"column:title;size:255;not null" json:"title"`
	Description string    `gorm:"column:description;type:mediumtext" json:"description"`
	Visibility  string    `gorm:"column:visibility;type:enum('public','private','class');default:'public';not null" json:"visibility"`
	CreatedBy   uint64    `gorm:"column:created_by;not null;index" json:"created_by"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Playlist) TableName() string { return "playlists" }
