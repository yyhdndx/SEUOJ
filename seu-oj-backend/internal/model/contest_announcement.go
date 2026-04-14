package model

import "time"

type ContestAnnouncement struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContestID uint64    `gorm:"column:contest_id;not null;index" json:"contest_id"`
	Title     string    `gorm:"column:title;size:255;not null" json:"title"`
	Content   string    `gorm:"column:content;type:mediumtext;not null" json:"content"`
	IsPinned  bool      `gorm:"column:is_pinned;not null;default:false" json:"is_pinned"`
	CreatedBy uint64    `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ContestAnnouncement) TableName() string {
	return "contest_announcements"
}
