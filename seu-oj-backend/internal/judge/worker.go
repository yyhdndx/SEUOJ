package judge

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
)

type Worker struct {
	db                   *gorm.DB
	judgeQueue           *queue.JudgeQueue
	problemRepo          *repository.ProblemRepository
	problemTestcaseRepo  *repository.ProblemTestcaseRepository
	submissionRepo       *repository.SubmissionRepository
	submissionResultRepo *repository.SubmissionResultRepository
	sandboxRunner        *sandbox.Runner
}

func NewWorker(
	db *gorm.DB,
	judgeQueue *queue.JudgeQueue,
	problemRepo *repository.ProblemRepository,
	problemTestcaseRepo *repository.ProblemTestcaseRepository,
	submissionRepo *repository.SubmissionRepository,
	submissionResultRepo *repository.SubmissionResultRepository,
	sandboxRunner *sandbox.Runner,
) *Worker {
	return &Worker{
		db:                   db,
		judgeQueue:           judgeQueue,
		problemRepo:          problemRepo,
		problemTestcaseRepo:  problemTestcaseRepo,
		submissionRepo:       submissionRepo,
		submissionResultRepo: submissionResultRepo,
		sandboxRunner:        sandboxRunner,
	}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		log.Printf("[judge-worker] started, waiting for submissions from %q", queue.JudgeQueueKey)
		for {
			select {
			case <-ctx.Done():
				log.Printf("[judge-worker] stopped")
				return
			default:
			}

			submissionID, err := w.judgeQueue.DequeueSubmission(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Printf("[judge-worker] context canceled")
					return
				}
				log.Printf("[judge-worker] dequeue failed: %v", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			log.Printf("[judge-worker] dequeued submission=%d", submissionID)
			_ = w.handleSubmission(ctx, submissionID)
		}
	}()
}

