package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/utils"
)

func openServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&model.User{},
		&model.Problem{},
		&model.ProblemTestcase{},
		&model.ProblemSolution{},
		&model.Submission{},
		&model.SubmissionResult{},
		&model.Announcement{},
		&model.Contest{},
		&model.ContestProblem{},
		&model.ContestRegistration{},
		&model.ContestAnnouncement{},
		&model.ForumTopic{},
		&model.ForumReply{},
		&model.ForumTopicLike{},
		&model.ForumTopicFavorite{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestAuthServiceLifecycle(t *testing.T) {
	db := openServiceTestDB(t)
	service := NewAuthService(db, "secret")

	user, err := service.Register(dto.RegisterRequest{Username: " alice ", UserID: " 090001 ", Password: "password"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Username != "alice" || user.UserID != "090001" || user.Role != "student" || user.Status != "active" {
		t.Fatalf("unexpected registered user: %+v", user)
	}
	if _, err := service.Register(dto.RegisterRequest{Username: "alice", UserID: "090002", Password: "password"}); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected username taken, got %v", err)
	}
	if _, err := service.Register(dto.RegisterRequest{Username: "bob", UserID: "090001", Password: "password"}); !errors.Is(err, ErrUserIDTaken) {
		t.Fatalf("expected userid taken, got %v", err)
	}
	if _, err := service.Login(dto.LoginRequest{Username: "alice", Password: "wrong-password"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	login, err := service.Login(dto.LoginRequest{Username: "alice", Password: "password"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	claims, err := utils.ParseToken("secret", login.Token)
	if err != nil {
		t.Fatalf("parse login token: %v", err)
	}
	if claims.UserID != user.ID || claims.Role != "student" {
		t.Fatalf("unexpected login claims: %+v", claims)
	}

	updated, err := service.UpdateProfile(user.ID, dto.UpdateProfileRequest{Username: "alice2", UserID: "090003"})
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.Username != "alice2" || updated.UserID != "090003" {
		t.Fatalf("unexpected updated profile: %+v", updated)
	}
	if err := service.ChangePassword(user.ID, dto.ChangePasswordRequest{CurrentPassword: "badpass", NewPassword: "new-password"}); !errors.Is(err, ErrInvalidCurrentPassword) {
		t.Fatalf("expected invalid current password, got %v", err)
	}
	if err := service.ChangePassword(user.ID, dto.ChangePasswordRequest{CurrentPassword: "password", NewPassword: "new-password"}); err != nil {
		t.Fatalf("change password: %v", err)
	}
	if _, err := service.Login(dto.LoginRequest{Username: "alice2", Password: "new-password"}); err != nil {
		t.Fatalf("login with new password: %v", err)
	}
	current, err := service.GetCurrentUser(user.ID)
	if err != nil {
		t.Fatalf("get current user: %v", err)
	}
	if current.Username != "alice2" {
		t.Fatalf("unexpected current user: %+v", current)
	}

	if err := db.Model(&model.User{}).Where("id = ?", user.ID).Update("status", "disabled").Error; err != nil {
		t.Fatalf("disable user: %v", err)
	}
	if _, err := service.Login(dto.LoginRequest{Username: "alice2", Password: "new-password"}); !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("expected disabled user error, got %v", err)
	}
}

func TestProblemServiceLifecycleStatsAndSolutions(t *testing.T) {
	db := openServiceTestDB(t)
	service := newProblemServiceForTest(db)
	req := dto.CreateProblemRequest{
		DisplayID:     " P100 ",
		Title:         " Sum ",
		Description:   "Add two numbers",
		JudgeMode:     "standard",
		Difficulty:    1,
		TimeLimitMS:   1000,
		MemoryLimitMB: 128,
		Visible:       true,
		Testcases: []dto.CreateProblemTestcaseInput{
			{CaseType: "sample", InputData: "1 1", OutputData: "2", SortOrder: 1, IsActive: true},
			{CaseType: "sample", InputData: "0 0", OutputData: "0", SortOrder: 2, IsActive: false},
			{CaseType: "hidden", InputData: "2 2", OutputData: "4", Score: 100, SortOrder: 3, IsActive: true},
		},
	}

	if _, err := service.CreateProblem(1, "student", req); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected permission denied, got %v", err)
	}
	id, err := service.CreateProblem(7, "admin", req)
	if err != nil {
		t.Fatalf("create problem: %v", err)
	}

	adminDetail, err := service.GetAdminProblemDetail("admin", id)
	if err != nil {
		t.Fatalf("admin detail: %v", err)
	}
	if adminDetail.DisplayID != "P100" || adminDetail.Title != "Sum" || len(adminDetail.Testcases) != 3 {
		t.Fatalf("unexpected admin detail: %+v", adminDetail)
	}
	adminList, err := service.ListAdminProblems("admin", 1, 20, "Sum", true)
	if err != nil {
		t.Fatalf("admin list problems: %v", err)
	}
	if adminList.Total != 1 || len(adminList.List) != 1 {
		t.Fatalf("unexpected admin problem list: %+v", adminList)
	}
	recentList, err := service.ListProblems(context.Background(), 1, 20, "")
	if err != nil {
		t.Fatalf("recent problem list: %v", err)
	}
	if recentList.Total != 1 || len(recentList.List) != 1 {
		t.Fatalf("unexpected recent problem list: %+v", recentList)
	}
	publicDetail, err := service.GetProblemDetail(id)
	if err != nil {
		t.Fatalf("public detail: %v", err)
	}
	if len(publicDetail.Testcases) != 1 || publicDetail.Testcases[0].CaseType != "sample" {
		t.Fatalf("expected only active sample testcases, got %+v", publicDetail.Testcases)
	}

	submissions := []model.Submission{
		{UserID: 9, ProblemID: id, Language: "cpp", Status: "Wrong Answer"},
		{UserID: 9, ProblemID: id, Language: "cpp", Status: "Accepted"},
		{UserID: 10, ProblemID: id, Language: "go", Status: "Accepted"},
	}
	if err := db.Create(&submissions).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}
	stats, err := service.GetProblemStats(id)
	if err != nil {
		t.Fatalf("problem stats: %v", err)
	}
	if stats.SubmissionsTotal != 3 || stats.AcceptedSubmissions != 2 || stats.AcceptedUsers != 2 {
		t.Fatalf("unexpected problem stats: %+v", stats)
	}

	difficulty := 1
	publicList, err := service.ListPublicProblems(1, 20, "P100", &difficulty)
	if err != nil {
		t.Fatalf("public list: %v", err)
	}
	if publicList.Total != 1 || len(publicList.List) != 1 || publicList.List[0].SubmissionCount != 3 {
		t.Fatalf("unexpected public list: %+v", publicList)
	}

	if _, err := service.CreateProblemSolution(11, "student", id, dto.CreateProblemSolutionRequest{Title: "No AC", Content: "content", Visibility: "public"}); !errors.Is(err, ErrSolutionPublishNotAllowed) {
		t.Fatalf("expected accepted submission requirement, got %v", err)
	}
	solution, err := service.CreateProblemSolution(9, "student", id, dto.CreateProblemSolutionRequest{Title: "  Accepted Way ", Content: "Use addition", Visibility: "public"})
	if err != nil {
		t.Fatalf("create solution: %v", err)
	}
	if solution.Title != "Accepted Way" {
		t.Fatalf("expected trimmed solution title, got %q", solution.Title)
	}
	if _, err := service.UpdateProblemSolution(10, "student", id, solution.ID, dto.CreateProblemSolutionRequest{Title: "Hack", Content: "No", Visibility: "public"}); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected update permission denied, got %v", err)
	}
	updated, err := service.UpdateProblemSolution(9, "student", id, solution.ID, dto.CreateProblemSolutionRequest{Title: "Updated", Content: "Better", Visibility: "private"})
	if err != nil {
		t.Fatalf("update solution: %v", err)
	}
	if updated.Visibility != "private" {
		t.Fatalf("unexpected updated solution: %+v", updated)
	}
	publicSolutions, err := service.ListProblemSolutions(id, false)
	if err != nil {
		t.Fatalf("list public solutions: %v", err)
	}
	if len(publicSolutions) != 0 {
		t.Fatalf("expected private solution hidden, got %+v", publicSolutions)
	}
	managedSolutions, err := service.ListManageProblemSolutions(9, "student", id)
	if err != nil {
		t.Fatalf("list managed solutions: %v", err)
	}
	if len(managedSolutions) != 1 || managedSolutions[0].ID != solution.ID {
		t.Fatalf("unexpected managed solutions: %+v", managedSolutions)
	}
	if err := service.DeleteProblemSolution(7, "admin", id, solution.ID); err != nil {
		t.Fatalf("delete solution as admin: %v", err)
	}

	req.Title = "Updated Problem"
	req.Testcases = req.Testcases[:1]
	if err := service.UpdateProblem("admin", id, req); err != nil {
		t.Fatalf("update problem: %v", err)
	}
	adminDetail, err = service.GetAdminProblemDetail("admin", id)
	if err != nil {
		t.Fatalf("admin detail after update: %v", err)
	}
	if adminDetail.Title != "Updated Problem" || len(adminDetail.Testcases) != 1 {
		t.Fatalf("unexpected updated problem detail: %+v", adminDetail)
	}
	if err := service.DeleteProblem("admin", id); err != nil {
		t.Fatalf("delete problem: %v", err)
	}
	if _, err := service.GetAdminProblemDetail("admin", id); !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("expected problem not found after delete, got %v", err)
	}
}

func TestProblemPackageImportExportIntegration(t *testing.T) {
	db := openServiceTestDB(t)
	service := newProblemServiceForTest(db)
	packageZip := buildZip(t, map[string]string{
		"problem.json": `{"display_id":"PKG1","title":"Package Problem","description":"desc","judge_mode":"standard","visible":true}`,
		"tests/2.in":   "2\n",
		"tests/2.out":  "4\n",
		"tests/1.in":   "1\n",
		"tests/1.out":  "2\n",
	})

	if _, err := service.ImportProblemPackage(1, "student", bytes.NewReader(packageZip)); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected import permission denied, got %v", err)
	}
	imported, err := service.ImportProblemPackage(1, "admin", bytes.NewReader(packageZip))
	if err != nil {
		t.Fatalf("import problem package: %v", err)
	}
	if imported.ProblemID == 0 || imported.Imported != 2 {
		t.Fatalf("unexpected import result: %+v", imported)
	}
	if _, err := service.ImportProblemPackage(1, "admin", bytes.NewReader(packageZip)); !errors.Is(err, ErrProblemDisplayIDConflict) {
		t.Fatalf("expected display id conflict, got %v", err)
	}

	exported, filename, err := service.ExportProblemPackage(imported.ProblemID)
	if err != nil {
		t.Fatalf("export problem package: %v", err)
	}
	if filename != "PKG1.zip" || !zipHasFile(t, exported, "problem.json") || !zipHasFile(t, exported, "tests/1.in") {
		t.Fatalf("unexpected exported package filename=%q", filename)
	}

	testcaseZip := buildZip(t, map[string]string{"1.in": "sample in", "1.out": "sample out"})
	result, err := service.ImportProblemTestcases(imported.ProblemID, bytes.NewReader(testcaseZip), TestcaseImportOptions{Replace: true, CaseType: "sample"})
	if err != nil {
		t.Fatalf("import testcases: %v", err)
	}
	if result.Imported != 1 || !result.Replaced {
		t.Fatalf("unexpected testcase import result: %+v", result)
	}
	if _, err := service.ImportProblemTestcases(imported.ProblemID, bytes.NewReader(testcaseZip), TestcaseImportOptions{CaseType: "bad"}); !errors.Is(err, ErrProblemPackageInvalid) {
		t.Fatalf("expected invalid case type, got %v", err)
	}

	exportedCases, caseFilename, err := service.ExportProblemTestcases(imported.ProblemID)
	if err != nil {
		t.Fatalf("export testcases: %v", err)
	}
	if caseFilename != "PKG1.zip" || !zipHasFile(t, exportedCases, "manifest.json") || !zipHasFile(t, exportedCases, "1.in") {
		t.Fatalf("unexpected exported testcases filename=%q", caseFilename)
	}
}

