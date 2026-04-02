package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"seu-oj-backend/internal/bootstrap"
	"seu-oj-backend/internal/config"
	"seu-oj-backend/internal/database"
	"seu-oj-backend/internal/judge"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
)

func main() {
	logFile, err := bootstrap.InitLogging()
	if err != nil {
		log.Fatalf("init logging failed: %v", err)
	}
	defer logFile.Close()

	cfg := config.Load()

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis failed: %v", err)
	}

	problemRepo := repository.NewProblemRepository(db)
	problemTestcaseRepo := repository.NewProblemTestcaseRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	submissionResultRepo := repository.NewSubmissionResultRepository(db)
	judgeQueue := queue.NewJudgeQueue(redisClient)
	sandboxRunner := sandbox.NewRunner(sandbox.Config{
		Image:              cfg.Sandbox.DockerImage,
		CompileImage:       cfg.Sandbox.CompileImage,
		RunImage:           cfg.Sandbox.RunImage,
		User:               cfg.Sandbox.User,
		CPUs:               cfg.Sandbox.CPUs,
		MemoryMB:           cfg.Sandbox.MemoryMB,
		PIDsLimit:          cfg.Sandbox.PIDsLimit,
		TmpfsMB:            cfg.Sandbox.TmpfsMB,
		FileSizeKB:         cfg.Sandbox.FileSizeKB,
		OutputLimitKB:      cfg.Sandbox.OutputLimitKB,
		CompileOutputKB:    cfg.Sandbox.CompileOutputKB,
		ReadOnlyRootFS:     cfg.Sandbox.ReadOnlyRootFS,
		CompileTimeoutSec:  cfg.Sandbox.CompileTimeoutSec,
		RunTimeoutBufferMS: cfg.Sandbox.RunTimeoutBufferMS,
	})
	if err := sandboxRunner.Validate(context.Background()); err != nil {
		log.Fatalf("validate sandbox failed: %v", err)
	}

	worker := judge.NewWorker(db, judgeQueue, problemRepo, problemTestcaseRepo, submissionRepo, submissionResultRepo, sandboxRunner)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	worker.Start(ctx)

	<-ctx.Done()
	log.Printf("[judge-worker] shutdown complete")
}
