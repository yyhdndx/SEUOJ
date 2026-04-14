package dto

import "time"

type RanklistQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type RanklistItem struct {
	Rank                int        `json:"rank"`
	UserID              uint64     `json:"user_id"`
	Username            string     `json:"username"`
	StudentID           string     `json:"userid"`
	SolvedCount         int64      `json:"solved_count"`
	AcceptedSubmissions int64      `json:"accepted_submissions"`
	TotalSubmissions    int64      `json:"total_submissions"`
	LastAcceptedAt      *time.Time `json:"last_accepted_at"`
}

type RanklistResponse struct {
	List     []RanklistItem `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
