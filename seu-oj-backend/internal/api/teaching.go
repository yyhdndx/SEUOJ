package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/response"
	"seu-oj-backend/internal/service"
)

type TeachingHandler struct {
	service *service.TeachingService
}

func NewTeachingHandler(s *service.TeachingService) *TeachingHandler {
	return &TeachingHandler{service: s}
}

func (h *TeachingHandler) ListPlaylists(c *gin.Context) {
	var query dto.PlaylistListQuery
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
	result, err := h.service.ListPublicPlaylists(query.Page, query.PageSize, query.Keyword)
	if err != nil {
		response.Error(c, "query playlists failed")
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) PlaylistDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid playlist id")
		return
	}
	result, err := h.service.GetPlaylistDetail(0, "", id, false)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlaylistNotFound):
			response.Error(c, "playlist not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "playlist access denied")
		default:
			response.Error(c, "query playlist failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) MyClasses(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	items, err := h.service.ListMyClasses(userID, role)
	if err != nil {
		response.Error(c, "query classes failed")
		return
	}
	response.OK(c, gin.H{"list": items})
}

func (h *TeachingHandler) JoinClass(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	var req dto.JoinClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.JoinClass(userID, req.JoinCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrJoinCodeInvalid):
			response.Error(c, "invalid join code")
		case errors.Is(err, service.ErrAlreadyJoinedClass):
			response.Error(c, "already joined this class")
		default:
			response.Error(c, "join class failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) ClassDetail(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	result, err := h.service.GetClassDetail(userID, role, id, false)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "class access denied")
		default:
			response.Error(c, "query class detail failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) AssignmentDetail(c *gin.Context) {
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
		response.Error(c, "invalid assignment id")
		return
	}
	result, err := h.service.GetAssignmentDetail(userID, role, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssignmentNotFound):
			response.Error(c, "assignment not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "assignment access denied")
		default:
			response.Error(c, "query assignment failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) TeacherPlaylists(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var query dto.PlaylistListQuery
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
	result, err := h.service.ListTeacherPlaylists(userID, role, query.Page, query.PageSize, query.Keyword)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "teacher permission required")
			return
		}
		response.Error(c, "query teacher playlists failed")
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) TeacherPlaylistDetail(c *gin.Context) {
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
		response.Error(c, "invalid playlist id")
		return
	}
	result, err := h.service.GetPlaylistDetail(userID, role, id, true)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlaylistNotFound):
			response.Error(c, "playlist not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "playlist access denied")
		default:
			response.Error(c, "query playlist failed")
		}
		return
	}
	response.OK(c, result)
}
func (h *TeachingHandler) CreatePlaylist(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var req dto.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.CreatePlaylist(userID, role, req)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "teacher permission required")
		} else if errors.Is(err, service.ErrProblemNotFound) {
			response.Error(c, "problem not found")
		} else {
			response.Error(c, "create playlist failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) UpdatePlaylist(c *gin.Context) {
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
		response.Error(c, "invalid playlist id")
		return
	}
	var req dto.UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.UpdatePlaylist(userID, role, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "playlist access denied")
		case errors.Is(err, service.ErrProblemNotFound):
			response.Error(c, "problem not found")
		case errors.Is(err, service.ErrPlaylistNotFound):
			response.Error(c, "playlist not found")
		default:
			response.Error(c, "update playlist failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) DeletePlaylist(c *gin.Context) {
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
		response.Error(c, "invalid playlist id")
		return
	}
	if err := h.service.DeletePlaylist(userID, role, id); err != nil {
		switch {
		case errors.Is(err, service.ErrPlaylistNotFound):
			response.Error(c, "playlist not found")
		case errors.Is(err, service.ErrPlaylistInUse):
			response.Error(c, "playlist is used by assignments")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "playlist access denied")
		default:
			response.Error(c, "delete playlist failed")
		}
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
func (h *TeachingHandler) TeacherClasses(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	items, err := h.service.ListTeacherClasses(userID, role)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "teacher permission required")
		} else {
			response.Error(c, "query teacher classes failed")
		}
		return
	}
	response.OK(c, gin.H{"list": items})
}

