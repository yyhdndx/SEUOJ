package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/middleware"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type SubmissionHandler struct {
	submissionService *service.SubmissionService
}

func NewSubmissionHandler(submissionService *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{submissionService: submissionService}
}

func (h *SubmissionHandler) Create(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}

	var req dto.CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}

	submissionID, status, err := h.submissionService.CreateSubmission(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProblemUnavailable):
			response.Error(c, "problem not found or not visible")
		case errors.Is(err, service.ErrQueueEnqueueFailed):
			response.Error(c, "submission created but enqueue failed")
		default:
			response.Error(c, "create submission failed")
		}
		return
	}

	response.OK(c, dto.CreateSubmissionResponse{SubmissionID: submissionID, Status: status})
}

func (h *SubmissionHandler) Run(c *gin.Context) {
	var req dto.RunSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}

	result, err := h.submissionService.RunSampleTests(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProblemUnavailable):
			response.Error(c, "problem not found or not visible")
		default:
			response.Error(c, "run sample tests failed")
		}
		return
	}

	response.OK(c, result)
}

func (h *SubmissionHandler) ListMy(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}

	var query dto.SubmissionListQuery
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

	var problemID *uint64
	if query.ProblemID != 0 {
		problemID = &query.ProblemID
	}

	var status *string
	if query.Status != "" {
		status = &query.Status
	}

	result, err := h.submissionService.ListMySubmissions(userID, query.Page, query.PageSize, problemID, status)
	if err != nil {
		response.Error(c, "query submission list failed")
		return
	}

	response.OK(c, result)
}

func (h *SubmissionHandler) Detail(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}

	submissionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid submission id")
		return
	}

	result, err := h.submissionService.GetSubmissionDetail(userID, role, submissionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubmissionNotFound):
			response.Error(c, "submission not found")
		case errors.Is(err, service.ErrSubmissionForbidden):
			response.Error(c, "submission access denied")
		default:
			response.Error(c, "query submission detail failed")
		}
		return
	}

	response.OK(c, result)
}

func getContextUserID(c *gin.Context) (uint64, bool) {
	rawUserID, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Error(c, "missing user id")
		return 0, false
	}

	userID, ok := rawUserID.(uint64)
	if !ok {
		response.Error(c, "invalid user context")
		return 0, false
	}

	return userID, true
}

func getContextRole(c *gin.Context) (string, bool) {
	rawRole, exists := c.Get(middleware.ContextRoleKey)
	if !exists {
		response.Error(c, "missing user role")
		return "", false
	}

	role, ok := rawRole.(string)
	if !ok {
		response.Error(c, "invalid user role")
		return "", false
	}

	return role, true
}
