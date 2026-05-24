package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

var (
	ErrProblemPackageInvalid    = errors.New("invalid problem package")
	ErrProblemDisplayIDConflict = errors.New("problem display id already exists")
)

const maxProblemPackageBytes = 32 << 20

type TestcaseImportOptions struct {
	Replace  bool
	CaseType string
}

type TestcaseImportResult struct {
	ProblemID uint64 `json:"problem_id"`
	Imported  int    `json:"imported"`
	Replaced  bool   `json:"replaced"`
}

type ProblemImportResult struct {
	ProblemID uint64 `json:"problem_id"`
	Imported  int    `json:"imported_testcases"`
}

type problemPackageJSON struct {
	DisplayID     string `json:"display_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	InputDesc     string `json:"input_desc"`
	OutputDesc    string `json:"output_desc"`
	SampleInput   string `json:"sample_input"`
	SampleOutput  string `json:"sample_output"`
	Hint          string `json:"hint"`
	Source        string `json:"source"`
	JudgeMode     string `json:"judge_mode"`
	Difficulty    int    `json:"difficulty"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	Visible       bool   `json:"visible"`
}

type testcaseFilePair struct {
	Order  int
	Input  string
	Output string
}

func (s *ProblemService) ImportProblemTestcases(problemID uint64, reader io.Reader, opts TestcaseImportOptions) (*TestcaseImportResult, error) {
	if opts.CaseType == "" {
		opts.CaseType = "hidden"
	}
	if opts.CaseType != "sample" && opts.CaseType != "hidden" {
		return nil, ErrProblemPackageInvalid
	}
	if _, err := s.problemRepo.GetByID(problemID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	data, err := readLimitedPackage(reader)
	if err != nil {
		return nil, err
	}
	pairs, err := parseTestcaseZip(data, "")
	if err != nil {
		return nil, err
	}
	testcases := make([]model.ProblemTestcase, 0, len(pairs))
	for _, pair := range pairs {
		testcases = append(testcases, model.ProblemTestcase{
			ProblemID:  problemID,
			CaseType:   opts.CaseType,
			InputData:  pair.Input,
			OutputData: pair.Output,
			Score:      0,
			SortOrder:  pair.Order,
			IsActive:   true,
		})
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if opts.Replace {
			return s.problemTestcaseRepo.ReplaceByProblemID(tx, problemID, testcases)
		}
		return s.problemTestcaseRepo.BatchCreate(tx, testcases)
	})
	if err != nil {
		return nil, err
	}
	s.invalidate()
	return &TestcaseImportResult{ProblemID: problemID, Imported: len(testcases), Replaced: opts.Replace}, nil
}

func (s *ProblemService) ExportProblemTestcases(problemID uint64) ([]byte, string, error) {
	problem, err := s.problemRepo.GetByID(problemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrProblemNotFound
		}
		return nil, "", err
	}
	testcases, err := s.problemTestcaseRepo.ListByProblemID(problemID)
	if err != nil {
		return nil, "", err
	}
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	manifest := map[string]any{
		"problem_id": problem.ID,
		"display_id": problem.DisplayID,
		"title":      problem.Title,
		"testcases":  buildTestcaseManifest(testcases, ""),
	}
	if err := writeJSONFile(zw, "manifest.json", manifest); err != nil {
		_ = zw.Close()
		return nil, "", err
	}
	for i, testcase := range testcases {
		order := testcase.SortOrder
		if order <= 0 {
			order = i + 1
		}
		if err := writeZipFile(zw, fmt.Sprintf("%d.in", order), testcase.InputData); err != nil {
			_ = zw.Close()
			return nil, "", err
		}
		if err := writeZipFile(zw, fmt.Sprintf("%d.out", order), testcase.OutputData); err != nil {
			_ = zw.Close()
			return nil, "", err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), safePackageName(problem.DisplayID, "testcases") + ".zip", nil
}

