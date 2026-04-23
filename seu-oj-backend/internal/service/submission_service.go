package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
)

var (
	ErrSubmissionNotFound  = errors.New("submission not found")
	ErrSubmissionForbidden = errors.New("submission forbidden")
	ErrProblemUnavailable  = errors.New("problem unavailable")
	ErrQueueEnqueueFailed  = errors.New("enqueue submission failed")
)

type SubmissionService struct {
	db                   *gorm.DB
	submissionRepo       *repository.SubmissionRepository
	submissionResultRepo *repository.SubmissionResultRepository
	problemRepo          *repository.ProblemRepository
	problemTestcaseRepo  *repository.ProblemTestcaseRepository
	judgeQueue           *queue.JudgeQueue
	sandboxRunner        *sandbox.Runner
	contestService       *ContestService
	cache                *cache.Cache
}

func NewSubmissionService(
	db *gorm.DB,
	submissionRepo *repository.SubmissionRepository,
	submissionResultRepo *repository.SubmissionResultRepository,
	problemRepo *repository.ProblemRepository,
	problemTestcaseRepo *repository.ProblemTestcaseRepository,
	judgeQueue *queue.JudgeQueue,
	sandboxRunner *sandbox.Runner,
	contestService *ContestService,
	cacheStore *cache.Cache,
) *SubmissionService {
	return &SubmissionService{
		db:                   db,
		submissionRepo:       submissionRepo,
		submissionResultRepo: submissionResultRepo,
		problemRepo:          problemRepo,
		problemTestcaseRepo:  problemTestcaseRepo,
		judgeQueue:           judgeQueue,
		sandboxRunner:        sandboxRunner,
		contestService:       contestService,
		cache:                cacheStore,
	}
}

func (s *SubmissionService) CreateSubmission(userID uint64, role string, req dto.CreateSubmissionRequest) (uint64, string, error) {
	isPractice := false
	if req.ContestID != nil {
		if err := s.contestService.ValidateSubmissionAccess(userID, role, *req.ContestID, req.ProblemID); err != nil {
			return 0, "", err
		}
		contest, err := s.contestService.getContest(*req.ContestID)
		if err != nil {
			return 0, "", err
		}
		isPractice = contestStatus(*contest) == "ended"
	} else {
		problem, err := s.problemRepo.GetByID(req.ProblemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, "", ErrProblemUnavailable
			}
			return 0, "", err
		}
		if !problem.Visible {
			return 0, "", ErrProblemUnavailable
		}
	}

	submission := model.Submission{
		UserID:     userID,
		ProblemID:  req.ProblemID,
		ContestID:  req.ContestID,
		IsPractice: isPractice,
		Language:   req.Language,
		Code:       req.Code,
		Status:     "Pending",
	}

	if err := s.submissionRepo.Create(&submission); err != nil {
		return 0, "", err
	}
	s.invalidateSubmissionCaches(userID, req.ContestID, submission.ID)

	if err := s.judgeQueue.EnqueueSubmission(context.Background(), submission.ID); err != nil {
		return 0, "", ErrQueueEnqueueFailed
	}

	return submission.ID, submission.Status, nil
}

func (s *SubmissionService) ListMySubmissions(ctx context.Context, userID uint64, page, pageSize int, problemID *uint64, contestID *uint64, status *string) (*dto.SubmissionListResponse, error) {
	if isUnfilteredRecentSubmissionQuery(page, pageSize, problemID, contestID, status) {
		return s.listRecentMySubmissions(ctx, userID, pageSize)
	}

	key := fmt.Sprintf(
		"cache:submissions:user:%d:list:p%d:s%d:problem:%s:contest:%s:status:%s",
		userID,
		page,
		pageSize,
		uintFilterKey(problemID),
		uintFilterKey(contestID),
		stringFilterKey(status),
	)
	return cache.GetOrSet(ctx, s.cache, key, 60*time.Second, func() (*dto.SubmissionListResponse, error) {
		return s.listMySubmissions(userID, page, pageSize, problemID, contestID, status)
	})
}

func (s *SubmissionService) listRecentMySubmissions(ctx context.Context, userID uint64, pageSize int) (*dto.SubmissionListResponse, error) {
	const cachedLimit = 100
	key := fmt.Sprintf("cache:submissions:user:%d:recent:limit%d", userID, cachedLimit)
	cached, err := cache.GetOrSet(ctx, s.cache, key, 60*time.Second, func() (*dto.SubmissionListResponse, error) {
		submissions, err := s.submissionRepo.ListRecentByUserID(userID, 1, cachedLimit, nil, nil, nil)
		if err != nil {
			return nil, err
		}
		return toSubmissionListResponse(submissions, int64(len(submissions)), 1, cachedLimit), nil
	})
	if err != nil {
		return nil, err
	}

	items := cached.List
	if len(items) > pageSize {
		items = items[:pageSize]
	}
	return &dto.SubmissionListResponse{
		List:     items,
		Total:    int64(len(cached.List)),
		Page:     1,
		PageSize: pageSize,
	}, nil
}

func (s *SubmissionService) listMySubmissions(userID uint64, page, pageSize int, problemID *uint64, contestID *uint64, status *string) (*dto.SubmissionListResponse, error) {
	if page == 1 && pageSize <= 20 {
		submissions, err := s.submissionRepo.ListRecentByUserID(userID, page, pageSize, problemID, contestID, status)
		if err != nil {
			return nil, err
		}
		return toSubmissionListResponse(submissions, int64(len(submissions)), page, pageSize), nil
	}

	submissions, total, err := s.submissionRepo.ListByUserID(userID, page, pageSize, problemID, contestID, status)
	if err != nil {
		return nil, err
	}
	return toSubmissionListResponse(submissions, total, page, pageSize), nil
}

