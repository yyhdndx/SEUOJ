package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"seu-oj-backend/internal/middleware"
	"seu-oj-backend/internal/testutil"
)

func buildTestZip(t *testing.T, files map[string]string) []byte {
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

func callMultipartHandler(t *testing.T, method, path, fieldName, filename string, fileData []byte, form map[string]string, params gin.Params, userID uint64, role string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range form {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatalf("write field %s: %v", k, err)
		}
	}
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(fileData); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Params = params
	if userID > 0 {
		c.Set(middleware.ContextUserIDKey, userID)
		c.Set(middleware.ContextRoleKey, role)
	}
	handler(c)
	return w
}

func TestAPIProblemPackageHandlers(t *testing.T) {
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "PKG-API", "title": "Package", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 1000, "memory_limit_mb": 128, "visible": true,
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

	packageZip := buildTestZip(t, map[string]string{
		"problem.json": `{"display_id":"PKG-IMP","title":"Imported","description":"d","judge_mode":"standard","visible":true}`,
		"tests/1.in":   "1\n",
		"tests/1.out":  "2\n",
	})
	w = callMultipartHandler(t, http.MethodPost, "/admin/problems/import", "file", "pkg.zip", packageZip, nil, nil, env.adminID, "admin", env.problem.ImportProblemPackage)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatalf("import package failed: %s", w.Body.String())
	}

	testcaseZip := buildTestZip(t, map[string]string{"1.in": "in", "1.out": "out"})
	w = callMultipartHandler(t, http.MethodPost, "/admin/problems/:id/testcases/import", "file", "cases.zip", testcaseZip,
		map[string]string{"replace": "true", "case_type": "sample"},
		gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.ImportTestcases)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("import testcases failed")
	}

	w = callHandler(t, http.MethodGet, "/admin/problems/:id/testcases/export", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.ExportTestcases)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/zip" || len(w.Body.Bytes()) == 0 {
		t.Fatalf("export testcases: code=%d type=%q len=%d", w.Code, w.Header().Get("Content-Type"), len(w.Body.Bytes()))
	}

	w = callHandler(t, http.MethodGet, "/admin/problems/:id/export", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.ExportProblemPackage)
	if w.Code != http.StatusOK || len(w.Body.Bytes()) == 0 {
		t.Fatalf("export package: code=%d len=%d", w.Code, len(w.Body.Bytes()))
	}

	w = callHandler(t, http.MethodPost, "/teacher/problems/:id/solutions", map[string]any{
		"title": "Sol", "content": "hint", "visibility": "public",
	}, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.CreateSolution)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("create solution failed")
	}
	w = callHandler(t, http.MethodGet, "/teacher/problems/:id/solutions", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.teacherID, "teacher", env.problem.TeacherSolutionList)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("teacher solution list failed")
	}

	w = callHandler(t, http.MethodDelete, "/admin/problems/:id", nil, gin.Params{{Key: "id", Value: fmt.Sprint(problemID.ProblemID)}}, env.adminID, "admin", env.problem.Delete)
	if decodeAPIEnvelope(t, w).Code != 0 {
		t.Fatal("delete problem failed")
	}

	w = callMultipartHandler(t, http.MethodPost, "/admin/problems/:id/testcases/import", "file", "bad.zip", []byte("not-a-zip"), nil,
		gin.Params{{Key: "id", Value: "999"}}, env.adminID, "admin", env.problem.ImportTestcases)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected import error for invalid id/zip")
	}
	w = callHandler(t, http.MethodPost, "/admin/problems/:id/testcases/import", nil, gin.Params{{Key: "id", Value: "bad"}}, env.adminID, "admin", env.problem.ImportTestcases)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected missing file error")
	}
}

func TestAPISubmissionRunHandler(t *testing.T) {
	testutil.SkipUnlessDocker(t)
	env := newAPITestEnv(t)

	w := callHandler(t, http.MethodPost, "/admin/problems", map[string]any{
		"display_id": "RUN-API", "title": "Run", "description": "d", "input_desc": "i", "output_desc": "o",
		"judge_mode": "standard", "difficulty": 1, "time_limit_ms": 2000, "memory_limit_mb": 128, "visible": true,
		"testcases": []map[string]any{
			{"case_type": "sample", "input_data": "2 3", "output_data": "5", "sort_order": 1, "is_active": true},
		},
	}, nil, env.adminID, "admin", env.problem.Create)
	prob := decodeAPIEnvelope(t, w)
	var problemID struct {
		ProblemID uint64 `json:"problem_id"`
	}
	_ = json.Unmarshal(prob.Data, &problemID)

	code := "a, b = map(int, input().split())\nprint(a + b)\n"
	w = callHandler(t, http.MethodPost, "/submissions/run", map[string]any{
		"problem_id": problemID.ProblemID, "language": "python3", "code": code,
	}, nil, env.studentID, "student", env.submission.Run)
	run := decodeAPIEnvelope(t, w)
	if run.Code != 0 {
		t.Fatalf("run sample tests: %+v", run)
	}

	w = callHandler(t, http.MethodPost, "/submissions/run", map[string]any{}, nil, env.studentID, "student", env.submission.Run)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected invalid run request")
	}
}

func TestAPISubmissionRunNoDocker(t *testing.T) {
	if testutil.DockerAvailable() {
		t.Skip("covered by docker run handler test")
	}
	env := newAPITestEnv(t)
	w := callHandler(t, http.MethodPost, "/submissions/run", map[string]any{
		"problem_id": 99999, "language": "cpp", "code": "int main(){}",
	}, nil, env.studentID, "student", env.submission.Run)
	if decodeAPIEnvelope(t, w).Code == 0 {
		t.Fatal("expected run failure for missing problem")
	}
}
