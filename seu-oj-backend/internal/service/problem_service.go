package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/repository"
)

var (
	ErrProblemNotFound            = errors.New("problem not found")
	ErrPermissionDenied           = errors.New("permission denied")
	ErrSolutionNotFound           = errors.New("solution not found")
	ErrSolutionPublishNotAllowed  = errors.New("solution publish not allowed")
	ErrInvalidProblemSolutionData = errors.New("invalid problem solution data")
)

type ProblemService struct {
	db                  *gorm.DB
	problemRepo         *repository.ProblemRepository
	problemTestcaseRepo *repository.ProblemTestcaseRepository
	cache               *cache.Cache
}

func NewProblemService(
	db *gorm.DB,
	problemRepo *repository.ProblemRepository,
	problemTestcaseRepo *repository.ProblemTestcaseRepository,
	cacheStore *cache.Cache,
) *ProblemService {
	return &ProblemService{
		db:                  db,
		problemRepo:         problemRepo,
		problemTestcaseRepo: problemTestcaseRepo,
		cache:               cacheStore,
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
		Difficulty:    req.Difficulty,
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

	s.invalidate()
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
			Difficulty:    problem.Difficulty,
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

func (s *ProblemService) ListProblems(ctx context.Context, page, pageSize int, keyword string) (*dto.ProblemListResponse, error) {
	if page == 1 && pageSize <= 100 && strings.TrimSpace(keyword) == "" {
		return s.listRecentProblems(ctx, pageSize)
	}

	key := fmt.Sprintf("cache:problems:list:p%d:s%d:k%s", page, pageSize, strings.TrimSpace(keyword))
	return cache.GetOrSet(ctx, s.cache, key, 60*time.Second, func() (*dto.ProblemListResponse, error) {
		return s.listProblems(page, pageSize, keyword)
	})
}

func (s *ProblemService) listRecentProblems(ctx context.Context, pageSize int) (*dto.ProblemListResponse, error) {
	const cachedLimit = 100
	key := fmt.Sprintf("cache:problems:list:recent:limit%d", cachedLimit)
	cached, err := cache.GetOrSet(ctx, s.cache, key, 60*time.Second, func() (*dto.ProblemListResponse, error) {
		return s.listProblems(1, cachedLimit, "")
	})
	if err != nil {
		return nil, err
	}

	items := cached.List
	if len(items) > pageSize {
		items = items[:pageSize]
	}
	return &dto.ProblemListResponse{
		List:     items,
		Total:    cached.Total,
		Page:     1,
		PageSize: pageSize,
	}, nil
}

func (s *ProblemService) listProblems(page, pageSize int, keyword string) (*dto.ProblemListResponse, error) {
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
			Difficulty:    problem.Difficulty,
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
	key := fmt.Sprintf("cache:problems:detail:%d", id)
	return cache.GetOrSet(context.Background(), s.cache, key, 60*time.Second, func() (*dto.ProblemDetailResponse, error) {
		return s.getProblemDetail(id)
	})
}

func (s *ProblemService) getProblemDetail(id uint64) (*dto.ProblemDetailResponse, error) {
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
	solutions, err := s.ListProblemSolutions(id, false)
	if err != nil {
		return nil, err
	}

	resp := toProblemDetailResponse(*problem, filterSampleTestcases(testcases), solutions)
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
	solutions, err := s.ListProblemSolutions(id, true)
	if err != nil {
		return nil, err
	}

	resp := toProblemDetailResponse(*problem, testcases, solutions)
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
	problem.Difficulty = req.Difficulty
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

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.problemRepo.Update(tx, problem); err != nil {
			return err
		}
		return s.problemTestcaseRepo.ReplaceByProblemID(tx, problem.ID, testcases)
	}); err != nil {
		return err
	}
	s.invalidate()
	return nil
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

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_id = ?", id).Delete(&model.ProblemSolution{}).Error; err != nil {
			return err
		}
		return s.problemRepo.Delete(tx, id)
	}); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

