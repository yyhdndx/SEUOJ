package model

import "time"

type ContestRegistration struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContestID uint64    `gorm:"column:contest_id;not null;index;uniqueIndex:uk_contest_user,priority:1" json:"contest_id"`
	UserID    uint64    `gorm:"column:user_id;not null;index;uniqueIndex:uk_contest_user,priority:2" json:"user_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ContestRegistration) TableName() string {
	return "contest_registrations"
}
