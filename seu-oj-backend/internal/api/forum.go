package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type ForumHandler struct{ service *service.ForumService }

func NewForumHandler(s *service.ForumService) *ForumHandler { return &ForumHandler{service: s} }

func (h *ForumHandler) ListTopics(c *gin.Context) {
	var query dto.ForumTopicListQuery
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
	var scopeID *uint64
	if query.ScopeID != 0 {
		scopeID = &query.ScopeID
	}
	optionalUserID := getOptionalUserID(c)
	result, err := h.service.ListTopics(query.Page, query.PageSize, query.Keyword, query.ScopeType, scopeID, optionalUserID)
	if err != nil {
		response.Error(c, "query forum topics failed")
		return
	}
	response.OK(c, result)
}

func (h *ForumHandler) TopicDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	optionalUserID := getOptionalUserID(c)
	result, err := h.service.GetTopicDetail(id, optionalUserID)
	if err != nil {
		if errors.Is(err, service.ErrForumTopicNotFound) {
			response.Error(c, "topic not found")
		} else {
			response.Error(c, "query topic detail failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ForumHandler) CreateTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	var req dto.CreateForumTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.CreateTopic(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProblemNotFound):
			response.Error(c, "problem not found")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "create topic failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ForumHandler) UpdateTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	var req dto.UpdateForumTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.UpdateTopic(userID, role, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForumTopicNotFound):
			response.Error(c, "topic not found")
		case errors.Is(err, service.ErrForumForbidden):
			response.Error(c, "topic access denied")
		default:
			response.Error(c, "update topic failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ForumHandler) DeleteTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	if err := h.service.DeleteTopic(userID, role, id); err != nil {
		switch {
		case errors.Is(err, service.ErrForumTopicNotFound):
			response.Error(c, "topic not found")
		case errors.Is(err, service.ErrForumForbidden):
			response.Error(c, "topic access denied")
		default:
			response.Error(c, "delete topic failed")
		}
		return
	}
	response.OK(c, gin.H{"topic_id": id})
}

func (h *ForumHandler) CreateReply(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	var req dto.CreateForumReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.CreateReply(userID, role, topicID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForumTopicNotFound):
			response.Error(c, "topic not found")
		case errors.Is(err, service.ErrForumLocked):
			response.Error(c, "topic is locked")
		default:
			response.Error(c, "create reply failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ForumHandler) UpdateReply(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	replyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid reply id")
		return
	}
	var req dto.CreateForumReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.UpdateReply(userID, role, replyID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForumReplyNotFound):
			response.Error(c, "reply not found")
		case errors.Is(err, service.ErrForumForbidden):
			response.Error(c, "reply access denied")
		default:
			response.Error(c, "update reply failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ForumHandler) DeleteReply(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	replyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid reply id")
		return
	}
	if err := h.service.DeleteReply(userID, role, replyID); err != nil {
		switch {
		case errors.Is(err, service.ErrForumReplyNotFound):
			response.Error(c, "reply not found")
		case errors.Is(err, service.ErrForumForbidden):
			response.Error(c, "reply access denied")
		default:
			response.Error(c, "delete reply failed")
		}
		return
	}
	response.OK(c, gin.H{"reply_id": replyID})
}

func (h *ForumHandler) LikeTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	if err := h.service.LikeTopic(c.Request.Context(), topicID, userID); err != nil {
		if errors.Is(err, service.ErrForumAlreadyLiked) {
			response.Error(c, "already liked")
		} else {
			response.Error(c, "like topic failed")
		}
		return
	}
	response.OK(c, gin.H{"liked": true})
}

func (h *ForumHandler) UnlikeTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	if err := h.service.UnlikeTopic(c.Request.Context(), topicID, userID); err != nil {
		response.Error(c, "unlike topic failed")
		return
	}
	response.OK(c, gin.H{"liked": false})
}

func (h *ForumHandler) FavoriteTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	if err := h.service.FavoriteTopic(c.Request.Context(), topicID, userID); err != nil {
		if errors.Is(err, service.ErrForumAlreadyLiked) {
			response.Error(c, "already favorited")
		} else {
			response.Error(c, "favorite topic failed")
		}
		return
	}
	response.OK(c, gin.H{"favorited": true})
}

func (h *ForumHandler) UnfavoriteTopic(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid topic id")
		return
	}
	if err := h.service.UnfavoriteTopic(c.Request.Context(), topicID, userID); err != nil {
		response.Error(c, "unfavorite topic failed")
		return
	}
	response.OK(c, gin.H{"favorited": false})
}