func TestContestServiceLifecycle(t *testing.T) {
	db := openServiceTestDB(t)
	problem := model.Problem{DisplayID: "A100", Title: "A+B", JudgeMode: "standard", Visible: true, TimeLimitMS: 1000, MemoryLimitMB: 128}
	user := model.User{Username: "contestant", UserID: "C001", Role: "student", Status: "active"}
	if err := db.Create(&problem).Error; err != nil {
		t.Fatalf("seed problem: %v", err)
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	service := NewContestService(
		db,
		repository.NewProblemRepository(db),
		repository.NewProblemTestcaseRepository(db),
		cache.New(nil),
	)
	now := time.Now()
	req := dto.CreateContestRequest{
		Title:         " Weekly ",
		Description:   "Practice contest",
		RuleType:      "acm",
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		IsPublic:      true,
		AllowPractice: true,
		Problems: []dto.CreateContestProblemDTO{
			{ProblemID: problem.ID},
		},
	}
	if _, err := service.Create(1, "student", req); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected contest create permission denied, got %v", err)
	}
	badReq := req
	badReq.EndTime = badReq.StartTime
	if _, err := service.Create(1, "admin", badReq); !errors.Is(err, ErrContestInvalidTimeRange) {
		t.Fatalf("expected invalid contest time range, got %v", err)
	}
	contestID, err := service.Create(1, "admin", req)
	if err != nil {
		t.Fatalf("create contest: %v", err)
	}

	list, err := service.List(context.Background(), 1, 20, "Weekly", "running")
	if err != nil {
		t.Fatalf("list contests: %v", err)
	}
	if list.Total != 1 || len(list.List) != 1 || list.List[0].Status != "running" {
		t.Fatalf("unexpected contest list: %+v", list)
	}
	detail, err := service.GetByID(contestID)
	if err != nil {
		t.Fatalf("contest detail: %v", err)
	}
	if detail.Title != "Weekly" || detail.ProblemCount != 1 {
		t.Fatalf("unexpected contest detail: %+v", detail)
	}
	adminDetail, err := service.GetAdminByID("admin", contestID)
	if err != nil {
		t.Fatalf("admin contest detail: %v", err)
	}
	if len(adminDetail.Problems) != 1 || adminDetail.Problems[0].ProblemCode != "A" {
		t.Fatalf("unexpected admin contest problems: %+v", adminDetail.Problems)
	}

	if err := service.Register(user.ID, "student", contestID); err != nil {
		t.Fatalf("register contest: %v", err)
	}
	me, err := service.GetMe(user.ID, "student", contestID)
	if err != nil {
		t.Fatalf("contest me: %v", err)
	}
	if !me.Registered || !me.CanSubmit || !me.CanViewProblems {
		t.Fatalf("unexpected contest me response: %+v", me)
	}
	if err := service.ValidateSubmissionAccess(user.ID, "student", contestID, problem.ID); err != nil {
		t.Fatalf("validate submission access: %v", err)
	}
	problems, err := service.ListProblems(user.ID, "student", contestID)
	if err != nil {
		t.Fatalf("list contest problems: %v", err)
	}
	if len(problems.List) != 1 || problems.List[0].ProblemID != problem.ID {
		t.Fatalf("unexpected contest problem list: %+v", problems)
	}

	announcementID, err := service.CreateAnnouncement(1, "admin", contestID, dto.CreateContestAnnouncementRequest{Title: "Pinned", Content: "Read me", IsPinned: true})
	if err != nil {
		t.Fatalf("create contest announcement: %v", err)
	}
	announcements, err := service.ListAnnouncements(contestID, 1, 20)
	if err != nil {
		t.Fatalf("list contest announcements: %v", err)
	}
	if announcements.Total != 1 || announcements.List[0].ID != announcementID {
		t.Fatalf("unexpected contest announcements: %+v", announcements)
	}
	if err := service.UpdateAnnouncement("admin", contestID, announcementID, dto.CreateContestAnnouncementRequest{Title: "Updated", Content: "Body"}); err != nil {
		t.Fatalf("update contest announcement: %v", err)
	}
	announcement, err := service.GetAnnouncement(contestID, announcementID)
	if err != nil {
		t.Fatalf("get contest announcement: %v", err)
	}
	if announcement.Title != "Updated" {
		t.Fatalf("unexpected updated contest announcement: %+v", announcement)
	}

	submissions := []model.Submission{
		{UserID: user.ID, ProblemID: problem.ID, ContestID: &contestID, Language: "cpp", Status: "Wrong Answer", CreatedAt: now.Add(10 * time.Minute)},
		{UserID: user.ID, ProblemID: problem.ID, ContestID: &contestID, Language: "cpp", Status: "Accepted", CreatedAt: now.Add(20 * time.Minute)},
	}
	if err := db.Create(&submissions).Error; err != nil {
		t.Fatalf("seed contest submissions: %v", err)
	}
	ranklist, err := service.GetRanklist(contestID)
	if err != nil {
		t.Fatalf("contest ranklist: %v", err)
	}
	if len(ranklist.List) != 1 || ranklist.List[0].SolvedCount != 1 || ranklist.List[0].Cells[0].WrongAttempts != 1 {
		t.Fatalf("unexpected contest ranklist: %+v", ranklist)
	}

	req.Title = "Weekly Updated"
	if err := service.Update("admin", contestID, req); err != nil {
		t.Fatalf("update contest: %v", err)
	}
	updated, err := service.GetAdminByID("admin", contestID)
	if err != nil {
		t.Fatalf("get updated contest: %v", err)
	}
	if updated.Title != "Weekly Updated" {
		t.Fatalf("unexpected updated contest: %+v", updated)
	}
	if err := service.DeleteAnnouncement("admin", contestID, announcementID); err != nil {
		t.Fatalf("delete contest announcement: %v", err)
	}
	if err := service.Delete("admin", contestID); err != nil {
		t.Fatalf("delete contest: %v", err)
	}
	if _, err := service.GetAdminByID("admin", contestID); !errors.Is(err, ErrContestNotFound) {
		t.Fatalf("expected contest not found after delete, got %v", err)
	}
}

