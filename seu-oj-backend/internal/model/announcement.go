package model

import "time"

type Announcement struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;index:idx_announcements_pinned_id,priority:2" json:"id"`
	Title     string    `gorm:"column:title;size:255;not null" json:"title"`
	Content   string    `gorm:"column:content;type:mediumtext;not null" json:"content"`
	IsPinned  bool      `gorm:"column:is_pinned;not null;default:false;index:idx_announcements_pinned_id,priority:1" json:"is_pinned"`
	CreatedBy uint64    `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}
