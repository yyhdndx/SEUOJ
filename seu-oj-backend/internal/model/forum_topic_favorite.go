package model

import "time"

type ForumTopicFavorite struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TopicID   uint64    `gorm:"column:topic_id;not null;uniqueIndex:idx_fav_topic_user" json:"topic_id"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:idx_fav_topic_user" json:"user_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ForumTopicFavorite) TableName() string { return "forum_topic_favorites" }
