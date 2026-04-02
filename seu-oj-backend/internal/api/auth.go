package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/middleware"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			response.Error(c, "username already exists")
		case errors.Is(err, service.ErrUserIDTaken):
			response.Error(c, "userid already exists")
		default:
			response.Error(c, "register failed")
		}
		return
	}

	response.OK(c, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}

	loginResp, err := h.authService.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Error(c, "invalid username or password")
		case errors.Is(err, service.ErrUserDisabled):
			response.Error(c, "user is disabled")
		default:
			response.Error(c, "login failed")
		}
		return
	}

	response.OK(c, loginResp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	rawUserID, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Error(c, "user not found in context")
		return
	}

	userID, ok := rawUserID.(uint64)
	if !ok {
		response.Error(c, "invalid user context")
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, "user not found")
			return
		}
		response.Error(c, "query current user failed")
		return
	}

	response.OK(c, user)
}
