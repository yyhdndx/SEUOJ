package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List(role string, page, pageSize int, keyword, targetRole, status string) (*dto.AdminUserListResponse, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}

	var users []model.User
	var total int64
	query := s.db.Model(&model.User{})
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		query = query.Where("username LIKE ? OR userid LIKE ?", "%"+trimmed+"%", "%"+trimmed+"%")
	}
	if targetRole != "" {
		query = query.Where("role = ?", targetRole)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, err
	}
	list := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		list = append(list, dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			UserID:   user.UserID,
			Role:     user.Role,
			Status:   user.Status,
		})
	}
	return &dto.AdminUserListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *UserService) Update(requestRole string, id uint64, req dto.AdminUserUpdateRequest) (*dto.UserResponse, error) {
	if requestRole != "admin" {
		return nil, ErrPermissionDenied
	}

	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	username := strings.TrimSpace(req.Username)
	studentID := strings.TrimSpace(req.UserID)

	var duplicate model.User
	if err := s.db.Where("username = ? AND id <> ?", username, id).First(&duplicate).Error; err == nil {
		return nil, ErrUsernameTaken
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := s.db.Where("userid = ? AND id <> ?", studentID, id).First(&duplicate).Error; err == nil {
		return nil, ErrUserIDTaken
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user.Username = username
	user.UserID = studentID
	user.Role = req.Role
	user.Status = req.Status
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	resp := dto.UserResponse{ID: user.ID, Username: user.Username, UserID: user.UserID, Role: user.Role, Status: user.Status}
	return &resp, nil
}
