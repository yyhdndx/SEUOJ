package model

import "time"

type AuditLog struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RequestID  string    `gorm:"column:request_id;size:64;index" json:"request_id"`
	ActorID    *uint64   `gorm:"column:actor_id;index" json:"actor_id"`
	ActorRole  string    `gorm:"column:actor_role;size:32;index" json:"actor_role"`
	Action     string    `gorm:"column:action;size:128;index" json:"action"`
	Method     string    `gorm:"column:method;size:16;not null" json:"method"`
	Path       string    `gorm:"column:path;size:512;not null" json:"path"`
	Route      string    `gorm:"column:route;size:512" json:"route"`
	StatusCode int       `gorm:"column:status_code;not null" json:"status_code"`
	ClientIP   string    `gorm:"column:client_ip;size:64" json:"client_ip"`
	UserAgent  string    `gorm:"column:user_agent;size:255" json:"user_agent"`
	ErrorText  string    `gorm:"column:error_text;type:text" json:"error_text"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
