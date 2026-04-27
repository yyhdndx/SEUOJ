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

type ProblemHandler struct {
	problemService *service.ProblemService
}

func NewProblemHandler(problemService *service.ProblemService) *ProblemHandler {
	return &ProblemHandler{problemService: problemService}
}

func (h *ProblemHandler) Create(c *gin.Context) {
	var req dto.CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}

	rawUserID, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Error(c, "missing user id")
		return
	}
	userID, ok := rawUserID.(uint64)
	if !ok {
		response.Error(c, "invalid user context")
		return
	}

	rawRole, exists := c.Get(middleware.ContextRoleKey)
	if !exists {
		response.Error(c, "missing user role")
		return
	}
	role, ok := rawRole.(string)
	if !ok {
		response.Error(c, "invalid user role")
		return
	}

	problemID, err := h.problemService.CreateProblem(userID, role, req)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "admin permission required")
			return
		}
		response.Error(c, "create problem failed")
		return
	}

	response.OK(c, dto.CreateProblemResponse{ProblemID: problemID})
}

func (h *ProblemHandler) AdminList(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}

	var query dto.AdminProblemListQuery
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

	result, err := h.problemService.ListAdminProblems(role, query.Page, query.PageSize, query.Keyword, query.IncludeHidden)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "admin permission required")
			return
		}
		response.Error(c, "query admin problem list failed")
		return
	}

	response.OK(c, result)
}

func (h *ProblemHandler) List(c *gin.Context) {
	var query dto.ProblemListQuery
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

	result, err := h.problemService.ListProblems(c.Request.Context(), query.Page, query.PageSize, query.Keyword)
	if err != nil {
		response.Error(c, "query problem list failed")
		return
	}

	response.OK(c, result)
}

func (h *ProblemHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}

	problem, err := h.problemService.GetProblemDetail(id)
	if err != nil {
		if errors.Is(err, service.ErrProblemNotFound) {
			response.Error(c, "problem not found")
			return
		}
		response.Error(c, "query problem detail failed")
		return
	}

	response.OK(c, problem)
}

func (h *ProblemHandler) Stats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}

	stats, err := h.problemService.GetProblemStats(id)
	if err != nil {
		if errors.Is(err, service.ErrProblemNotFound) {
			response.Error(c, "problem not found")
			return
		}
		response.Error(c, "query problem stats failed")
		return
	}

	response.OK(c, stats)
}

func (h *ProblemHandler) SolutionList(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	items, err := h.problemService.ListProblemSolutions(problemID, false)
	if err != nil {
		if errors.Is(err, service.ErrProblemNotFound) {
			response.Error(c, "problem not found")
			return
		}
		response.Error(c, "query problem solutions failed")
		return
	}
	response.OK(c, gin.H{"list": items})
}

func (h *ProblemHandler) TeacherSolutionList(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	items, err := h.problemService.ListManageProblemSolutions(userID, role, problemID)
	if err != nil {
		if errors.Is(err, service.ErrProblemNotFound) {
			response.Error(c, "problem not found")
			return
		}
		response.Error(c, "query problem solutions failed")
		return
	}
	response.OK(c, gin.H{"list": items})
}

func (h *ProblemHandler) CreateSolution(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	var req dto.CreateProblemSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.problemService.CreateProblemSolution(userID, role, problemID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "solution access denied")
		case errors.Is(err, service.ErrSolutionPublishNotAllowed):
			response.Error(c, "accepted submission required before publishing a solution")
		case errors.Is(err, service.ErrProblemNotFound):
			response.Error(c, "problem not found")
		case errors.Is(err, service.ErrInvalidProblemSolutionData):
			response.Error(c, "invalid request parameters")
		default:
			response.Error(c, "create solution failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ProblemHandler) UpdateSolution(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	solutionID, err := strconv.ParseUint(c.Param("solution_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid solution id")
		return
	}
	var req dto.CreateProblemSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.problemService.UpdateProblemSolution(userID, role, problemID, solutionID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "solution access denied")
		case errors.Is(err, service.ErrSolutionNotFound):
			response.Error(c, "solution not found")
		case errors.Is(err, service.ErrInvalidProblemSolutionData):
			response.Error(c, "invalid request parameters")
		default:
			response.Error(c, "update solution failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ProblemHandler) DeleteSolution(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	solutionID, err := strconv.ParseUint(c.Param("solution_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid solution id")
		return
	}
	if err := h.problemService.DeleteProblemSolution(userID, role, problemID, solutionID); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "solution access denied")
		case errors.Is(err, service.ErrSolutionNotFound):
			response.Error(c, "solution not found")
		default:
			response.Error(c, "delete solution failed")
		}
		return
	}
	response.OK(c, gin.H{"problem_id": problemID, "solution_id": solutionID})
}

func (h *ProblemHandler) AdminDetail(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}

	problem, err := h.problemService.GetAdminProblemDetail(role, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "admin permission required")
		case errors.Is(err, service.ErrProblemNotFound):
			response.Error(c, "problem not found")
		default:
			response.Error(c, "query admin problem detail failed")
		}
		return
	}

	response.OK(c, problem)
}

func (h *ProblemHandler) Update(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}

	var req dto.CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}

	if err := h.problemService.UpdateProblem(role, id, req); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "admin permission required")
		case errors.Is(err, service.ErrProblemNotFound):
			response.Error(c, "problem not found")
		default:
			response.Error(c, "update problem failed")
		}
		return
	}

	response.OK(c, dto.CreateProblemResponse{ProblemID: id})
}

func (h *ProblemHandler) Delete(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}

	if err := h.problemService.DeleteProblem(role, id); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "admin permission required")
		case errors.Is(err, service.ErrProblemNotFound):
			response.Error(c, "problem not found")
		default:
			response.Error(c, "delete problem failed")
		}
		return
	}

	response.OK(c, gin.H{"problem_id": id})
}