func (h *TeachingHandler) CreateClass(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	var req dto.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.CreateClassWithUniqueCode(userID, role, req)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "teacher permission required")
		} else {
			response.Error(c, "create class failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) TeacherClassDetail(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	result, err := h.service.GetClassDetail(userID, role, id, true)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "class access denied")
		default:
			response.Error(c, "query class detail failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) TeacherClassAnalytics(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	result, err := h.service.GetTeacherClassAnalytics(userID, role, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "class access denied")
		default:
			response.Error(c, "query class analytics failed")
		}
		return
	}
	response.OK(c, result)
}
func (h *TeachingHandler) UpdateClass(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	var req dto.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.UpdateClass(userID, role, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "class access denied")
		default:
			response.Error(c, "update class failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) ClassMembers(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	result, err := h.service.ListClassMembers(userID, role, id)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "class access denied")
		} else if errors.Is(err, service.ErrClassNotFound) {
			response.Error(c, "class not found")
		} else {
			response.Error(c, "query class members failed")
		}
		return
	}
	response.OK(c, gin.H{"list": result})
}

func (h *TeachingHandler) UpdateClassMember(c *gin.Context) {
	userID, ok := getContextUserID(c)
	if !ok {
		return
	}
	role, ok := getContextRole(c)
	if !ok {
		return
	}
	classID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid class id")
		return
	}
	memberUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.Error(c, "invalid member user id")
		return
	}
	var req dto.UpdateClassMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.UpdateClassMember(userID, role, classID, memberUserID, req)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "class access denied")
		} else if errors.Is(err, service.ErrClassNotFound) {
			response.Error(c, "class not found")
		} else {
			response.Error(c, "update class member failed")
		}
		return
	}
	response.OK(c, gin.H{"list": result})
}

func (h *TeachingHandler) TeacherAssignments(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	items, err := h.service.ListTeacherAssignments(userID, role, id)
	if err != nil {
		if errors.Is(err, service.ErrTeachingForbidden) {
			response.Error(c, "class access denied")
		} else if errors.Is(err, service.ErrClassNotFound) {
			response.Error(c, "class not found")
		} else {
			response.Error(c, "query assignments failed")
		}
		return
	}
	response.OK(c, gin.H{"list": items})
}

func (h *TeachingHandler) CreateAssignment(c *gin.Context) {
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
		response.Error(c, "invalid class id")
		return
	}
	var req dto.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.CreateAssignment(userID, role, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "class access denied")
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrPlaylistNotFound):
			response.Error(c, "playlist not found")
		default:
			response.Error(c, "create assignment failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) UpdateAssignment(c *gin.Context) {
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
		response.Error(c, "invalid assignment id")
		return
	}
	var req dto.UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "invalid request parameters")
		return
	}
	result, err := h.service.UpdateAssignment(userID, role, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssignmentNotFound):
			response.Error(c, "assignment not found")
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrPlaylistNotFound):
			response.Error(c, "playlist not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "assignment access denied")
		default:
			response.Error(c, "update assignment failed")
		}
		return
	}
	response.OK(c, result)
}

func (h *TeachingHandler) DeleteAssignment(c *gin.Context) {
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
		response.Error(c, "invalid assignment id")
		return
	}
	if err := h.service.DeleteAssignment(userID, role, id); err != nil {
		switch {
		case errors.Is(err, service.ErrAssignmentNotFound):
			response.Error(c, "assignment not found")
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "assignment access denied")
		default:
			response.Error(c, "delete assignment failed")
		}
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
func (h *TeachingHandler) TeacherAssignmentOverview(c *gin.Context) {
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
		response.Error(c, "invalid assignment id")
		return
	}
	result, err := h.service.GetTeacherAssignmentOverview(userID, role, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAssignmentNotFound):
			response.Error(c, "assignment not found")
		case errors.Is(err, service.ErrClassNotFound):
			response.Error(c, "class not found")
		case errors.Is(err, service.ErrTeachingForbidden):
			response.Error(c, "assignment access denied")
		default:
			response.Error(c, "query assignment overview failed")
		}
		return
	}
	response.OK(c, result)
}
