package model

import "time"

type ForumReply struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TopicID   uint64    `gorm:"column:topic_id;not null;index:idx_forum_replies_topic_id" json:"topic_id"`
	Content   string    `gorm:"column:content;type:longtext;not null" json:"content"`
	AuthorID  uint64    `gorm:"column:author_id;not null" json:"author_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ForumReply) TableName() string { return "forum_replies" }
