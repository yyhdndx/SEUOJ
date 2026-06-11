package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIAuthSuccessPaths(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/register", map[string]any{
		"username": "loginok", "userid": "LOG001", "password": "Secret123!",
	}, nil, 0, "", env.auth.Register)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("register failed")
	}

	w = callHandler(t, http.MethodPost, "/login", map[string]any{
		"username": "loginok", "password": "Secret123!",
	}, nil, 0, "", env.auth.Login)
	loginEnv := decodeAPIEnvelope(t, w)
	if loginEnv.Code != 0 {
		t.Fatalf("login failed: %s", w.Body.String())
	}
	var loginData struct {
		User struct {
			ID uint64 `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginEnv.Data, &loginData); err != nil {
		t.Fatalf("decode login: %v", err)
	}

	w = callHandler(t, http.MethodPut, "/password", map[string]any{
		"current_password": "Secret123!", "new_password": "NewSecret456!",
	}, nil, loginData.User.ID, "student", env.auth.ChangePassword)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("change password failed")
	}
	w = callHandler(t, http.MethodPost, "/login", map[string]any{
		"username": "loginok", "password": "NewSecret456!",
	}, nil, 0, "", env.auth.Login)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("login with new password failed")
	}

	w = callHandler(t, http.MethodPost, "/register", map[string]any{
		"username": "loginok", "userid": "LOG002", "password": "Secret123!",
	}, nil, 0, "", env.auth.Register)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected duplicate username error")
	}
}

func TestAPIHandlersNotFoundPaths(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodGet, "/problems/:id", nil, gin.Params{{Key: "id", Value: "999999"}}, 0, "", env.problem.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected problem not found")
	}
	w = callHandler(t, http.MethodGet, "/submissions/:id", nil, gin.Params{{Key: "id", Value: "999999"}}, env.studentID, "student", env.submission.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected submission not found")
	}
	w = callHandler(t, http.MethodGet, "/announcements/:id", nil, gin.Params{{Key: "id", Value: "999999"}}, 0, "", env.announcement.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected announcement not found")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id", nil, gin.Params{{Key: "id", Value: "999999"}}, 0, "", env.contest.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected contest not found")
	}
}
