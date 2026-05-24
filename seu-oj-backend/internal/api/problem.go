package api

import (
	"errors"
	"fmt"
	"io"
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

func (h *ProblemHandler) PublicList(c *gin.Context) {
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
	var difficulty *int
	if raw := c.Query("difficulty"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 || parsed > 3 {
			response.Error(c, "invalid difficulty")
			return
		}
		difficulty = &parsed
	}
	result, err := h.problemService.ListPublicProblems(query.Page, query.PageSize, query.Keyword, difficulty)
	if err != nil {
		response.Error(c, "query public problem list failed")
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

func (h *ProblemHandler) PublicDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}

	problem, err := h.problemService.PublicProblemDetail(id)
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

func (h *ProblemHandler) ImportTestcases(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, "missing upload file")
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.Error(c, "open upload file failed")
		return
	}
	defer opened.Close()

	result, err := h.problemService.ImportProblemTestcases(id, opened, service.TestcaseImportOptions{
		Replace:  parseBoolForm(c.PostForm("replace")),
		CaseType: c.DefaultPostForm("case_type", "hidden"),
	})
	if err != nil {
		handleProblemPackageError(c, err, "import testcases failed")
		return
	}
	response.OK(c, result)
}

func (h *ProblemHandler) ExportTestcases(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	data, filename, err := h.problemService.ExportProblemTestcases(id)
	if err != nil {
		handleProblemPackageError(c, err, "export testcases failed")
		return
	}
	writeZip(c, filename, data)
}

func (h *ProblemHandler) ImportProblemPackage(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, "missing upload file")
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.Error(c, "open upload file failed")
		return
	}
	defer opened.Close()

	result, err := h.problemService.ImportProblemPackage(userID, role, opened)
	if err != nil {
		handleProblemPackageError(c, err, "import problem package failed")
		return
	}
	response.OK(c, result)
}

func (h *ProblemHandler) ExportProblemPackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	data, filename, err := h.problemService.ExportProblemPackage(id)
	if err != nil {
		handleProblemPackageError(c, err, "export problem package failed")
		return
	}
	writeZip(c, filename, data)
}

func parseBoolForm(value string) bool {
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func writeZip(c *gin.Context, filename string, data []byte) {
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Data(200, "application/zip", data)
}

func handleProblemPackageError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrPermissionDenied):
		response.Error(c, "admin permission required")
	case errors.Is(err, service.ErrProblemNotFound):
		response.Error(c, "problem not found")
	case errors.Is(err, service.ErrProblemPackageInvalid):
		response.Error(c, "invalid problem package")
	case errors.Is(err, service.ErrProblemDisplayIDConflict):
		response.Error(c, "problem display id already exists")
	case errors.Is(err, io.ErrUnexpectedEOF):
		response.Error(c, "invalid problem package")
	default:
		response.Error(c, fallback)
	}
}
