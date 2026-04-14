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

type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
}

func NewAnnouncementHandler(announcementService *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{announcementService: announcementService}
}

func (h *AnnouncementHandler) List(c *gin.Context) {
	var query dto.AnnouncementListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "invalid query parameters")
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	result, err := h.announcementService.List(query.Page, query.PageSize)
	if err != nil {
		response.Error(c, "query announcements failed")
		return
	}
	response.OK(c, result)
}

func (h *AnnouncementHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	item, err := h.announcementService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, service.ErrAnnouncementNotFound) {
			response.Error(c, "announcement not found")
			return
		}
		response.Error(c, "query announcement failed")
		return
	}
	response.OK(c, item)
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var req dto.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	id, err := h.announcementService.Create(userID, role, req)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "permission denied")
			return
		}
		response.Error(c, "create announcement failed")
		return
	}
	response.OK(c, gin.H{"announcement_id": id})
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	var req dto.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	if err := h.announcementService.Update(role, id, req); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrAnnouncementNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			response.Error(c, "announcement not found")
		default:
			response.Error(c, "update announcement failed")
		}
		return
	}
	response.OK(c, gin.H{"announcement_id": id})
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	if err := h.announcementService.Delete(role, id); err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "permission denied")
			return
		}
		response.Error(c, "delete announcement failed")
		return
	}
	response.OK(c, gin.H{"announcement_id": id})
}