func (s *ProblemService) ImportProblemPackage(userID uint64, role string, reader io.Reader) (*ProblemImportResult, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}
	data, err := readLimitedPackage(reader)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, ErrProblemPackageInvalid
	}
	var problemSpec problemPackageJSON
	foundProblemJSON := false
	for _, file := range zr.File {
		name, ok := cleanZipName(file.Name)
		if !ok {
			return nil, ErrProblemPackageInvalid
		}
		if name != "problem.json" {
			continue
		}
		content, err := readZipText(file)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(content), &problemSpec); err != nil {
			return nil, ErrProblemPackageInvalid
		}
		foundProblemJSON = true
		break
	}
	if !foundProblemJSON || strings.TrimSpace(problemSpec.DisplayID) == "" || strings.TrimSpace(problemSpec.Title) == "" {
		return nil, ErrProblemPackageInvalid
	}
	if problemSpec.JudgeMode == "" {
		problemSpec.JudgeMode = "standard"
	}
	pairs, err := parseTestcaseZip(data, "tests/")
	if err != nil {
		return nil, err
	}
	var problem model.Problem
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Problem{}).Where("display_id = ?", strings.TrimSpace(problemSpec.DisplayID)).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrProblemDisplayIDConflict
		}
		problem = model.Problem{
			DisplayID:     strings.TrimSpace(problemSpec.DisplayID),
			Title:         strings.TrimSpace(problemSpec.Title),
			Description:   problemSpec.Description,
			InputDesc:     problemSpec.InputDesc,
			OutputDesc:    problemSpec.OutputDesc,
			SampleInput:   problemSpec.SampleInput,
			SampleOutput:  problemSpec.SampleOutput,
			Hint:          problemSpec.Hint,
			Source:        strings.TrimSpace(problemSpec.Source),
			JudgeMode:     problemSpec.JudgeMode,
			Difficulty:    problemSpec.Difficulty,
			TimeLimitMS:   problemSpec.TimeLimitMS,
			MemoryLimitMB: problemSpec.MemoryLimitMB,
			Visible:       problemSpec.Visible,
			CreatedBy:     userID,
		}
		if problem.TimeLimitMS <= 0 {
			problem.TimeLimitMS = 1000
		}
		if problem.MemoryLimitMB <= 0 {
			problem.MemoryLimitMB = 256
		}
		if err := s.problemRepo.Create(tx, &problem); err != nil {
			return err
		}
		testcases := make([]model.ProblemTestcase, 0, len(pairs))
		for _, pair := range pairs {
			testcases = append(testcases, model.ProblemTestcase{
				ProblemID:  problem.ID,
				CaseType:   "hidden",
				InputData:  pair.Input,
				OutputData: pair.Output,
				SortOrder:  pair.Order,
				IsActive:   true,
			})
		}
		return s.problemTestcaseRepo.BatchCreate(tx, testcases)
	})
	if err != nil {
		return nil, err
	}
	s.invalidate()
	return &ProblemImportResult{ProblemID: problem.ID, Imported: len(pairs)}, nil
}

func (s *ProblemService) ExportProblemPackage(problemID uint64) ([]byte, string, error) {
	problem, err := s.problemRepo.GetByID(problemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrProblemNotFound
		}
		return nil, "", err
	}
	testcases, err := s.problemTestcaseRepo.ListByProblemID(problemID)
	if err != nil {
		return nil, "", err
	}
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	spec := problemPackageJSON{
		DisplayID:     problem.DisplayID,
		Title:         problem.Title,
		Description:   problem.Description,
		InputDesc:     problem.InputDesc,
		OutputDesc:    problem.OutputDesc,
		SampleInput:   problem.SampleInput,
		SampleOutput:  problem.SampleOutput,
		Hint:          problem.Hint,
		Source:        problem.Source,
		JudgeMode:     problem.JudgeMode,
		Difficulty:    problem.Difficulty,
		TimeLimitMS:   problem.TimeLimitMS,
		MemoryLimitMB: problem.MemoryLimitMB,
		Visible:       problem.Visible,
	}
	if err := writeJSONFile(zw, "problem.json", spec); err != nil {
		_ = zw.Close()
		return nil, "", err
	}
	for i, testcase := range testcases {
		order := testcase.SortOrder
		if order <= 0 {
			order = i + 1
		}
		if err := writeZipFile(zw, fmt.Sprintf("tests/%d.in", order), testcase.InputData); err != nil {
			_ = zw.Close()
			return nil, "", err
		}
		if err := writeZipFile(zw, fmt.Sprintf("tests/%d.out", order), testcase.OutputData); err != nil {
			_ = zw.Close()
			return nil, "", err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), safePackageName(problem.DisplayID, "problem") + ".zip", nil
}

