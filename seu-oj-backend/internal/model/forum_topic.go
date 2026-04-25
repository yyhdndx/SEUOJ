package model

import "time"

type ForumTopic struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"column:title;size:255;not null" json:"title"`
	Content     string     `gorm:"column:content;type:longtext;not null" json:"content"`
	ScopeType   string     `gorm:"column:scope_type;size:32;not null;default:general;index:idx_forum_topics_scope" json:"scope_type"`
	ScopeID     *uint64    `gorm:"column:scope_id;index:idx_forum_topics_scope" json:"scope_id"`
	AuthorID    uint64     `gorm:"column:author_id;not null" json:"author_id"`
	ReplyCount  int        `gorm:"column:reply_count;not null;default:0" json:"reply_count"`
	IsPinned    bool       `gorm:"column:is_pinned;not null;default:false" json:"is_pinned"`
	IsLocked    bool       `gorm:"column:is_locked;not null;default:false" json:"is_locked"`
	LastReplyAt *time.Time `gorm:"column:last_reply_at" json:"last_reply_at"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (ForumTopic) TableName() string { return "forum_topics" }
