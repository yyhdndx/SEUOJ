package api

import (
	"errors"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	result, err := h.statsService.Overview()
	if err != nil {
		response.Error(c, "query overview stats failed")
		return
	}
	response.OK(c, result)
}

func (h *StatsHandler) Me(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}

	result, err := h.statsService.My(userID)
	if err != nil {
		response.Error(c, "query user stats failed")
		return
	}
	response.OK(c, result)
}

func (h *StatsHandler) Admin(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}

	result, err := h.statsService.Admin(role)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "permission denied")
			return
		}
		response.Error(c, "query admin stats failed")
		return
	}
	response.OK(c, result)
}
