package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/model"
)

func TestAPIHandlersMoreIntegration(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodGet, "/problems?page=1&page_size=10", nil, nil, 0, "", env.problem.List)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("problem list failed")
	}
	w = callHandler(t, http.MethodGet, "/public/problems?page=1&page_size=10", nil, nil, 0, "", env.problem.PublicList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("public problem list failed")
	}
	w = callHandler(t, http.MethodGet, "/admin/problems?page=1&page_size=10", nil, nil, env.adminID, "admin", env.problem.AdminList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin problem list failed")
	}

	w = callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "API-P2", "title": "More", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 2, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	prob := decodeAPIEnvelope(t, w)
	if prob.Code != 0 {
		t.Fatalf("create problem: %+v", prob)
	}
	var problemID struct {
		ProblemID uint64 `json:"problem_id"`
	}
	if err := json.Unmarshal(prob.Data, &problemID); err != nil {
		t.Fatalf("decode problem id: %v", err)
	}

	w = callHandler(t, http.MethodGet, "/admin/problems/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.AdminDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin problem detail failed")
	}
	w = callHandler(t, http.MethodPut, "/admin/problems/:id", map[string]any{
		"display_id": "API-P2", "title": "More Updated", "description": "d2", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 2, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "1", "output_data": "1", "sort_order": 1, "is_active": true},
		},
	}, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.Update)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update problem failed")
	}

	w = callHandler(t, http.MethodPost, "/admin/problems/:id/solutions", map[string]any{
		"title": "Admin Sol", "content": "hint", "visibility": "public",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.CreateSolution)
	sol := decodeAPIEnvelope(t, w)
	if sol.Code != 0 {
		t.Fatalf("create solution: %+v", sol)
	}
	var solutionID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(sol.Data, &solutionID)
	w = callHandler(t, http.MethodDelete, "/teacher/problems/:id/solutions/:solution_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}, {Key: "solution_id", Value: fmt.Sprint(solutionID.ID)}},
		env.adminID, "admin", env.problem.DeleteSolution)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete solution failed")
	}

	now := time.Now()
	w = callHandler(t, http.MethodPost, "/admin/contests", map[string]any{
		"title": "More Contest", "description": "d", "rule_type": "acm",
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
		"is_public": true, "allow_practice": true,
		"problems": []map[string]any{{"problem_id": problemID.ProblemID}},
	}, nil, env.adminID, "admin", env.contest.Create)
	ct := decodeAPIEnvelope(t, w)
	if ct.Code != 0 {
		t.Fatalf("create contest: %+v", ct)
	}
	var contestID struct {
		ContestID uint64 `json:"contest_id"`
	}
	_ = json.Unmarshal(ct.Data, &contestID)

	w = callHandler(t, http.MethodGet, "/admin/contests/:id/ranklist", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.adminID, "admin", env.contest.AdminRanklist)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin ranklist failed")
	}
	w = callHandler(t, http.MethodGet, "/contests/:id/announcements", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, 0, "", env.contest.AnnouncementList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest announcement list failed")
	}
	w = callHandler(t, http.MethodPost, "/admin/contests/:id/announcements", map[string]any{
		"title": "Notice", "content": "Go!", "is_pinned": true,
	}, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.adminID, "admin", env.contest.CreateAnnouncement)
	ann := decodeAPIEnvelope(t, w)
	if ann.Code != 0 {
		t.Fatalf("create contest announcement: %+v", ann)
	}
	var contestAnnID struct {
		AnnouncementID uint64 `json:"announcement_id"`
	}
	_ = json.Unmarshal(ann.Data, &contestAnnID)
	w = callHandler(t, http.MethodGet, "/contests/:id/announcements/:announcement_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: fmt.Sprint(contestAnnID.AnnouncementID)}},
		0, "", env.contest.AnnouncementDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest announcement detail failed")
	}
	w = callHandler(t, http.MethodGet, "/admin/contests/:id/announcements", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.adminID, "admin", env.contest.AdminAnnouncementList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin contest announcement list failed")
	}
	w = callHandler(t, http.MethodGet, "/admin/contests/:id/announcements/:announcement_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: fmt.Sprint(contestAnnID.AnnouncementID)}},
		env.adminID, "admin", env.contest.AdminAnnouncementDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin contest announcement detail failed")
	}
	w = callHandler(t, http.MethodPut, "/admin/contests/:id/announcements/:announcement_id", map[string]any{
		"title": "Notice v2", "content": "Updated", "is_pinned": false,
	}, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: fmt.Sprint(contestAnnID.AnnouncementID)}},
		env.adminID, "admin", env.contest.UpdateAnnouncement)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update contest announcement failed")
	}
	w = callHandler(t, http.MethodGet, "/admin/contests/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.adminID, "admin", env.contest.AdminDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("admin contest detail failed")
	}

	w = callHandler(t, http.MethodPost, "/teacher/classes", map[string]any{
		"name": "Members Class", "description": "d",
	}, nil, env.teacherID, "teacher", env.teaching.CreateClass)
	cls := decodeAPIEnvelope(t, w)
	if cls.Code != 0 {
		t.Fatalf("create class: %+v", cls)
	}
	var classInfo struct {
		ID       uint64 `json:"id"`
		JoinCode string `json:"join_code"`
	}
	_ = json.Unmarshal(cls.Data, &classInfo)
	w = callHandler(t, http.MethodPost, "/classes/join", map[string]any{"join_code": classInfo.JoinCode},
		nil, env.studentID, "student", env.teaching.JoinClass)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("join class failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/classes", nil, nil, env.teacherID, "teacher", env.teaching.TeacherClasses)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("teacher classes failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/classes/:id/members", nil, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.ClassMembers)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("class members failed")
	}
	w = callHandler(t, http.MethodPut, "/teacher/classes/:id/members/:user_id", map[string]any{
		"role": "assistant", "status": "active",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}, {Key: "user_id", Value: fmt.Sprint(env.studentID)}},
		env.teacherID, "teacher", env.teaching.UpdateClassMember)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update class member failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/classes/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.TeacherClassDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("teacher class detail failed")
	}
	w = callHandler(t, http.MethodPut, "/teacher/classes/:id", map[string]any{
		"name": "Members Class v2", "description": "updated", "status": "active",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.UpdateClass)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update class failed")
	}

	w = callHandler(t, http.MethodPost, "/teacher/playlists", map[string]any{
		"title": "HW Playlist", "description": "d", "visibility": "public",
		"problems": []map[string]any{{"problem_id": problemID.ProblemID, "display_order": 1}},
	}, nil, env.teacherID, "teacher", env.teaching.CreatePlaylist)
	pl := decodeAPIEnvelope(t, w)
	if pl.Code != 0 {
		t.Fatalf("create playlist: %+v", pl)
	}
	var playlistID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(pl.Data, &playlistID)
	w = callHandler(t, http.MethodGet, "/playlists/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(playlistID.ID)}}, env.studentID, "student", env.teaching.PlaylistDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("playlist detail failed")
	}
	w = callHandler(t, http.MethodPost, "/teacher/classes/:id/assignments", map[string]any{
		"playlist_id": playlistID.ID, "title": "HW1", "description": "first", "type": "homework",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.CreateAssignment)
	asg := decodeAPIEnvelope(t, w)
	if asg.Code != 0 {
		t.Fatalf("create assignment: %+v", asg)
	}
	var assignmentID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(asg.Data, &assignmentID)
	w = callHandler(t, http.MethodGet, "/teacher/classes/:id/assignments", nil, gin.Params{{Key: "id", Value: fmt.Sprint(classInfo.ID)}}, env.teacherID, "teacher", env.teaching.TeacherAssignments)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("teacher assignments failed")
	}
	w = callHandler(t, http.MethodGet, "/assignments/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(assignmentID.ID)}}, env.studentID, "student", env.teaching.AssignmentDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("assignment detail failed")
	}

	w = callHandler(t, http.MethodPost, "/forum/topics", map[string]any{
		"title": "Detail Topic", "content": "body", "scope_type": "general",
	}, nil, env.studentID, "student", env.forum.CreateTopic)
	topic := decodeAPIEnvelope(t, w)
	if topic.Code != 0 {
		t.Fatalf("create topic: %+v", topic)
	}
	var topicID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(topic.Data, &topicID)
	w = callHandler(t, http.MethodGet, "/forum/topics/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.TopicDetail)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("topic detail failed")
	}
	w = callHandler(t, http.MethodPut, "/forum/topics/:id", map[string]any{
		"title": "Detail Topic v2", "content": "updated body",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.UpdateTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("update topic failed")
	}
	w = callHandler(t, http.MethodPost, "/forum/topics/:id/favorite", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.FavoriteTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("favorite topic failed")
	}
	w = callHandler(t, http.MethodDelete, "/forum/topics/:id/favorite", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.UnfavoriteTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("unfavorite topic failed")
	}
	w = callHandler(t, http.MethodPost, "/forum/topics/:id/like", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.LikeTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("like topic failed")
	}
	w = callHandler(t, http.MethodDelete, "/forum/topics/:id/like", nil, gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.UnlikeTopic)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("unlike topic failed")
	}
	w = callHandler(t, http.MethodPost, "/forum/topics/:id/replies", map[string]any{"content": "to delete"},
		gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}}, env.studentID, "student", env.forum.CreateReply)
	reply := decodeAPIEnvelope(t, w)
	if reply.Code != 0 {
		t.Fatalf("create reply: %+v", reply)
	}
	var replyID struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(reply.Data, &replyID)
	w = callHandler(t, http.MethodDelete, "/forum/topics/:id/replies/:reply_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(topicID.ID)}, {Key: "reply_id", Value: fmt.Sprint(replyID.ID)}},
		env.studentID, "student", env.forum.DeleteReply)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete reply failed")
	}

	if err := env.db.Create(&model.Submission{
		UserID: env.studentID, ProblemID: problemID.ProblemID, Language: "cpp", Code: "main", Status: "Accepted",
	}).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}
	w = callHandler(t, http.MethodPost, "/contests/:id/register", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}}, env.studentID, "student", env.contest.Register)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("contest register failed")
	}
	w = callHandler(t, http.MethodPost, "/submissions", map[string]any{
		"problem_id": problemID.ProblemID, "contest_id": contestID.ContestID, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Create)
	subResp := decodeAPIEnvelope(t, w)
	if subResp.Code != 0 {
		t.Fatal("contest submission create failed")
	}
	var contestSubID struct {
		SubmissionID uint64 `json:"submission_id"`
	}
	_ = json.Unmarshal(subResp.Data, &contestSubID)
	w = callHandler(t, http.MethodPost, "/admin/submissions/:id/rejudge", nil, gin.Params{{Key: "id", Value: fmt.Sprint(contestSubID.SubmissionID)}}, env.adminID, "admin", env.submission.Rejudge)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("rejudge failed")
	}

	w = callHandler(t, http.MethodDelete, "/admin/contests/:id/announcements/:announcement_id", nil,
		gin.Params{{Key: "id", Value: fmt.Sprint(contestID.ContestID)}, {Key: "announcement_id", Value: fmt.Sprint(contestAnnID.AnnouncementID)}},
		env.adminID, "admin", env.contest.DeleteAnnouncement)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete contest announcement failed")
	}

	w = callHandler(t, http.MethodGet, "/problems/:id", nil, gin.Params{{Key: "id", Value: "bad"}}, 0, "", env.problem.Detail)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid problem id error")
	}
}
