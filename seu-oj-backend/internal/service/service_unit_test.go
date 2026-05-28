package service

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

func TestParseTestcaseZipSortsPairsAndHonorsPrefix(t *testing.T) {
	data := buildZip(t, map[string]string{
		"tests/2.in":   "2\n",
		"tests/2.out":  "4\n",
		"tests/1.in":   "1\n",
		"tests/1.out":  "2\n",
		"other/9.in":   "ignored",
		"other/9.out":  "ignored",
		"problem.json": "{}",
	})

	pairs, err := parseTestcaseZip(data, "tests/")
	if err != nil {
		t.Fatalf("parse zip: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected two testcase pairs, got %+v", pairs)
	}
	if pairs[0].Order != 1 || pairs[0].Input != "1\n" || pairs[1].Order != 2 {
		t.Fatalf("unexpected parsed pairs: %+v", pairs)
	}
}

func TestParseTestcaseZipRejectsInvalidPackages(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "not zip", data: []byte("not a zip")},
		{name: "missing output", data: buildZip(t, map[string]string{"1.in": "1"})},
		{name: "extra output", data: buildZip(t, map[string]string{"1.in": "1", "1.out": "1", "2.out": "2"})},
		{name: "bad order", data: buildZip(t, map[string]string{"0.in": "1", "0.out": "1"})},
		{name: "path traversal", data: buildZip(t, map[string]string{"../1.in": "1", "../1.out": "1"})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseTestcaseZip(tt.data, ""); !errors.Is(err, ErrProblemPackageInvalid) {
				t.Fatalf("expected invalid package, got %v", err)
			}
		})
	}
}

func TestProblemPackageHelpers(t *testing.T) {
	if got := safePackageName(" P-100_测试! ", "fallback"); got != "P-100_" {
		t.Fatalf("unexpected safe package name %q", got)
	}
	if got := safePackageName("测试", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback for unsafe name, got %q", got)
	}
	if _, err := readLimitedPackage(strings.NewReader("")); !errors.Is(err, ErrProblemPackageInvalid) {
		t.Fatalf("expected empty package invalid, got %v", err)
	}
	data, err := readLimitedPackage(strings.NewReader("abc"))
	if err != nil || string(data) != "abc" {
		t.Fatalf("unexpected limited package data=%q err=%v", string(data), err)
	}

	manifest := buildTestcaseManifest([]model.ProblemTestcase{
		{CaseType: "sample", Score: 10, SortOrder: 0, IsActive: true},
		{CaseType: "hidden", Score: 90, SortOrder: 5, IsActive: false},
	}, "tests/")
	if len(manifest) != 2 || manifest[0]["input_file"] != "tests/1.in" || manifest[1]["output_file"] != "tests/5.out" {
		t.Fatalf("unexpected testcase manifest: %+v", manifest)
	}
}

func TestOutputNormalizationComparison(t *testing.T) {
	if !compareOutput("1  \r\n2\t\n\n", "1\n2") {
		t.Fatal("expected output normalization to ignore trailing whitespace and final blank lines")
	}
	if compareOutput("1 2", "1  2") {
		t.Fatal("did not expect internal whitespace to be ignored")
	}
}

func TestProblemDetailConversionAndSampleFiltering(t *testing.T) {
	now := time.Now().UTC()
	problem := model.Problem{
		ID:            10,
		DisplayID:     "A100",
		Title:         "A+B",
		Description:   "desc",
		JudgeMode:     "standard",
		Difficulty:    1,
		TimeLimitMS:   1000,
		MemoryLimitMB: 128,
		Visible:       true,
		CreatedBy:     7,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	testcases := []model.ProblemTestcase{
		{ID: 1, CaseType: "sample", InputData: "1", OutputData: "1", SortOrder: 1, IsActive: true},
		{ID: 2, CaseType: "sample", InputData: "2", OutputData: "2", SortOrder: 2, IsActive: false},
		{ID: 3, CaseType: "hidden", InputData: "3", OutputData: "3", SortOrder: 3, IsActive: true},
	}

	filtered := filterSampleTestcases(testcases)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("unexpected sample filter result: %+v", filtered)
	}
	resp := toProblemDetailResponse(problem, filtered, []dto.ProblemSolutionResponse{{ID: 99, Title: "solution"}})
	if resp.ID != problem.ID || len(resp.Testcases) != 1 || resp.Testcases[0].OutputData != "1" || len(resp.Solutions) != 1 {
		t.Fatalf("unexpected problem detail response: %+v", resp)
	}
}

