package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

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
	submissionRepo       *repository.SubmissionRepository
	submissionResultRepo *repository.SubmissionResultRepository
	problemRepo          *repository.ProblemRepository
	problemTestcaseRepo  *repository.ProblemTestcaseRepository
	judgeQueue           *queue.JudgeQueue
	sandboxRunner        *sandbox.Runner
}

func NewSubmissionService(
	submissionRepo *repository.SubmissionRepository,
	submissionResultRepo *repository.SubmissionResultRepository,
	problemRepo *repository.ProblemRepository,
	problemTestcaseRepo *repository.ProblemTestcaseRepository,
	judgeQueue *queue.JudgeQueue,
	sandboxRunner *sandbox.Runner,
) *SubmissionService {
	return &SubmissionService{
		submissionRepo:       submissionRepo,
		submissionResultRepo: submissionResultRepo,
		problemRepo:          problemRepo,
		problemTestcaseRepo:  problemTestcaseRepo,
		judgeQueue:           judgeQueue,
		sandboxRunner:        sandboxRunner,
	}
}

func (s *SubmissionService) CreateSubmission(userID uint64, req dto.CreateSubmissionRequest) (uint64, string, error) {
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

	submission := model.Submission{
		UserID:    userID,
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
		Status:    "Pending",
	}

	if err := s.submissionRepo.Create(&submission); err != nil {
		return 0, "", err
	}

	if err := s.judgeQueue.EnqueueSubmission(context.Background(), submission.ID); err != nil {
		return 0, "", ErrQueueEnqueueFailed
	}

	return submission.ID, submission.Status, nil
}

func (s *SubmissionService) ListMySubmissions(userID uint64, page, pageSize int, problemID *uint64, status *string) (*dto.SubmissionListResponse, error) {
	submissions, total, err := s.submissionRepo.ListByUserID(userID, page, pageSize, problemID, status)
	if err != nil {
		return nil, err
	}

	items := make([]dto.SubmissionListItem, 0, len(submissions))
	for _, submission := range submissions {
		items = append(items, dto.SubmissionListItem{
			ID:          submission.ID,
			ProblemID:   submission.ProblemID,
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
	}, nil
}

func (s *SubmissionService) GetSubmissionDetail(requestUserID uint64, requestRole string, submissionID uint64) (*dto.SubmissionDetailResponse, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}
	if requestRole != "admin" && submission.UserID != requestUserID {
		return nil, ErrSubmissionForbidden
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
