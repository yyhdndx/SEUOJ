package service

import (
	"errors"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/utils"
)

var (
	ErrUsernameTaken         = errors.New("username already exists")
	ErrUserIDTaken           = errors.New("userid already exists")
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUserDisabled          = errors.New("user is disabled")
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
)

type AuthService struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	username := strings.TrimSpace(req.Username)
	userID := strings.TrimSpace(req.UserID)

	var existing model.User
	err := s.db.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil, ErrUsernameTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	err = s.db.Where("userid = ?", userID).First(&existing).Error
	if err == nil {
		return nil, ErrUserIDTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username:     username,
		UserID:       userID,
		PasswordHash: string(passwordHash),
		Role:         "student",
		Status:       "active",
	}

	if err := s.db.Create(&user).Error; err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			if strings.Contains(strings.ToLower(mysqlErr.Message), "userid") {
				return nil, ErrUserIDTaken
			}
			return nil, ErrUsernameTaken
		}
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)

	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status == "disabled" {
		return nil, ErrUserDisabled
	}

	token, err := utils.GenerateToken(s.jwtSecret, user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *AuthService) GetCurrentUser(userID uint64) (*dto.UserResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *AuthService) UpdateProfile(userID uint64, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	username := strings.TrimSpace(req.Username)
	studentID := strings.TrimSpace(req.UserID)

	var duplicate model.User
	if err := s.db.Where("username = ? AND id <> ?", username, userID).First(&duplicate).Error; err == nil {
		return nil, ErrUsernameTaken
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.db.Where("userid = ? AND id <> ?", studentID, userID).First(&duplicate).Error; err == nil {
		return nil, ErrUserIDTaken
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user.Username = username
	user.UserID = studentID
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *AuthService) ChangePassword(userID uint64, req dto.ChangePasswordRequest) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCurrentPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(passwordHash)
	return s.db.Save(&user).Error
}

func toUserResponse(user model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		UserID:   user.UserID,
		Role:     user.Role,
		Status:   user.Status,
	}
}
