package service

import (
	"context"
	"errors"
	"testing"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

func TestForumServiceScopeAndLockErrors(t *testing.T) {
	db := openServiceTestDB(t)
	svc := NewForumService(db, cache.New(nil))
	user := model.User{Username: "forumu", UserID: "FOR01", PasswordHash: "h", Role: "student", Status: "active"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := svc.CreateTopic(user.ID, dto.CreateForumTopicRequest{
		Title: "Bad Scope", Content: "body", ScopeType: "problem",
	}); !errors.Is(err, ErrForumForbidden) {
		t.Fatalf("expected forbidden scope, got %v", err)
	}

	if _, err := svc.CreateTopic(user.ID, dto.CreateForumTopicRequest{
		Title: "Missing Problem", Content: "body", ScopeType: "problem", ScopeID: ptrUint64(99999),
	}); !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("expected problem not found, got %v", err)
	}

	if _, err := svc.CreateTopic(user.ID, dto.CreateForumTopicRequest{
		Title: "Missing Contest", Content: "body", ScopeType: "contest", ScopeID: ptrUint64(99999),
	}); !errors.Is(err, ErrContestNotFound) {
		t.Fatalf("expected contest not found, got %v", err)
	}

	topic, err := svc.CreateTopic(user.ID, dto.CreateForumTopicRequest{
		Title: "General", Content: "body", ScopeType: "general",
	})
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}

	locked := true
	if _, err := svc.UpdateTopic(user.ID, "admin", topic.ID, dto.UpdateForumTopicRequest{
		Title: "Locked", Content: "body", IsLocked: &locked,
	}); err != nil {
		t.Fatalf("lock topic: %v", err)
	}

	if _, err := svc.CreateReply(user.ID, "student", topic.ID, dto.CreateForumReplyRequest{Content: "nope"}); !errors.Is(err, ErrForumLocked) {
		t.Fatalf("expected locked topic error, got %v", err)
	}

	if _, err := svc.CreateReply(user.ID, "admin", topic.ID, dto.CreateForumReplyRequest{Content: "admin ok"}); err != nil {
		t.Fatalf("admin reply on locked topic: %v", err)
	}

	if err := svc.DeleteTopic(user.ID+1, "student", topic.ID); !errors.Is(err, ErrForumForbidden) {
		t.Fatalf("expected delete forbidden, got %v", err)
	}

	if _, err := svc.UpdateTopic(user.ID+1, "student", topic.ID, dto.UpdateForumTopicRequest{
		Title: "Hack", Content: "body",
	}); !errors.Is(err, ErrForumForbidden) {
		t.Fatalf("expected update forbidden, got %v", err)
	}

	if err := svc.LikeTopic(context.Background(), topic.ID, user.ID); err != nil {
		t.Fatalf("like topic: %v", err)
	}
	if err := svc.UnlikeTopic(context.Background(), topic.ID, user.ID); err != nil {
		t.Fatalf("unlike topic: %v", err)
	}

	if err := svc.FavoriteTopic(context.Background(), topic.ID, user.ID); err != nil {
		t.Fatalf("favorite topic: %v", err)
	}
	if err := svc.UnfavoriteTopic(context.Background(), topic.ID, user.ID); err != nil {
		t.Fatalf("unfavorite topic: %v", err)
	}
}

func TestTeachingServiceJoinAndPermissionErrors(t *testing.T) {
	db := openServiceTestDB(t)
	migrateTeachingTablesSQLite(t, db)
	teacher, student := seedTeachingUsers(t, db)
	svc := NewTeachingService(db)

	if _, err := svc.JoinClass(student.ID, "INVALID"); !errors.Is(err, ErrJoinCodeInvalid) {
		t.Fatalf("expected invalid join code, got %v", err)
	}

	classDetail, err := svc.CreateClass(teacher.ID, "teacher", dto.CreateClassRequest{Name: "Join Test", Description: "d"})
	if err != nil {
		t.Fatalf("create class: %v", err)
	}
	if _, err := svc.JoinClass(student.ID, classDetail.JoinCode); err != nil {
		t.Fatalf("join class: %v", err)
	}
	if _, err := svc.JoinClass(student.ID, classDetail.JoinCode); !errors.Is(err, ErrAlreadyJoinedClass) {
		t.Fatalf("expected duplicate join, got %v", err)
	}

	if _, err := svc.GetClassDetail(student.ID+100, "student", classDetail.ID, false); !errors.Is(err, ErrTeachingForbidden) {
		t.Fatalf("expected class forbidden, got %v", err)
	}

	if _, err := svc.ListClassMembers(teacher.ID+100, "teacher", classDetail.ID); !errors.Is(err, ErrTeachingForbidden) {
		t.Fatalf("expected members forbidden, got %v", err)
	}
}

func ptrUint64(v uint64) *uint64 {
	return &v
}
