package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	UserID   string `json:"userid" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserResponse struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	UserID   string `json:"userid"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
