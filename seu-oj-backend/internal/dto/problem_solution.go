package dto

import "time"

type CreateProblemSolutionRequest struct {
	Title      string `json:"title" binding:"required,max=255"`
	Content    string `json:"content" binding:"required"`
	Visibility string `json:"visibility" binding:"required,oneof=public private"`
}

type ProblemSolutionResponse struct {
	ID         uint64    `json:"id"`
	ProblemID  uint64    `json:"problem_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Visibility string    `json:"visibility"`
	AuthorID   uint64    `json:"author_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
