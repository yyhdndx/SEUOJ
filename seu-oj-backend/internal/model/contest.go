package model

import "time"

type Contest struct {
	ID               uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title            string     `gorm:"column:title;size:255;not null;index" json:"title"`
	Description      string     `gorm:"column:description;type:longtext" json:"description"`
	RuleType         string     `gorm:"column:rule_type;size:32;not null;default:acm" json:"rule_type"`
	StartTime        time.Time  `gorm:"column:start_time;not null;index" json:"start_time"`
	EndTime          time.Time  `gorm:"column:end_time;not null;index" json:"end_time"`
	IsPublic         bool       `gorm:"column:is_public;not null;default:true;index" json:"is_public"`
	AllowPractice    bool       `gorm:"column:allow_practice;not null;default:false" json:"allow_practice"`
	RanklistFreezeAt *time.Time `gorm:"column:ranklist_freeze_at" json:"ranklist_freeze_at"`
	CreatedBy        uint64     `gorm:"column:created_by;not null;index" json:"created_by"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (Contest) TableName() string {
	return "contests"
}