func TestForumServiceLifecycle(t *testing.T) {
	db := openServiceTestDB(t)
	author := model.User{Username: "author", UserID: "F001", Role: "student", Status: "active"}
	teacher := model.User{Username: "teacher", UserID: "F002", Role: "teacher", Status: "active"}
	if err := db.Create(&[]model.User{author, teacher}).Error; err != nil {
		t.Fatalf("seed forum users: %v", err)
	}
	var users []model.User
	if err := db.Order("id ASC").Find(&users).Error; err != nil {
		t.Fatalf("load forum users: %v", err)
	}
	author = users[0]
	teacher = users[1]

	service := NewForumService(db, cache.New(nil))
	if _, err := service.CreateTopic(author.ID, dto.CreateForumTopicRequest{Title: "Bad", Content: "No", ScopeType: "problem"}); !errors.Is(err, ErrForumForbidden) {
		t.Fatalf("expected forum scope forbidden, got %v", err)
	}
	topic, err := service.CreateTopic(author.ID, dto.CreateForumTopicRequest{Title: " Hello ", Content: "World", ScopeType: "general"})
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if topic.Title != "Hello" || topic.AuthorName != "author" {
		t.Fatalf("unexpected topic: %+v", topic)
	}

	if err := service.LikeTopic(context.Background(), topic.ID, author.ID); err != nil {
		t.Fatalf("like topic: %v", err)
	}
	if err := service.FavoriteTopic(context.Background(), topic.ID, author.ID); err != nil {
		t.Fatalf("favorite topic: %v", err)
	}
	withReactions, err := service.GetTopicDetail(topic.ID, &author.ID)
	if err != nil {
		t.Fatalf("topic detail with reactions: %v", err)
	}
	if !withReactions.IsLiked || !withReactions.IsFavorited {
		t.Fatalf("expected hydrated reactions, got %+v", withReactions)
	}
	if err := service.UnlikeTopic(context.Background(), topic.ID, author.ID); err != nil {
		t.Fatalf("unlike topic: %v", err)
	}
	if err := service.UnfavoriteTopic(context.Background(), topic.ID, author.ID); err != nil {
		t.Fatalf("unfavorite topic: %v", err)
	}

	reply, err := service.CreateReply(author.ID, "student", topic.ID, dto.CreateForumReplyRequest{Content: "First reply"})
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}
	if reply.AuthorName != "author" || reply.TopicID != topic.ID {
		t.Fatalf("unexpected reply: %+v", reply)
	}
	updatedTopic, err := service.UpdateTopic(teacher.ID, "teacher", topic.ID, dto.UpdateForumTopicRequest{Title: "Locked", Content: "Body", IsLocked: boolPtr(true)})
	if err != nil {
		t.Fatalf("lock topic: %v", err)
	}
	if !updatedTopic.IsLocked {
		t.Fatalf("expected locked topic: %+v", updatedTopic)
	}
	if _, err := service.CreateReply(author.ID, "student", topic.ID, dto.CreateForumReplyRequest{Content: "Blocked"}); !errors.Is(err, ErrForumLocked) {
		t.Fatalf("expected locked topic error, got %v", err)
	}
	if _, err := service.CreateReply(teacher.ID, "teacher", topic.ID, dto.CreateForumReplyRequest{Content: "Allowed"}); err != nil {
		t.Fatalf("teacher reply on locked topic: %v", err)
	}
	if _, err := service.UpdateReply(author.ID, "student", reply.ID, dto.CreateForumReplyRequest{Content: "Edited"}); err != nil {
		t.Fatalf("update own reply: %v", err)
	}
	if err := service.DeleteTopic(author.ID, "student", topic.ID); err != nil {
		t.Fatalf("delete own topic: %v", err)
	}
	if _, err := service.GetTopicDetail(topic.ID, nil); !errors.Is(err, ErrForumTopicNotFound) {
		t.Fatalf("expected deleted topic not found, got %v", err)
	}
}

