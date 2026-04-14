package dto

type AdminUserListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
	Status   string `form:"status"`
}

type AdminUserUpdateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	UserID   string `json:"userid" binding:"required,min=3,max=50"`
	Role     string `json:"role" binding:"required,oneof=student admin teacher"`
	Status   string `json:"status" binding:"required,oneof=active disabled"`
}

type AdminUserListResponse struct {
	List     []UserResponse `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
