package dto

import "time"

type AnnouncementListQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type CreateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required,max=255"`
	Content  string `json:"content" binding:"required"`
	IsPinned bool   `json:"is_pinned"`
}

type AnnouncementItem struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsPinned  bool      `json:"is_pinned"`
	CreatedBy uint64    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AnnouncementListResponse struct {
	List     []AnnouncementItem `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}
