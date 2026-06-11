package service

import (
	"errors"
	"testing"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

func seedTeachingUsers(t *testing.T, db *gorm.DB) (teacher, student model.User) {
	t.Helper()
	teacher = model.User{Username: "t_full", UserID: "TFULL", PasswordHash: "hash", Role: "teacher", Status: "active"}
	student = model.User{Username: "s_full", UserID: "SFULL", PasswordHash: "hash", Role: "student", Status: "active"}
	if err := db.Create(&teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("create student: %v", err)
	}
	return teacher, student
}

func seedTeachingProblem(t *testing.T, db *gorm.DB, teacherID uint64, displayID string) model.Problem {
	t.Helper()
	problem := model.Problem{
		DisplayID: displayID, Title: displayID, Description: "desc", JudgeMode: "standard",
		TimeLimitMS: 1000, MemoryLimitMB: 128, Visible: true, CreatedBy: teacherID,
	}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	return problem
}

func TestTeachingModuleLifecycle(t *testing.T) {
	db := openServiceTestDB(t)
	migrateTeachingTablesSQLite(t, db)
	teacher, student := seedTeachingUsers(t, db)
	problem := seedTeachingProblem(t, db, teacher.ID, "TEACH-P1")
	svc := NewTeachingService(db)

	publicList, err := svc.ListPublicPlaylists(1, 20, "")
	if err != nil {
		t.Fatalf("list public playlists: %v", err)
	}
	if publicList.Total != 0 {
		t.Fatalf("expected empty public playlists, got %+v", publicList)
	}

	created, err := svc.CreatePlaylist(teacher.ID, "teacher", dto.CreatePlaylistRequest{
		Title: "Week 1", Description: "Intro", Visibility: "public",
		Problems: []dto.PlaylistProblemRequest{{ProblemID: problem.ID, DisplayOrder: 1}},
	})
	if err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	if created.Title != "Week 1" || len(created.Problems) != 1 {
		t.Fatalf("unexpected created playlist: %+v", created)
	}

	publicList, err = svc.ListPublicPlaylists(1, 20, "Week")
	if err != nil {
		t.Fatalf("list public playlists after create: %v", err)
	}
	if publicList.Total != 1 {
		t.Fatalf("expected one public playlist, got %+v", publicList)
	}

	teacherList, err := svc.ListTeacherPlaylists(teacher.ID, "teacher", 1, 20, "")
	if err != nil {
		t.Fatalf("list teacher playlists: %v", err)
	}
	if teacherList.Total != 1 {
		t.Fatalf("unexpected teacher playlists: %+v", teacherList)
	}

	updated, err := svc.UpdatePlaylist(teacher.ID, "teacher", created.ID, dto.UpdatePlaylistRequest{
		Title: "Week 1 Updated", Description: "Intro v2", Visibility: "public",
		Problems: []dto.PlaylistProblemRequest{{ProblemID: problem.ID, DisplayOrder: 1}},
	})
	if err != nil {
		t.Fatalf("update playlist: %v", err)
	}
	if updated.Title != "Week 1 Updated" {
		t.Fatalf("unexpected updated playlist: %+v", updated)
	}

	classDetail, err := svc.CreateClass(teacher.ID, "teacher", dto.CreateClassRequest{
		Name: "DS Class", Description: "2026 spring",
	})
	if err != nil {
		t.Fatalf("create class: %v", err)
	}
	if classDetail.Name != "DS Class" || classDetail.JoinCode == "" {
		t.Fatalf("unexpected class detail: %+v", classDetail)
	}

	joined, err := svc.JoinClass(student.ID, classDetail.JoinCode)
	if err != nil {
		t.Fatalf("join class: %v", err)
	}
	if joined.ID != classDetail.ID {
		t.Fatalf("unexpected joined class: %+v", joined)
	}
	if _, err := svc.JoinClass(student.ID, classDetail.JoinCode); !errors.Is(err, ErrAlreadyJoinedClass) {
		t.Fatalf("expected already joined, got %v", err)
	}

	studentClasses, err := svc.ListMyClasses(student.ID, "student")
	if err != nil {
		t.Fatalf("list student classes: %v", err)
	}
	if len(studentClasses) != 1 || studentClasses[0].Name != "DS Class" {
		t.Fatalf("unexpected student classes: %+v", studentClasses)
	}

	teacherClasses, err := svc.ListTeacherClasses(teacher.ID, "teacher")
	if err != nil {
		t.Fatalf("list teacher classes: %v", err)
	}
	if len(teacherClasses) != 1 {
		t.Fatalf("unexpected teacher classes: %+v", teacherClasses)
	}

	classWithMembers, err := svc.GetClassDetail(teacher.ID, "teacher", classDetail.ID, true)
	if err != nil {
		t.Fatalf("get class detail: %v", err)
	}
	if len(classWithMembers.Members) < 2 {
		t.Fatalf("expected teacher and student members, got %+v", classWithMembers.Members)
	}

	members, err := svc.ListClassMembers(teacher.ID, "teacher", classDetail.ID)
	if err != nil {
		t.Fatalf("list class members: %v", err)
	}
	if len(members) < 2 {
		t.Fatalf("unexpected members: %+v", members)
	}

	assignment, err := svc.CreateAssignment(teacher.ID, "teacher", classDetail.ID, dto.CreateAssignmentRequest{
		PlaylistID: created.ID, Title: "HW1", Description: "First homework", Type: "homework",
	})
	if err != nil {
		t.Fatalf("create assignment: %v", err)
	}
	if assignment.Title != "HW1" || assignment.TotalCount != 1 {
		t.Fatalf("unexpected assignment: %+v", assignment)
	}

	assignments, err := svc.ListTeacherAssignments(teacher.ID, "teacher", classDetail.ID)
	if err != nil {
		t.Fatalf("list teacher assignments: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("unexpected assignment list: %+v", assignments)
	}

	studentAssignment, err := svc.GetAssignmentDetail(student.ID, "student", assignment.ID)
	if err != nil {
		t.Fatalf("student assignment detail: %v", err)
	}
	if studentAssignment.TotalCount != 1 {
		t.Fatalf("unexpected student assignment: %+v", studentAssignment)
	}

	if err := db.Create(&model.Submission{
		UserID: student.ID, ProblemID: problem.ID, Language: "cpp", Code: "main", Status: "Accepted",
	}).Error; err != nil {
		t.Fatalf("seed accepted submission: %v", err)
	}

	studentAssignment, err = svc.GetAssignmentDetail(student.ID, "student", assignment.ID)
	if err != nil {
		t.Fatalf("assignment detail after ac: %v", err)
	}
	if studentAssignment.SolvedCount != 1 {
		t.Fatalf("expected solved count 1, got %+v", studentAssignment)
	}

	overview, err := svc.GetTeacherAssignmentOverview(teacher.ID, "teacher", assignment.ID)
	if err != nil {
		t.Fatalf("assignment overview: %v", err)
	}
	if overview.MemberCount < 2 || len(overview.Members) < 2 {
		t.Fatalf("unexpected overview: member_count=%d members=%d", overview.MemberCount, len(overview.Members))
	}

	analytics, err := svc.GetTeacherClassAnalytics(teacher.ID, "teacher", classDetail.ID)
	if err != nil {
		t.Fatalf("class analytics: %v", err)
	}
	if len(analytics.Assignments) != 1 || len(analytics.TopStudents) == 0 {
		t.Fatalf("unexpected analytics: assignments=%d top=%d", len(analytics.Assignments), len(analytics.TopStudents))
	}

	updatedAssignment, err := svc.UpdateAssignment(teacher.ID, "teacher", assignment.ID, dto.UpdateAssignmentRequest{
		PlaylistID: created.ID, Title: "HW1 Revised", Description: "Updated", Type: "exam",
	})
	if err != nil {
		t.Fatalf("update assignment: %v", err)
	}
	if updatedAssignment.Title != "HW1 Revised" || updatedAssignment.Type != "exam" {
		t.Fatalf("unexpected updated assignment: %+v", updatedAssignment)
	}

	updatedClass, err := svc.UpdateClass(teacher.ID, "teacher", classDetail.ID, dto.UpdateClassRequest{
		Name: "DS Class Updated", Description: "v2", Status: "active",
	})
	if err != nil {
		t.Fatalf("update class: %v", err)
	}
	if updatedClass.Name != "DS Class Updated" {
		t.Fatalf("unexpected updated class: %+v", updatedClass)
	}

	if err := svc.DeleteAssignment(teacher.ID, "teacher", assignment.ID); err != nil {
		t.Fatalf("delete assignment: %v", err)
	}
	if err := svc.DeletePlaylist(teacher.ID, "teacher", created.ID); err != nil {
		t.Fatalf("delete playlist: %v", err)
	}
	if _, err := svc.GetPlaylistDetail(student.ID, "student", created.ID, false); !errors.Is(err, ErrPlaylistNotFound) {
		t.Fatalf("expected playlist not found after delete, got %v", err)
	}
}

func TestTeachingClassMemberAndUniqueCode(t *testing.T) {
	db := openServiceTestDB(t)
	migrateTeachingTablesSQLite(t, db)
	teacher, student := seedTeachingUsers(t, db)
	svc := NewTeachingService(db)

	uniqueClass, err := svc.CreateClassWithUniqueCode(teacher.ID, "teacher", dto.CreateClassRequest{
		Name: "Unique Class", Description: "auto code",
	})
	if err != nil {
		t.Fatalf("create class with unique code: %v", err)
	}
	if uniqueClass.JoinCode == "" || len(uniqueClass.Members) < 1 {
		t.Fatalf("unexpected unique class: %+v", uniqueClass)
	}

	if _, err := svc.JoinClass(student.ID, uniqueClass.JoinCode); err != nil {
		t.Fatalf("student join: %v", err)
	}

	updatedMembers, err := svc.UpdateClassMember(teacher.ID, "teacher", uniqueClass.ID, student.ID, dto.UpdateClassMemberRequest{
		Role: "assistant", Status: "active",
	})
	if err != nil {
		t.Fatalf("update class member: %v", err)
	}
	found := false
	for _, m := range updatedMembers {
		if m.UserID == student.ID && m.Role == "assistant" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected assistant role for student, got %+v", updatedMembers)
	}
}
