package dto

import "time"

type ContestAnnouncementListQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type CreateContestAnnouncementRequest struct {
	Title    string `json:"title" binding:"required,max=255"`
	Content  string `json:"content" binding:"required"`
	IsPinned bool   `json:"is_pinned"`
}

type ContestAnnouncementItem struct {
	ID        uint64    `json:"id"`
	ContestID uint64    `json:"contest_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsPinned  bool      `json:"is_pinned"`
	CreatedBy uint64    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ContestAnnouncementListResponse struct {
	List     []ContestAnnouncementItem `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}