func TestSubmissionListHelpers(t *testing.T) {
	status := ""
	problemID := uint64(1)
	if !isUnfilteredRecentSubmissionQuery(1, 100, nil, nil, nil) {
		t.Fatal("expected recent query for unfiltered first page")
	}
	if !isUnfilteredRecentSubmissionQuery(1, 20, nil, nil, &status) {
		t.Fatal("expected empty status to count as unfiltered")
	}
	if isUnfilteredRecentSubmissionQuery(2, 20, nil, nil, nil) || isUnfilteredRecentSubmissionQuery(1, 20, &problemID, nil, nil) {
		t.Fatal("did not expect filtered or non-first-page query to be recent")
	}
	if uintFilterKey(nil) != "all" || uintFilterKey(&problemID) != "1" {
		t.Fatal("unexpected uint filter key")
	}
	status = "Accepted"
	if stringFilterKey(nil) != "all" || stringFilterKey(&status) != "Accepted" {
		t.Fatal("unexpected string filter key")
	}

	runtime := 12
	resp := toSubmissionListResponse([]model.Submission{
		{ID: 1, UserID: 2, ProblemID: 3, Language: "go", Status: "Accepted", RuntimeMS: &runtime},
	}, 5, 1, 20)
	if resp.Total != 5 || len(resp.List) != 1 || *resp.List[0].RuntimeMS != 12 {
		t.Fatalf("unexpected submission list response: %+v", resp)
	}
}

func TestContestHelperFunctions(t *testing.T) {
	now := time.Now()
	if contestStatus(model.Contest{StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour)}) != "upcoming" {
		t.Fatal("expected upcoming contest")
	}
	if contestStatus(model.Contest{StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour)}) != "running" {
		t.Fatal("expected running contest")
	}
	ended := model.Contest{StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Hour), AllowPractice: true}
	if contestStatus(ended) != "ended" || !contestPracticeEnabled(ended) {
		t.Fatal("expected ended contest with practice enabled")
	}

	freezeAt := now.Add(-time.Minute)
	runningFrozen := model.Contest{StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), RanklistFreezeAt: &freezeAt}
	if !contestRanklistFrozen(runningFrozen, false) {
		t.Fatal("expected running contest ranklist to be frozen")
	}
	if contestRanklistFrozen(runningFrozen, true) {
		t.Fatal("includeFrozen should bypass freeze hiding")
	}

	if defaultContestProblemCode(0) != "A" || defaultContestProblemCode(25) != "Z" || defaultContestProblemCode(26) != "P27" {
		t.Fatal("unexpected default problem code")
	}
	if !sameContestRank(contestRankEntry{SolvedCount: 1, PenaltyMinutes: 10}, contestRankEntry{SolvedCount: 1, PenaltyMinutes: 10}) {
		t.Fatal("expected same rank for equal solved and penalty")
	}
	if sameContestRank(contestRankEntry{SolvedCount: 1}, contestRankEntry{SolvedCount: 2}) {
		t.Fatal("did not expect same rank for different solved counts")
	}
}

func TestTeachingHelperFunctions(t *testing.T) {
	if problemDifficultyString(1) != "easy" || problemDifficultyString(2) != "medium" || problemDifficultyString(3) != "hard" || problemDifficultyString(0) != "unknown" {
		t.Fatal("unexpected difficulty strings")
	}
	if !isTeacherRole("teacher") || !isTeacherRole("admin") || isTeacherRole("student") {
		t.Fatal("unexpected teacher role check")
	}

	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)
	if assignmentStatus(model.Assignment{StartAt: &future}) != "upcoming" {
		t.Fatal("expected upcoming assignment")
	}
	if assignmentStatus(model.Assignment{DueAt: &past}) != "closed" {
		t.Fatal("expected closed assignment")
	}
	if assignmentStatus(model.Assignment{StartAt: &past, DueAt: &future}) != "open" {
		t.Fatal("expected open assignment")
	}

	code := randomJoinCode()
	if len(code) != 8 {
		t.Fatalf("expected 8-character join code, got %q", code)
	}
	for _, ch := range code {
		if !strings.ContainsRune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789", ch) {
			t.Fatalf("unexpected join code character %q in %q", ch, code)
		}
	}
}

func TestBuildPlaylistProgress(t *testing.T) {
	empty := buildPlaylistProgress(nil)
	if empty.ProblemCount != 0 || empty.NextProblemID != nil {
		t.Fatalf("unexpected empty progress: %+v", empty)
	}

	progress := buildPlaylistProgress([]dto.PlaylistProblemItem{
		{ProblemID: 1, DisplayID: "A", Status: "accepted"},
		{ProblemID: 2, DisplayID: "B", Status: "attempted"},
		{ProblemID: 3, DisplayID: "C"},
	})
	if progress.ProblemCount != 3 || progress.SolvedCount != 1 || progress.AttemptedCount != 1 || progress.ProgressPercent != 33 {
		t.Fatalf("unexpected progress counts: %+v", progress)
	}
	if progress.NextProblemID == nil || *progress.NextProblemID != 2 || progress.NextProblemDisplayID != "B" {
		t.Fatalf("unexpected next problem: %+v", progress)
	}

	allSolved := buildPlaylistProgress([]dto.PlaylistProblemItem{
		{ProblemID: 7, DisplayID: "G", Status: "accepted"},
	})
	if allSolved.NextProblemID == nil || *allSolved.NextProblemID != 7 || allSolved.NextProblemDisplayID != "G" {
		t.Fatalf("expected first problem as next when all solved, got %+v", allSolved)
	}
}

func buildZip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
