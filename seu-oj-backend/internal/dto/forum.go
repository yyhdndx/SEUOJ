package dto

import "time"

type ForumTopicListQuery struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword   string `form:"keyword"`
	ScopeType string `form:"scope_type" binding:"omitempty,oneof=general problem contest"`
	ScopeID   uint64 `form:"scope_id"`
}

type CreateForumTopicRequest struct {
	Title     string  `json:"title" binding:"required,max=255"`
	Content   string  `json:"content" binding:"required"`
	ScopeType string  `json:"scope_type" binding:"required,oneof=general problem contest"`
	ScopeID   *uint64 `json:"scope_id"`
}

type UpdateForumTopicRequest struct {
	Title    string `json:"title" binding:"required,max=255"`
	Content  string `json:"content" binding:"required"`
	IsPinned *bool  `json:"is_pinned,omitempty"`
	IsLocked *bool  `json:"is_locked,omitempty"`
}

type CreateForumReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

type ForumTopicListItem struct {
	ID             uint64     `json:"id"`
	Title          string     `json:"title"`
	ContentPreview string     `json:"content_preview"`
	ScopeType      string     `json:"scope_type"`
	ScopeID        *uint64    `json:"scope_id"`
	AuthorID       uint64     `json:"author_id"`
	AuthorName     string     `json:"author_name"`
	ReplyCount     int        `json:"reply_count"`
	IsPinned       bool       `json:"is_pinned"`
	IsLocked       bool       `json:"is_locked"`
	LastReplyAt    *time.Time `json:"last_reply_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ForumTopicListResponse struct {
	List     []ForumTopicListItem `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type ForumReplyResponse struct {
	ID         uint64    `json:"id"`
	TopicID    uint64    `json:"topic_id"`
	Content    string    `json:"content"`
	AuthorID   uint64    `json:"author_id"`
	AuthorName string    `json:"author_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ForumTopicDetailResponse struct {
	ID          uint64               `json:"id"`
	Title       string               `json:"title"`
	Content     string               `json:"content"`
	ScopeType   string               `json:"scope_type"`
	ScopeID     *uint64              `json:"scope_id"`
	AuthorID    uint64               `json:"author_id"`
	AuthorName  string               `json:"author_name"`
	ReplyCount  int                  `json:"reply_count"`
	IsPinned    bool                 `json:"is_pinned"`
	IsLocked    bool                 `json:"is_locked"`
	LastReplyAt *time.Time           `json:"last_reply_at"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Replies     []ForumReplyResponse `json:"replies" gorm:"-"`
}