func (s *SubmissionService) ListSubmissions(requestRole string, page, pageSize int, userID *uint64, problemID *uint64, contestID *uint64, status *string) (*dto.SubmissionListResponse, error) {
	if requestRole != "admin" {
		return nil, ErrPermissionDenied
	}

	submissions, total, err := s.submissionRepo.List(page, pageSize, userID, problemID, contestID, status)
	if err != nil {
		return nil, err
	}
	return toSubmissionListResponse(submissions, total, page, pageSize), nil
}

func (s *SubmissionService) GetSubmissionDetail(ctx context.Context, requestUserID uint64, requestRole string, submissionID uint64) (*dto.SubmissionDetailResponse, error) {
	key := fmt.Sprintf("cache:submissions:detail:%d", submissionID)
	result, err := cache.GetOrSet(ctx, s.cache, key, 60*time.Second, func() (*dto.SubmissionDetailResponse, error) {
		return s.getSubmissionDetail(submissionID)
	})
	if err != nil {
		return nil, err
	}
	if requestRole != "admin" && result.UserID != requestUserID {
		return nil, ErrSubmissionForbidden
	}
	return result, nil
}

func (s *SubmissionService) getSubmissionDetail(submissionID uint64) (*dto.SubmissionDetailResponse, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}

	results, err := s.submissionResultRepo.ListBySubmissionID(submission.ID)
	if err != nil {
		return nil, err
	}

	resultItems := make([]dto.SubmissionResultResponse, 0, len(results))
	for _, result := range results {
		resultItems = append(resultItems, dto.SubmissionResultResponse{
			ID:         result.ID,
			TestcaseID: result.TestcaseID,
			Status:     result.Status,
			RuntimeMS:  result.RuntimeMS,
			MemoryKB:   result.MemoryKB,
			ErrorMsg:   result.ErrorMsg,
			CreatedAt:  result.CreatedAt,
		})
	}

	return &dto.SubmissionDetailResponse{
		ID:          submission.ID,
		UserID:      submission.UserID,
		ProblemID:   submission.ProblemID,
		ContestID:   submission.ContestID,
		IsPractice:  submission.IsPractice,
		Language:    submission.Language,
		Code:        submission.Code,
		Status:      submission.Status,
		PassedCount: submission.PassedCount,
		TotalCount:  submission.TotalCount,
		RuntimeMS:   submission.RuntimeMS,
		MemoryKB:    submission.MemoryKB,
		CompileInfo: submission.CompileInfo,
		ErrorMsg:    submission.ErrorMsg,
		CreatedAt:   submission.CreatedAt,
		JudgedAt:    submission.JudgedAt,
		Results:     resultItems,
	}, nil
}

func (s *SubmissionService) RejudgeSubmission(requestRole string, submissionID uint64) (uint64, string, error) {
	if requestRole != "admin" {
		return 0, "", ErrPermissionDenied
	}

	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", ErrSubmissionNotFound
		}
		return 0, "", err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		submission.Status = "Pending"
		submission.PassedCount = 0
		submission.TotalCount = 0
		submission.RuntimeMS = nil
		submission.MemoryKB = nil
		submission.CompileInfo = ""
		submission.ErrorMsg = ""
		submission.JudgedAt = nil
		if err := s.submissionRepo.Update(tx, submission); err != nil {
			return err
		}
		return s.submissionResultRepo.ReplaceBySubmissionID(tx, submission.ID, nil)
	})
	if err != nil {
		return 0, "", err
	}
	s.invalidateSubmissionCaches(submission.UserID, submission.ContestID, submission.ID)

	if err := s.judgeQueue.EnqueueSubmission(context.Background(), submission.ID); err != nil {
		return 0, "", ErrQueueEnqueueFailed
	}

	return submission.ID, submission.Status, nil
}

func (s *SubmissionService) invalidateSubmissionCaches(userID uint64, contestID *uint64, submissionID uint64) {
	prefixes := []string{
		"cache:stats:",
		"cache:ranklist:",
		"cache:problems:stats:",
		fmt.Sprintf("cache:submissions:user:%d:", userID),
	}
	if submissionID > 0 {
		prefixes = append(prefixes, fmt.Sprintf("cache:submissions:detail:%d", submissionID))
	}
	if contestID != nil {
		prefixes = append(prefixes, "cache:contests:ranklist:")
	}
	s.cache.DeletePrefixes(context.Background(), prefixes...)
}

func isUnfilteredRecentSubmissionQuery(page, pageSize int, problemID *uint64, contestID *uint64, status *string) bool {
	return page == 1 && pageSize <= 100 && problemID == nil && contestID == nil && (status == nil || *status == "")
}

func uintFilterKey(value *uint64) string {
	if value == nil {
		return "all"
	}
	return fmt.Sprintf("%d", *value)
}

func stringFilterKey(value *string) string {
	if value == nil || *value == "" {
		return "all"
	}
	return *value
}

func toSubmissionListResponse(submissions []model.Submission, total int64, page, pageSize int) *dto.SubmissionListResponse {
	items := make([]dto.SubmissionListItem, 0, len(submissions))
	for _, submission := range submissions {
		items = append(items, dto.SubmissionListItem{
			ID:          submission.ID,
			UserID:      submission.UserID,
			ProblemID:   submission.ProblemID,
			ContestID:   submission.ContestID,
			IsPractice:  submission.IsPractice,
			Language:    submission.Language,
			Status:      submission.Status,
			PassedCount: submission.PassedCount,
			TotalCount:  submission.TotalCount,
			RuntimeMS:   submission.RuntimeMS,
			CreatedAt:   submission.CreatedAt,
			JudgedAt:    submission.JudgedAt,
		})
	}

	return &dto.SubmissionListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
