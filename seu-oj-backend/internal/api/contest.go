package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type ContestHandler struct {
	contestService *service.ContestService
}

func NewContestHandler(contestService *service.ContestService) *ContestHandler {
	return &ContestHandler{contestService: contestService}
}

func (h *ContestHandler) List(c *gin.Context) {
	var query dto.ContestListQuery
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
	result, err := h.contestService.List(query.Page, query.PageSize, query.Keyword, query.Status)
	if err != nil {
		response.Error(c, "query contest list failed")
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) Detail(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	result, err := h.contestService.GetByID(contestID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "query contest detail failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) Me(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	result, err := h.contestService.GetMe(userID, role, contestID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestForbidden):
			response.Error(c, "contest access denied")
		default:
			response.Error(c, "query contest membership failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) Register(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	if err := h.contestService.Register(userID, role, contestID); err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestForbidden):
			response.Error(c, "contest access denied")
		case errors.Is(err, service.ErrContestRegistrationClosed):
			response.Error(c, "contest registration closed")
		default:
			response.Error(c, "register contest failed")
		}
		return
	}
	response.OK(c, dto.ContestRegisterResponse{ContestID: contestID, Registered: true})
}

func (h *ContestHandler) Problems(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	result, err := h.contestService.ListProblems(userID, role, contestID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestForbidden), errors.Is(err, service.ErrContestNotRegistered):
			response.Error(c, "contest problems unavailable")
		default:
			response.Error(c, "query contest problems failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) ProblemDetail(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	problemID, err := strconv.ParseUint(c.Param("problem_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid problem id")
		return
	}
	result, err := h.contestService.GetProblemDetail(userID, role, contestID, problemID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestProblemNotFound):
			response.Error(c, "contest problem not found")
		case errors.Is(err, service.ErrContestForbidden), errors.Is(err, service.ErrContestNotRegistered):
			response.Error(c, "contest problem unavailable")
		default:
			response.Error(c, "query contest problem failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) Ranklist(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	result, err := h.contestService.GetRanklist(contestID)
	if err != nil {
		if errors.Is(err, service.ErrContestNotFound) {
			response.Error(c, "contest not found")
		} else {
			response.Error(c, "query contest ranklist failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) AdminRanklist(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	result, err := h.contestService.GetAdminRanklist(role, contestID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "query admin contest ranklist failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) AnnouncementList(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	var query dto.ContestAnnouncementListQuery
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
	result, err := h.contestService.ListAnnouncements(contestID, query.Page, query.PageSize)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "query contest announcements failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) AnnouncementDetail(c *gin.Context) {
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	announcementID, err := strconv.ParseUint(c.Param("announcement_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	result, err := h.contestService.GetAnnouncement(contestID, announcementID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestAnnouncementNotFound):
			response.Error(c, "contest announcement not found")
		default:
			response.Error(c, "query contest announcement failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) AdminAnnouncementList(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	var query dto.ContestAnnouncementListQuery
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
	result, err := h.contestService.ListAdminAnnouncements(role, contestID, query.Page, query.PageSize)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "query admin contest announcements failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) AdminAnnouncementDetail(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	announcementID, err := strconv.ParseUint(c.Param("announcement_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	result, err := h.contestService.GetAdminAnnouncement(role, contestID, announcementID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestAnnouncementNotFound):
			response.Error(c, "contest announcement not found")
		default:
			response.Error(c, "query admin contest announcement failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) CreateAnnouncement(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	var req dto.CreateContestAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	announcementID, err := h.contestService.CreateAnnouncement(userID, role, contestID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "create contest announcement failed")
		}
		return
	}
	response.OK(c, gin.H{"announcement_id": announcementID, "contest_id": contestID})
}

func (h *ContestHandler) UpdateAnnouncement(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	announcementID, err := strconv.ParseUint(c.Param("announcement_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	var req dto.CreateContestAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	if err := h.contestService.UpdateAnnouncement(role, contestID, announcementID, req); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestAnnouncementNotFound):
			response.Error(c, "contest announcement not found")
		default:
			response.Error(c, "update contest announcement failed")
		}
		return
	}
	response.OK(c, gin.H{"announcement_id": announcementID, "contest_id": contestID})
}

func (h *ContestHandler) DeleteAnnouncement(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	announcementID, err := strconv.ParseUint(c.Param("announcement_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid announcement id")
		return
	}
	if err := h.contestService.DeleteAnnouncement(role, contestID, announcementID); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestAnnouncementNotFound):
			response.Error(c, "contest announcement not found")
		default:
			response.Error(c, "delete contest announcement failed")
		}
		return
	}
	response.OK(c, gin.H{"announcement_id": announcementID, "contest_id": contestID})
}

func (h *ContestHandler) AdminList(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var query dto.ContestListQuery
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
	result, err := h.contestService.ListAdmin(role, query.Page, query.PageSize, query.Keyword, query.Status)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			response.Error(c, "permission denied")
		} else {
			response.Error(c, "query admin contest list failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) AdminDetail(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	result, err := h.contestService.GetAdminByID(role, contestID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "query admin contest detail failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *ContestHandler) Create(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var req dto.CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	contestID, err := h.contestService.Create(userID, role, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestInvalidTimeRange), errors.Is(err, service.ErrContestInvalidProblemSet):
			response.Error(c, err.Error())
		default:
			response.Error(c, "create contest failed")
		}
		return
	}
	response.OK(c, dto.CreateContestResponse{ContestID: contestID})
}

func (h *ContestHandler) Update(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	var req dto.CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	if err := h.contestService.Update(role, contestID, req); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		case errors.Is(err, service.ErrContestInvalidTimeRange), errors.Is(err, service.ErrContestInvalidProblemSet):
			response.Error(c, err.Error())
		default:
			response.Error(c, "update contest failed")
		}
		return
	}
	response.OK(c, dto.CreateContestResponse{ContestID: contestID})
}

func (h *ContestHandler) Delete(c *gin.Context) {
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	contestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid contest id")
		return
	}
	if err := h.contestService.Delete(role, contestID); err != nil {
		switch {
		case errors.Is(err, service.ErrPermissionDenied):
			response.Error(c, "permission denied")
		case errors.Is(err, service.ErrContestNotFound):
			response.Error(c, "contest not found")
		default:
			response.Error(c, "delete contest failed")
		}
		return
	}
	response.OK(c, gin.H{"contest_id": contestID})
}