func TestAnnouncementRanklistStatsAndUserServices(t *testing.T) {
	db := openServiceTestDB(t)
	hash := mustPasswordHash(t, "password")
	users := []model.User{
		{Username: "alice", UserID: "S001", PasswordHash: hash, Role: "student", Status: "active"},
		{Username: "bob", UserID: "S002", PasswordHash: hash, Role: "teacher", Status: "active"},
		{Username: "disabled", UserID: "S003", PasswordHash: hash, Role: "student", Status: "disabled"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	userService := NewUserService(db)
	if _, err := userService.List("student", 1, 20, "", "", ""); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected user list permission denied, got %v", err)
	}
	userList, err := userService.List("admin", 1, 20, "S00", "", "active")
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if userList.Total != 2 || len(userList.List) != 2 {
		t.Fatalf("unexpected user list: %+v", userList)
	}
	if _, err := userService.Update("admin", users[1].ID, dto.AdminUserUpdateRequest{Username: "alice", UserID: "S009", Role: "teacher", Status: "active"}); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected duplicate username, got %v", err)
	}
	updatedUser, err := userService.Update("admin", users[1].ID, dto.AdminUserUpdateRequest{Username: "bob2", UserID: "S022", Role: "teacher", Status: "active"})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updatedUser.Username != "bob2" || updatedUser.UserID != "S022" {
		t.Fatalf("unexpected updated user: %+v", updatedUser)
	}

	announcementService := NewAnnouncementService(db, cache.New(nil))
	if _, err := announcementService.Create(users[0].ID, "student", dto.CreateAnnouncementRequest{Title: "Denied", Content: "No"}); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected announcement permission denied, got %v", err)
	}
	firstID, err := announcementService.Create(users[0].ID, "admin", dto.CreateAnnouncementRequest{Title: " First ", Content: "Hello"})
	if err != nil {
		t.Fatalf("create first announcement: %v", err)
	}
	secondID, err := announcementService.Create(users[0].ID, "admin", dto.CreateAnnouncementRequest{Title: "Pinned", Content: "World", IsPinned: true})
	if err != nil {
		t.Fatalf("create second announcement: %v", err)
	}
	announcements, err := announcementService.List(1, 20)
	if err != nil {
		t.Fatalf("list announcements: %v", err)
	}
	if announcements.Total != 2 || announcements.List[0].ID != secondID {
		t.Fatalf("expected pinned announcement first, got %+v", announcements)
	}
	if err := announcementService.Update("admin", firstID, dto.CreateAnnouncementRequest{Title: "Updated", Content: "Body", IsPinned: true}); err != nil {
		t.Fatalf("update announcement: %v", err)
	}
	announcement, err := announcementService.GetByID(firstID)
	if err != nil {
		t.Fatalf("get announcement: %v", err)
	}
	if announcement.Title != "Updated" || !announcement.IsPinned {
		t.Fatalf("unexpected announcement after update: %+v", announcement)
	}
	if err := announcementService.Delete("admin", secondID); err != nil {
		t.Fatalf("delete announcement: %v", err)
	}

	problems := []model.Problem{
		{DisplayID: "A", Title: "A", JudgeMode: "standard", Visible: true},
		{DisplayID: "B", Title: "B", JudgeMode: "standard", Visible: false},
	}
	if err := db.Create(&problems).Error; err != nil {
		t.Fatalf("seed problems: %v", err)
	}
	submissions := []model.Submission{
		{UserID: users[0].ID, ProblemID: problems[0].ID, Language: "cpp", Status: "Wrong Answer"},
		{UserID: users[0].ID, ProblemID: problems[0].ID, Language: "cpp", Status: "Accepted"},
		{UserID: users[1].ID, ProblemID: problems[0].ID, Language: "go", Status: "Accepted"},
	}
	if err := db.Create(&submissions).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	statsService := NewStatsService(db, nil, cache.New(nil))
	overview, err := statsService.Overview(context.Background())
	if err != nil {
		t.Fatalf("overview stats: %v", err)
	}
	if overview.ProblemsTotal != 2 || overview.VisibleProblems != 1 || overview.UsersTotal != 3 || overview.AcceptedSubmissions != 2 {
		t.Fatalf("unexpected overview stats: %+v", overview)
	}

	if err := db.Model(&model.Submission{}).Where("status = ?", "Accepted").Update("status", "Wrong Answer").Error; err != nil {
		t.Fatalf("reset accepted submissions for sqlite ranklist scan: %v", err)
	}
	ranklistService := NewRanklistService(db, cache.New(nil))
	ranklist, err := ranklistService.List(1, 20)
	if err != nil {
		t.Fatalf("ranklist: %v", err)
	}
	if ranklist.Total != 2 || len(ranklist.List) != 2 || ranklist.List[0].SolvedCount != 0 {
		t.Fatalf("unexpected ranklist: %+v", ranklist)
	}
}

func newProblemServiceForTest(db *gorm.DB) *ProblemService {
	return NewProblemService(
		db,
		repository.NewProblemRepository(db),
		repository.NewProblemTestcaseRepository(db),
		cache.New(nil),
	)
}

func mustPasswordHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(hash)
}

func zipHasFile(t *testing.T, data []byte, name string) bool {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, file := range reader.File {
		if file.Name == name {
			return true
		}
	}
	return false
}

func boolPtr(value bool) *bool {
	return &value
}