func (s *ProblemService) PublicProblemDetail(id uint64) (*dto.ProblemDetailResponse, error) {
	return s.GetProblemDetail(id)
}

func readLimitedPackage(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxProblemPackageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > maxProblemPackageBytes {
		return nil, ErrProblemPackageInvalid
	}
	return data, nil
}

func parseTestcaseZip(data []byte, prefix string) ([]testcaseFilePair, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, ErrProblemPackageInvalid
	}
	inputs := map[int]string{}
	outputs := map[int]string{}
	for _, file := range zr.File {
		name, ok := cleanZipName(file.Name)
		if !ok {
			return nil, ErrProblemPackageInvalid
		}
		if file.FileInfo().IsDir() || name == "manifest.json" || name == "problem.json" {
			continue
		}
		if prefix != "" {
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			name = strings.TrimPrefix(name, prefix)
		}
		if strings.Contains(name, "/") {
			continue
		}
		ext := path.Ext(name)
		if ext != ".in" && ext != ".out" {
			continue
		}
		order, err := strconv.Atoi(strings.TrimSuffix(name, ext))
		if err != nil || order <= 0 {
			return nil, ErrProblemPackageInvalid
		}
		content, err := readZipText(file)
		if err != nil {
			return nil, err
		}
		if ext == ".in" {
			inputs[order] = content
		} else {
			outputs[order] = content
		}
	}
	orders := make([]int, 0, len(inputs))
	for order := range inputs {
		if _, ok := outputs[order]; !ok {
			return nil, ErrProblemPackageInvalid
		}
		orders = append(orders, order)
	}
	if len(orders) == 0 || len(outputs) != len(inputs) {
		return nil, ErrProblemPackageInvalid
	}
	sort.Ints(orders)
	pairs := make([]testcaseFilePair, 0, len(orders))
	for _, order := range orders {
		pairs = append(pairs, testcaseFilePair{Order: order, Input: inputs[order], Output: outputs[order]})
	}
	return pairs, nil
}

func cleanZipName(name string) (string, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "/")
	cleaned := path.Clean(name)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." || path.IsAbs(cleaned) {
		return "", false
	}
	return cleaned, true
}

func readZipText(file *zip.File) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, maxProblemPackageBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxProblemPackageBytes {
		return "", ErrProblemPackageInvalid
	}
	return string(data), nil
}

func writeZipFile(zw *zip.Writer, name string, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func writeJSONFile(zw *zip.Writer, name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(append(data, '\n'))
	return err
}

func buildTestcaseManifest(testcases []model.ProblemTestcase, prefix string) []map[string]any {
	items := make([]map[string]any, 0, len(testcases))
	for i, testcase := range testcases {
		order := testcase.SortOrder
		if order <= 0 {
			order = i + 1
		}
		items = append(items, map[string]any{
			"sort_order":  order,
			"case_type":   testcase.CaseType,
			"score":       testcase.Score,
			"is_active":   testcase.IsActive,
			"input_file":  fmt.Sprintf("%s%d.in", prefix, order),
			"output_file": fmt.Sprintf("%s%d.out", prefix, order),
		})
	}
	return items
}

func safePackageName(displayID, fallback string) string {
	displayID = strings.TrimSpace(displayID)
	if displayID == "" {
		return fallback
	}
	var b strings.Builder
	for _, ch := range displayID {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z', ch >= '0' && ch <= '9', ch == '-', ch == '_':
			b.WriteRune(ch)
		}
	}
	if b.Len() == 0 {
		return fallback
	}
	return b.String()
}
