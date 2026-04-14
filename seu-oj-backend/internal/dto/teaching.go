package dto

import "time"

type PlaylistProblemRequest struct {
	ProblemID    uint64 `json:"problem_id" binding:"required"`
	DisplayOrder int    `json:"display_order" binding:"required,min=1"`
}

type CreatePlaylistRequest struct {
	Title       string                   `json:"title" binding:"required,min=2,max=255"`
	Description string                   `json:"description"`
	Visibility  string                   `json:"visibility" binding:"required,oneof=public private class"`
	Problems    []PlaylistProblemRequest `json:"problems" binding:"required,min=1,dive"`
}

type UpdatePlaylistRequest struct {
	Title       string                   `json:"title" binding:"required,min=2,max=255"`
	Description string                   `json:"description"`
	Visibility  string                   `json:"visibility" binding:"required,oneof=public private class"`
	Problems    []PlaylistProblemRequest `json:"problems" binding:"required,min=1,dive"`
}

type PlaylistListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
}

type PlaylistProblemItem struct {
	ProblemID    uint64 `json:"problem_id"`
	DisplayOrder int    `json:"display_order"`
	DisplayID    string `json:"display_id"`
	Title        string `json:"title"`
}

type PlaylistListItem struct {
	ID           uint64    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Visibility   string    `json:"visibility"`
	CreatedBy    uint64    `json:"created_by"`
	ProblemCount int       `json:"problem_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type PlaylistDetailResponse struct {
	ID          uint64                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Visibility  string                `json:"visibility"`
	CreatedBy   uint64                `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Problems    []PlaylistProblemItem `json:"problems"`
}

type PlaylistListResponse struct {
	List     []PlaylistListItem `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type CreateClassRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description"`
}

type UpdateClassRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=active archived"`
}

type JoinClassRequest struct {
	JoinCode string `json:"join_code" binding:"required,min=4,max=32"`
}

type UpdateClassMemberRequest struct {
	Role   string `json:"role" binding:"required,oneof=student assistant"`
	Status string `json:"status" binding:"required,oneof=active removed"`
}

type CreateAssignmentRequest struct {
	PlaylistID  uint64     `json:"playlist_id" binding:"required"`
	Title       string     `json:"title" binding:"required,min=2,max=255"`
	Description string     `json:"description"`
	Type        string     `json:"type" binding:"required,oneof=homework exam"`
	StartAt     *time.Time `json:"start_at"`
	DueAt       *time.Time `json:"due_at"`
}

type UpdateAssignmentRequest struct {
	PlaylistID  uint64     `json:"playlist_id" binding:"required"`
	Title       string     `json:"title" binding:"required,min=2,max=255"`
	Description string     `json:"description"`
	Type        string     `json:"type" binding:"required,oneof=homework exam"`
	StartAt     *time.Time `json:"start_at"`
	DueAt       *time.Time `json:"due_at"`
}

type ClassroomListItem struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	JoinCode        string    `json:"join_code,omitempty"`
	TeacherID       uint64    `json:"teacher_id"`
	TeacherName     string    `json:"teacher_name"`
	Status          string    `json:"status"`
	MemberRole      string    `json:"member_role,omitempty"`
	MemberCount     int64     `json:"member_count"`
	AssignmentCount int64     `json:"assignment_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type ClassMemberItem struct {
	UserID    uint64    `json:"user_id"`
	Username  string    `json:"username"`
	StudentID string    `json:"userid"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	JoinedAt  time.Time `json:"joined_at"`
}

