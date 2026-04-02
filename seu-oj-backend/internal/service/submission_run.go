package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

func (s *SubmissionService) RunSampleTests(req dto.RunSubmissionRequest) (*dto.RunSubmissionResponse, error) {
	problem, err := s.problemRepo.GetByID(req.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemUnavailable
		}
		return nil, err
	}
	if !problem.Visible {
		return nil, ErrProblemUnavailable
	}

	testcases, err := s.problemTestcaseRepo.ListByProblemID(problem.ID)
	if err != nil {
		return nil, err
	}

	sampleCases := make([]model.ProblemTestcase, 0)
	for _, testcase := range testcases {
		if testcase.IsActive && testcase.CaseType == "sample" {
			sampleCases = append(sampleCases, testcase)
		}
	}
	if len(sampleCases) == 0 {
		return &dto.RunSubmissionResponse{
			Status:   "System Error",
			ErrorMsg: "no sample testcases",
			Results:  []dto.RunSubmissionCaseResult{},
		}, nil
	}

	compileResult, err := s.sandboxRunner.Compile(req.Language, req.Code)
	if err != nil {
		return &dto.RunSubmissionResponse{
			Status:   "System Error",
			ErrorMsg: err.Error(),
			Results:  []dto.RunSubmissionCaseResult{},
		}, nil
	}
	defer s.sandboxRunner.Cleanup(compileResult.Program)

	if compileResult.CompileError != "" {
		return &dto.RunSubmissionResponse{
			Status:      "Compile Error",
			CompileInfo: compileResult.CompileError,
			Results:     []dto.RunSubmissionCaseResult{},
		}, nil
	}

	resp := &dto.RunSubmissionResponse{
		Status:  "Accepted",
		Results: make([]dto.RunSubmissionCaseResult, 0, len(sampleCases)),
	}

	for _, testcase := range sampleCases {
		runResult := s.sandboxRunner.Run(compileResult.Program, testcase.InputData, problem.TimeLimitMS)
		if runResult.Status == "Accepted" && !compareOutput(runResult.Output, testcase.OutputData) {
			runResult.Status = "Wrong Answer"
			runResult.ErrorMsg = "output mismatch"
		}
		if resp.Status == "Accepted" && runResult.Status != "Accepted" {
			resp.Status = runResult.Status
			resp.ErrorMsg = runResult.ErrorMsg
		}

		resp.Results = append(resp.Results, dto.RunSubmissionCaseResult{
			CaseType:       testcase.CaseType,
			SortOrder:      testcase.SortOrder,
			Status:         runResult.Status,
			RuntimeMS:      runResult.RuntimeMS,
			InputData:      testcase.InputData,
			ExpectedOutput: testcase.OutputData,
			ActualOutput:   runResult.Output,
			ErrorMessage:   runResult.ErrorMsg,
		})
	}

	return resp, nil
}

func compareOutput(actual, expected string) bool {
	return normalizeOutput(actual) == normalizeOutput(expected)
}

func normalizeOutput(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