func (w *Worker) handleSubmission(ctx context.Context, submissionID uint64) error {
	startedAt := time.Now()
	submission, err := w.submissionRepo.GetByID(submissionID)
	if err != nil {
		log.Printf("[judge-worker] load submission=%d failed: %v", submissionID, err)
		return err
	}
	if submission.Status != "Pending" {
		log.Printf("[judge-worker] skip submission=%d because status=%s", submission.ID, submission.Status)
		return nil
	}

	problem, err := w.problemRepo.GetByID(submission.ProblemID)
	if err != nil {
		log.Printf("[judge-worker] submission=%d problem=%d not found", submission.ID, submission.ProblemID)
		return w.markSystemError(submission, "problem not found")
	}

	testcases, err := w.problemTestcaseRepo.ListByProblemID(problem.ID)
	if err != nil {
		log.Printf("[judge-worker] submission=%d load testcases failed: %v", submission.ID, err)
		return w.markSystemError(submission, "load testcases failed")
	}

	activeCases := make([]model.ProblemTestcase, 0)
	for _, testcase := range testcases {
		if testcase.IsActive {
			activeCases = append(activeCases, testcase)
		}
	}
	if len(activeCases) == 0 {
		log.Printf("[judge-worker] submission=%d has no active testcases", submission.ID)
		return w.markSystemError(submission, "no active testcases")
	}

	if err := w.updateSubmissionState(submission, "Running"); err != nil {
		log.Printf("[judge-worker] submission=%d set status Running failed: %v", submission.ID, err)
		return err
	}
	log.Printf("[judge-worker] submission=%d running with %d active testcase(s)", submission.ID, len(activeCases))

	compileResult, err := w.sandboxRunner.Compile(submission.Language, submission.Code)
	if err != nil {
		log.Printf("[judge-worker] submission=%d compile command failed: %v", submission.ID, err)
		return w.finishSubmission(submission, submissionFinish{
			Status:     "System Error",
			TotalCount: len(activeCases),
			ErrorMsg:   err.Error(),
		}, nil)
	}
	defer w.sandboxRunner.Cleanup(compileResult.Program)

	if compileResult.CompileError != "" {
		log.Printf("[judge-worker] submission=%d compile error: %s", submission.ID, summarizeText(compileResult.CompileError, 300))
		return w.finishSubmission(submission, submissionFinish{
			Status:      "Compile Error",
			TotalCount:  len(activeCases),
			CompileInfo: compileResult.CompileError,
		}, nil)
	}

	results := make([]model.SubmissionResult, 0, len(activeCases))
	finalStatus := "Accepted"
	passedCount := 0
	totalRuntime := 0

	for _, testcase := range activeCases {
		runResult := w.sandboxRunner.Run(compileResult.Program, testcase.InputData, problem.TimeLimitMS)
		if runResult.Status == "Accepted" && !compareOutput(runResult.Output, testcase.OutputData) {
			runResult.Status = "Wrong Answer"
			runResult.ErrorMsg = "output mismatch"
		}
		log.Printf(
			"[judge-worker] submission=%d testcase=%d status=%s runtime_ms=%s error=%s",
			submission.ID,
			testcase.ID,
			runResult.Status,
			formatOptionalInt(runResult.RuntimeMS),
			summarizeText(runResult.ErrorMsg, 160),
		)

		result := model.SubmissionResult{
			SubmissionID: submission.ID,
			TestcaseID:   testcase.ID,
			Status:       runResult.Status,
			RuntimeMS:    runResult.RuntimeMS,
			MemoryKB:     nil,
			ErrorMsg:     runResult.ErrorMsg,
		}
		results = append(results, result)

		if runResult.RuntimeMS != nil {
			totalRuntime += *runResult.RuntimeMS
		}

		if runResult.Status == "Accepted" {
			passedCount++
			continue
		}

		if finalStatus == "Accepted" {
			finalStatus = runResult.Status
		}
	}

	runtimeMS := totalRuntime
	err = w.finishSubmission(submission, submissionFinish{
		Status:      finalStatus,
		PassedCount: passedCount,
		TotalCount:  len(activeCases),
		RuntimeMS:   &runtimeMS,
	}, results)
	if err != nil {
		log.Printf("[judge-worker] submission=%d finalize failed: %v", submission.ID, err)
		return err
	}
	log.Printf(
		"[judge-worker] submission=%d finished status=%s passed=%d/%d total_runtime_ms=%d elapsed_ms=%d",
		submission.ID,
		finalStatus,
		passedCount,
		len(activeCases),
		totalRuntime,
		time.Since(startedAt).Milliseconds(),
	)
	return nil
}

type submissionFinish struct {
	Status      string
	PassedCount int
	TotalCount  int
	RuntimeMS   *int
	CompileInfo string
	ErrorMsg    string
}

func (w *Worker) finishSubmission(submission *model.Submission, finish submissionFinish, results []model.SubmissionResult) error {
	now := time.Now()
	submission.Status = finish.Status
	submission.PassedCount = finish.PassedCount
	submission.TotalCount = finish.TotalCount
	submission.RuntimeMS = finish.RuntimeMS
	submission.MemoryKB = nil
	submission.CompileInfo = finish.CompileInfo
	submission.ErrorMsg = finish.ErrorMsg
	submission.JudgedAt = &now

	return w.db.Transaction(func(tx *gorm.DB) error {
		if err := w.submissionRepo.Update(tx, submission); err != nil {
			return err
		}
		return w.submissionResultRepo.ReplaceBySubmissionID(tx, submission.ID, results)
	})
}

func (w *Worker) updateSubmissionState(submission *model.Submission, status string) error {
	submission.Status = status
	return w.db.Transaction(func(tx *gorm.DB) error {
		return w.submissionRepo.Update(tx, submission)
	})
}

func (w *Worker) markSystemError(submission *model.Submission, message string) error {
	return w.finishSubmission(submission, submissionFinish{
		Status:   "System Error",
		ErrorMsg: message,
	}, nil)
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

func summarizeText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func formatOptionalInt(value *int) string {
	if value == nil {
		return "-"
	}
	return strconv.Itoa(*value)
}