type AssignmentListItem struct {
	ID          uint64     `json:"id"`
	ClassID     uint64     `json:"class_id"`
	PlaylistID  uint64     `json:"playlist_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	StartAt     *time.Time `json:"start_at"`
	DueAt       *time.Time `json:"due_at"`
	Status      string     `json:"status"`
	SolvedCount int        `json:"solved_count"`
	TotalCount  int        `json:"total_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ClassroomDetailResponse struct {
	ID          uint64               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	JoinCode    string               `json:"join_code,omitempty"`
	TeacherID   uint64               `json:"teacher_id"`
	TeacherName string               `json:"teacher_name"`
	Status      string               `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Members     []ClassMemberItem    `json:"members,omitempty"`
	Assignments []AssignmentListItem `json:"assignments"`
}

type AssignmentProblemItem struct {
	ProblemID    uint64 `json:"problem_id"`
	DisplayID    string `json:"display_id"`
	Title        string `json:"title"`
	DisplayOrder int    `json:"display_order"`
	Solved       bool   `json:"solved"`
	LastStatus   string `json:"last_status,omitempty"`
}

type AssignmentDetailResponse struct {
	ID            uint64                  `json:"id"`
	ClassID       uint64                  `json:"class_id"`
	ClassName     string                  `json:"class_name"`
	PlaylistID    uint64                  `json:"playlist_id"`
	PlaylistTitle string                  `json:"playlist_title"`
	Title         string                  `json:"title"`
	Description   string                  `json:"description"`
	Type          string                  `json:"type"`
	StartAt       *time.Time              `json:"start_at"`
	DueAt         *time.Time              `json:"due_at"`
	Status        string                  `json:"status"`
	SolvedCount   int                     `json:"solved_count"`
	TotalCount    int                     `json:"total_count"`
	Problems      []AssignmentProblemItem `json:"problems"`
}

type AssignmentStudentProblemStatus struct {
	ProblemID    uint64 `json:"problem_id"`
	DisplayID    string `json:"display_id"`
	Title        string `json:"title"`
	DisplayOrder int    `json:"display_order"`
	Solved       bool   `json:"solved"`
	LastStatus   string `json:"last_status,omitempty"`
}

type AssignmentStudentProgressItem struct {
	UserID           uint64                           `json:"user_id"`
	Username         string                           `json:"username"`
	StudentID        string                           `json:"userid"`
	Role             string                           `json:"role"`
	SolvedCount      int                              `json:"solved_count"`
	TotalCount       int                              `json:"total_count"`
	ProgressStatus   string                           `json:"progress_status"`
	LastSubmissionAt *time.Time                       `json:"last_submission_at"`
	ProblemStatuses  []AssignmentStudentProblemStatus `json:"problem_statuses"`
}

type ClassAnalyticsTopStudentItem struct {
	UserID               uint64     `json:"user_id"`
	Username             string     `json:"username"`
	StudentID            string     `json:"userid"`
	Role                 string     `json:"role"`
	SolvedProblemCount   int        `json:"solved_problem_count"`
	CompletedAssignments int        `json:"completed_assignments"`
	LastSubmissionAt     *time.Time `json:"last_submission_at"`
}

type ClassAnalyticsAssignmentItem struct {
	AssignmentID   uint64 `json:"assignment_id"`
	Title          string `json:"title"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	ProblemCount   int    `json:"problem_count"`
	StartedCount   int    `json:"started_count"`
	CompletedCount int    `json:"completed_count"`
	CompletionRate int    `json:"completion_rate"`
}

type ClassAnalyticsResponse struct {
	ClassID               uint64                         `json:"class_id"`
	ClassName             string                         `json:"class_name"`
	TeacherName           string                         `json:"teacher_name"`
	MemberCount           int                            `json:"member_count"`
	AssignmentCount       int                            `json:"assignment_count"`
	UniqueProblemCount    int                            `json:"unique_problem_count"`
	OverallCompletionRate int                            `json:"overall_completion_rate"`
	Assignments           []ClassAnalyticsAssignmentItem `json:"assignments"`
	TopStudents           []ClassAnalyticsTopStudentItem `json:"top_students"`
}

type AssignmentOverviewResponse struct {
	ID             uint64                          `json:"id"`
	ClassID        uint64                          `json:"class_id"`
	ClassName      string                          `json:"class_name"`
	PlaylistID     uint64                          `json:"playlist_id"`
	PlaylistTitle  string                          `json:"playlist_title"`
	Title          string                          `json:"title"`
	Description    string                          `json:"description"`
	Type           string                          `json:"type"`
	StartAt        *time.Time                      `json:"start_at"`
	DueAt          *time.Time                      `json:"due_at"`
	Status         string                          `json:"status"`
	MemberCount    int                             `json:"member_count"`
	CompletedCount int                             `json:"completed_count"`
	StartedCount   int                             `json:"started_count"`
	ProblemCount   int                             `json:"problem_count"`
	Members        []AssignmentStudentProgressItem `json:"members"`
}
