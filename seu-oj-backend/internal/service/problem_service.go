package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/repository"
)

var (
	ErrProblemNotFound  = errors.New("problem not found")
	ErrPermissionDenied = errors.New("permission denied")
)

type ProblemService struct {
	db                  *gorm.DB
	problemRepo         *repository.ProblemRepository
	problemTestcaseRepo *repository.ProblemTestcaseRepository
}

func NewProblemService(
	db *gorm.DB,
	problemRepo *repository.ProblemRepository,
	problemTestcaseRepo *repository.ProblemTestcaseRepository,
) *ProblemService {
	return &ProblemService{
		db:                  db,
		problemRepo:         problemRepo,
		problemTestcaseRepo: problemTestcaseRepo,
	}
}

func (s *ProblemService) CreateProblem(ctxUserID uint64, ctxRole string, req dto.CreateProblemRequest) (uint64, error) {
	if ctxRole != "admin" {
		return 0, ErrPermissionDenied
	}

	problem := model.Problem{
		DisplayID:     strings.TrimSpace(req.DisplayID),
		Title:         strings.TrimSpace(req.Title),
		Description:   req.Description,
		InputDesc:     req.InputDesc,
		OutputDesc:    req.OutputDesc,
		SampleInput:   req.SampleInput,
		SampleOutput:  req.SampleOutput,
		Hint:          req.Hint,
		Source:        strings.TrimSpace(req.Source),
		JudgeMode:     req.JudgeMode,
		TimeLimitMS:   req.TimeLimitMS,
		MemoryLimitMB: req.MemoryLimitMB,
		Visible:       req.Visible,
		CreatedBy:     ctxUserID,
	}

	testcases := make([]model.ProblemTestcase, 0, len(req.Testcases))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.problemRepo.Create(tx, &problem); err != nil {
			return err
		}

		for _, item := range req.Testcases {
			testcases = append(testcases, model.ProblemTestcase{
				ProblemID:  problem.ID,
				CaseType:   item.CaseType,
				InputData:  item.InputData,
				OutputData: item.OutputData,
				Score:      item.Score,
				SortOrder:  item.SortOrder,
				IsActive:   item.IsActive,
			})
		}

		if len(testcases) > 0 {
			if err := s.problemTestcaseRepo.BatchCreate(tx, testcases); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return problem.ID, nil
}

func (s *ProblemService) ListAdminProblems(ctxRole string, page, pageSize int, keyword string, includeHidden bool) (*dto.ProblemListResponse, error) {
	if ctxRole != "admin" {
		return nil, ErrPermissionDenied
	}

	problems, total, err := s.problemRepo.List(page, pageSize, keyword, includeHidden)
	if err != nil {
		return nil, err
	}

	items := make([]dto.ProblemListItem, 0, len(problems))
	for _, problem := range problems {
		items = append(items, dto.ProblemListItem{
			ID:            problem.ID,
			DisplayID:     problem.DisplayID,
			Title:         problem.Title,
			JudgeMode:     problem.JudgeMode,
			TimeLimitMS:   problem.TimeLimitMS,
			MemoryLimitMB: problem.MemoryLimitMB,
			Visible:       problem.Visible,
			CreatedAt:     problem.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &dto.ProblemListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProblemService) ListProblems(page, pageSize int, keyword string) (*dto.ProblemListResponse, error) {
	problems, total, err := s.problemRepo.ListVisible(page, pageSize, keyword)
	if err != nil {
		return nil, err
	}

	items := make([]dto.ProblemListItem, 0, len(problems))
	for _, problem := range problems {
		items = append(items, dto.ProblemListItem{
			ID:            problem.ID,
			DisplayID:     problem.DisplayID,
			Title:         problem.Title,
			JudgeMode:     problem.JudgeMode,
			TimeLimitMS:   problem.TimeLimitMS,
			MemoryLimitMB: problem.MemoryLimitMB,
			Visible:       problem.Visible,
			CreatedAt:     problem.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &dto.ProblemListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProblemService) GetProblemDetail(id uint64) (*dto.ProblemDetailResponse, error) {
	problem, err := s.problemRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	if !problem.Visible {
		return nil, ErrProblemNotFound
	}

	testcases, err := s.problemTestcaseRepo.ListByProblemID(problem.ID)
	if err != nil {
		return nil, err
	}

	resp := toProblemDetailResponse(*problem, filterSampleTestcases(testcases))
	return &resp, nil
}

func (s *ProblemService) GetAdminProblemDetail(ctxRole string, id uint64) (*dto.ProblemDetailResponse, error) {
	if ctxRole != "admin" {
		return nil, ErrPermissionDenied
	}

	problem, err := s.problemRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	testcases, err := s.problemTestcaseRepo.ListByProblemID(problem.ID)
	if err != nil {
		return nil, err
	}

	resp := toProblemDetailResponse(*problem, testcases)
	return &resp, nil
}

func (s *ProblemService) UpdateProblem(ctxRole string, id uint64, req dto.CreateProblemRequest) error {
	if ctxRole != "admin" {
		return ErrPermissionDenied
	}

	problem, err := s.problemRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProblemNotFound
		}
		return err
	}

	problem.DisplayID = strings.TrimSpace(req.DisplayID)
	problem.Title = strings.TrimSpace(req.Title)
	problem.Description = req.Description
	problem.InputDesc = req.InputDesc
	problem.OutputDesc = req.OutputDesc
	problem.SampleInput = req.SampleInput
	problem.SampleOutput = req.SampleOutput
	problem.Hint = req.Hint
	problem.Source = strings.TrimSpace(req.Source)
	problem.JudgeMode = req.JudgeMode
	problem.TimeLimitMS = req.TimeLimitMS
	problem.MemoryLimitMB = req.MemoryLimitMB
	problem.Visible = req.Visible

	testcases := make([]model.ProblemTestcase, 0, len(req.Testcases))
	for _, item := range req.Testcases {
		testcases = append(testcases, model.ProblemTestcase{
			ProblemID:  problem.ID,
			CaseType:   item.CaseType,
			InputData:  item.InputData,
			OutputData: item.OutputData,
			Score:      item.Score,
			SortOrder:  item.SortOrder,
			IsActive:   item.IsActive,
		})
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.problemRepo.Update(tx, problem); err != nil {
			return err
		}
		return s.problemTestcaseRepo.ReplaceByProblemID(tx, problem.ID, testcases)
	})
}

func (s *ProblemService) DeleteProblem(ctxRole string, id uint64) error {
	if ctxRole != "admin" {
		return ErrPermissionDenied
	}

	if _, err := s.problemRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProblemNotFound
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.problemRepo.Delete(tx, id)
	})
}

func toProblemDetailResponse(problem model.Problem, testcases []model.ProblemTestcase) dto.ProblemDetailResponse {
	respTestcases := make([]dto.ProblemTestcaseResponse, 0, len(testcases))
	for _, testcase := range testcases {
		respTestcases = append(respTestcases, dto.ProblemTestcaseResponse{
			ID:         testcase.ID,
			CaseType:   testcase.CaseType,
			InputData:  testcase.InputData,
			OutputData: testcase.OutputData,
			Score:      testcase.Score,
			SortOrder:  testcase.SortOrder,
			IsActive:   testcase.IsActive,
		})
	}

	return dto.ProblemDetailResponse{
		ID:            problem.ID,
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
		TimeLimitMS:   problem.TimeLimitMS,
		MemoryLimitMB: problem.MemoryLimitMB,
		Visible:       problem.Visible,
		CreatedBy:     problem.CreatedBy,
		CreatedAt:     problem.CreatedAt,
		UpdatedAt:     problem.UpdatedAt,
		Testcases:     respTestcases,
	}
}

func filterSampleTestcases(testcases []model.ProblemTestcase) []model.ProblemTestcase {
	filtered := make([]model.ProblemTestcase, 0)
	for _, testcase := range testcases {
		if testcase.CaseType == "sample" && testcase.IsActive {
			filtered = append(filtered, testcase)
		}
	}
	return filtered
}
