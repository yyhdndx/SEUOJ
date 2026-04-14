package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) AdminList(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var query dto.AdminUserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "invalid query parameters")
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	result, err := h.userService.List(role, query.Page, query.PageSize, query.Keyword, query.Role, query.Status)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "permission denied")
			return
		}
		response.Error(c, "query user list failed")
		return
	}
	response.OK(c, result)
}

func (h *UserHandler) AdminUpdate(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid user id")
		return
	}
	var req dto.AdminUserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	user, err := h.userService.Update(role, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrUsernameTaken):
			response.Error(c, "username already exists")
		case errors.Is(err, service.ErrUserIDTaken):
			response.Error(c, "userid already exists")
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.Error(c, "user not found")
		default:
			response.Error(c, "update user failed")
		}
		return
	}
	response.OK(c, user)
}
