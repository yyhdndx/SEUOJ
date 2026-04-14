package api

import (
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RanklistHandler struct {
	ranklistService *service.RanklistService
}

func NewRanklistHandler(ranklistService *service.RanklistService) *RanklistHandler {
	return &RanklistHandler{ranklistService: ranklistService}
}

func (h *RanklistHandler) List(c *gin.Context) {
	var query dto.RanklistQuery
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
	result, err := h.ranklistService.List(query.Page, query.PageSize)
	if err != nil {
		response.Error(c, "query ranklist failed")
		return
	}
	response.OK(c, result)
}