func (s *ProblemService) ListProblemSolutions(problemID uint64, includePrivate bool) ([]dto.ProblemSolutionResponse, error) {
	problem, err := s.problemRepo.GetByID(problemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	if !includePrivate && !problem.Visible {
		return nil, ErrProblemNotFound
	}

	query := s.db.Model(&model.ProblemSolution{}).Where("problem_id = ?", problemID)
	if !includePrivate {
		query = query.Where("visibility = ?", "public")
	}
	var items []model.ProblemSolution
	if err := query.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]dto.ProblemSolutionResponse, 0, len(items))
	for _, item := range items {
		result = append(result, dto.ProblemSolutionResponse{
			ID:         item.ID,
			ProblemID:  item.ProblemID,
			Title:      item.Title,
			Content:    item.Content,
			Visibility: item.Visibility,
			AuthorID:   item.AuthorID,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}
	return result, nil
}

func (s *ProblemService) ListManageProblemSolutions(userID uint64, role string, problemID uint64) ([]dto.ProblemSolutionResponse, error) {
	problem, err := s.problemRepo.GetByID(problemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	if role != "admin" && !problem.Visible {
		return nil, ErrProblemNotFound
	}

	query := s.db.Model(&model.ProblemSolution{}).Where("problem_id = ?", problemID)
	if role != "admin" {
		query = query.Where("author_id = ?", userID)
	}
	var items []model.ProblemSolution
	if err := query.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return toProblemSolutionResponses(items), nil
}

func (s *ProblemService) CreateProblemSolution(userID uint64, role string, problemID uint64, req dto.CreateProblemSolutionRequest) (*dto.ProblemSolutionResponse, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || strings.TrimSpace(req.Content) == "" {
		return nil, ErrInvalidProblemSolutionData
	}
	problem, err := s.problemRepo.GetByID(problemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	if role != "admin" && !problem.Visible {
		return nil, ErrProblemNotFound
	}
	if role != "admin" {
		ok, err := s.hasAcceptedProblem(userID, problemID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrSolutionPublishNotAllowed
		}
	}
	solution := model.ProblemSolution{
		ProblemID:  problemID,
		Title:      title,
		Content:    req.Content,
		Visibility: req.Visibility,
		AuthorID:   userID,
	}
	if err := s.db.Create(&solution).Error; err != nil {
		return nil, err
	}
	s.invalidate()
	resp := dto.ProblemSolutionResponse{ID: solution.ID, ProblemID: solution.ProblemID, Title: solution.Title, Content: solution.Content, Visibility: solution.Visibility, AuthorID: solution.AuthorID, CreatedAt: solution.CreatedAt, UpdatedAt: solution.UpdatedAt}
	return &resp, nil
}

func (s *ProblemService) UpdateProblemSolution(userID uint64, role string, problemID, solutionID uint64, req dto.CreateProblemSolutionRequest) (*dto.ProblemSolutionResponse, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || strings.TrimSpace(req.Content) == "" {
		return nil, ErrInvalidProblemSolutionData
	}
	var solution model.ProblemSolution
	if err := s.db.Where("id = ? AND problem_id = ?", solutionID, problemID).First(&solution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSolutionNotFound
		}
		return nil, err
	}
	if role != "admin" && solution.AuthorID != userID {
		return nil, ErrPermissionDenied
	}
	solution.Title = title
	solution.Content = req.Content
	solution.Visibility = req.Visibility
	if err := s.db.Save(&solution).Error; err != nil {
		return nil, err
	}
	s.invalidate()
	resp := dto.ProblemSolutionResponse{ID: solution.ID, ProblemID: solution.ProblemID, Title: solution.Title, Content: solution.Content, Visibility: solution.Visibility, AuthorID: solution.AuthorID, CreatedAt: solution.CreatedAt, UpdatedAt: solution.UpdatedAt}
	return &resp, nil
}

func (s *ProblemService) DeleteProblemSolution(userID uint64, role string, problemID, solutionID uint64) error {
	var solution model.ProblemSolution
	if err := s.db.Where("id = ? AND problem_id = ?", solutionID, problemID).First(&solution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSolutionNotFound
		}
		return err
	}
	if role != "admin" && solution.AuthorID != userID {
		return ErrPermissionDenied
	}
	if err := s.db.Delete(&solution).Error; err != nil {
		return err
	}
	s.invalidate()
	return nil
}

func (s *ProblemService) hasAcceptedProblem(userID, problemID uint64) (bool, error) {
	var count int64
	if err := s.db.Model(&model.Submission{}).
		Where("user_id = ? AND problem_id = ? AND status = ?", userID, problemID, "Accepted").
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func toProblemSolutionResponses(items []model.ProblemSolution) []dto.ProblemSolutionResponse {
	result := make([]dto.ProblemSolutionResponse, 0, len(items))
	for _, item := range items {
		result = append(result, dto.ProblemSolutionResponse{
			ID:         item.ID,
			ProblemID:  item.ProblemID,
			Title:      item.Title,
			Content:    item.Content,
			Visibility: item.Visibility,
			AuthorID:   item.AuthorID,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}
	return result
}

func toProblemDetailResponse(problem model.Problem, testcases []model.ProblemTestcase, solutions []dto.ProblemSolutionResponse) dto.ProblemDetailResponse {
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
		Difficulty:    problem.Difficulty,
		TimeLimitMS:   problem.TimeLimitMS,
		MemoryLimitMB: problem.MemoryLimitMB,
		Visible:       problem.Visible,
		CreatedBy:     problem.CreatedBy,
		CreatedAt:     problem.CreatedAt,
		UpdatedAt:     problem.UpdatedAt,
		Testcases:     respTestcases,
		Solutions:     solutions,
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

func (s *ProblemService) GetProblemStats(id uint64) (*dto.ProblemStatsResponse, error) {
	key := fmt.Sprintf("cache:problems:stats:%d", id)
	return cache.GetOrSet(context.Background(), s.cache, key, 15*time.Second, func() (*dto.ProblemStatsResponse, error) {
		return s.getProblemStats(id)
	})
}

func (s *ProblemService) getProblemStats(id uint64) (*dto.ProblemStatsResponse, error) {
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

	resp := &dto.ProblemStatsResponse{ProblemID: id}

	if err := s.db.Model(&model.Submission{}).Where("problem_id = ?", id).Count(&resp.SubmissionsTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("problem_id = ? AND status = ?", id, "Accepted").Count(&resp.AcceptedSubmissions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Distinct("user_id").Where("problem_id = ? AND status = ?", id, "Accepted").Count(&resp.AcceptedUsers).Error; err != nil {
		return nil, err
	}
	if resp.SubmissionsTotal > 0 {
		resp.AcceptedRate = float64(resp.AcceptedSubmissions) / float64(resp.SubmissionsTotal)
	}
	if err := s.db.Model(&model.Submission{}).
		Select("language AS name, COUNT(*) AS count").
		Where("problem_id = ?", id).
		Group("language").
		Order("count DESC").
		Scan(&resp.LanguageBreakdown).Error; err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *ProblemService) invalidate() {
	s.cache.DeletePrefixes(context.Background(), "cache:problems:", "cache:stats:", "cache:ranklist:", "cache:contests:")
}
