package service

import (
	"testing"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceRegisterLoginAndChangePassword(t *testing.T) {
	db := openServiceTestDB(t)
	svc := NewAuthService(db, "auth-test-secret")

	user, err := svc.Register(dto.RegisterRequest{
		Username: "authflow", UserID: "AF001", Password: "OldPass123!",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Username != "authflow" {
		t.Fatalf("unexpected user: %+v", user)
	}

	login, err := svc.Login(dto.LoginRequest{Username: "authflow", Password: "OldPass123!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.Token == "" {
		t.Fatal("expected login token")
	}

	current, err := svc.GetCurrentUser(user.ID)
	if err != nil || current.Username != "authflow" {
		t.Fatalf("get current user: err=%v user=%+v", err, current)
	}

	updated, err := svc.UpdateProfile(user.ID, dto.UpdateProfileRequest{
		Username: "authflow2", UserID: "AF002",
	})
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.Username != "authflow2" || updated.UserID != "AF002" {
		t.Fatalf("unexpected profile update: %+v", updated)
	}

	if err := svc.ChangePassword(user.ID, dto.ChangePasswordRequest{
		CurrentPassword: "OldPass123!", NewPassword: "NewPass456!",
	}); err != nil {
		t.Fatalf("change password: %v", err)
	}

	if _, err := svc.Login(dto.LoginRequest{Username: "authflow2", Password: "NewPass456!"}); err != nil {
		t.Fatalf("login after password change: %v", err)
	}
	if err := svc.ChangePassword(user.ID, dto.ChangePasswordRequest{
		CurrentPassword: "wrong", NewPassword: "Another1!",
	}); err != ErrInvalidCurrentPassword {
		t.Fatalf("expected invalid current password, got %v", err)
	}
}

func TestAuthServiceRegisterDuplicateUserID(t *testing.T) {
	db := openServiceTestDB(t)
	svc := NewAuthService(db, "dup-test")
	if _, err := svc.Register(dto.RegisterRequest{Username: "u1", UserID: "ID1", Password: "pass"}); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if _, err := svc.Register(dto.RegisterRequest{Username: "u2", UserID: "ID1", Password: "pass"}); err != ErrUserIDTaken {
		t.Fatalf("expected userid taken, got %v", err)
	}
}

func TestAuthServiceRegisterDuplicateUsername(t *testing.T) {
	db := openServiceTestDB(t)
	svc := NewAuthService(db, "dup-test")
	if _, err := svc.Register(dto.RegisterRequest{Username: "same", UserID: "ID1", Password: "pass"}); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if _, err := svc.Register(dto.RegisterRequest{Username: "same", UserID: "ID2", Password: "pass"}); err != ErrUsernameTaken {
		t.Fatalf("expected username taken, got %v", err)
	}
}

func TestAuthServiceRejectsDisabledUser(t *testing.T) {
	db := openServiceTestDB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	user := model.User{Username: "disabled", UserID: "DIS01", PasswordHash: string(hash), Role: "student", Status: "disabled"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := svcLogin(NewAuthService(db, "s"), "disabled", "pass"); err != ErrUserDisabled {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func svcLogin(s *AuthService, username, password string) (*dto.LoginResponse, error) {
	return s.Login(dto.LoginRequest{Username: username, Password: password})
}
