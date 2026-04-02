package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"seu-oj-backend/internal/api"
	"seu-oj-backend/internal/config"
	"seu-oj-backend/internal/middleware"
	"seu-oj-backend/internal/queue"
	"seu-oj-backend/internal/repository"
	"seu-oj-backend/internal/sandbox"
	"seu-oj-backend/internal/service"
)

func New(db *gorm.DB, redisClient *redis.Client, cfg config.Config) *gin.Engine {
	engine := gin.Default()
	engine.Use(middleware.CORS())

	authService := service.NewAuthService(db, cfg.Auth.JWTSecret)
	authHandler := api.NewAuthHandler(authService)
	problemRepo := repository.NewProblemRepository(db)
	problemTestcaseRepo := repository.NewProblemTestcaseRepository(db)
	problemService := service.NewProblemService(db, problemRepo, problemTestcaseRepo)
	problemHandler := api.NewProblemHandler(problemService)
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
	submissionService := service.NewSubmissionService(submissionRepo, submissionResultRepo, problemRepo, problemTestcaseRepo, judgeQueue, sandboxRunner)
	submissionHandler := api.NewSubmissionHandler(submissionService)

	apiGroup := engine.Group("/api")
	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"status": "ok",
			},
		})
	})

	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", middleware.JWTAuth(cfg.Auth.JWTSecret), authHandler.Me)
	}

	problemGroup := apiGroup.Group("/problems")
	{
		problemGroup.GET("", problemHandler.List)
		problemGroup.GET("/:id", problemHandler.Detail)
	}

	submissionGroup := apiGroup.Group("/submissions")
	submissionGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
	{
		submissionGroup.POST("/run", submissionHandler.Run)
		submissionGroup.POST("", submissionHandler.Create)
		submissionGroup.GET("/my", submissionHandler.ListMy)
		submissionGroup.GET("/:id", submissionHandler.Detail)
	}

	adminProblemGroup := apiGroup.Group("/admin/problems")
	adminProblemGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireAdmin())
	{
		adminProblemGroup.GET("", problemHandler.AdminList)
		adminProblemGroup.POST("", problemHandler.Create)
		adminProblemGroup.GET("/:id", problemHandler.AdminDetail)
		adminProblemGroup.PUT("/:id", problemHandler.Update)
		adminProblemGroup.DELETE("/:id", problemHandler.Delete)
	}

	registerStaticRoutes(engine)

	return engine
}

func registerStaticRoutes(engine *gin.Engine) {
	frontendDir := resolveFrontendDir()
	indexPath := filepath.Join(frontendDir, "index.html")

	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	engine.StaticFile("/", indexPath)
	engine.StaticFile("/index.html", indexPath)
	engine.StaticFile("/app.js", filepath.Join(frontendDir, "app.js"))
	engine.StaticFile("/styles.css", filepath.Join(frontendDir, "styles.css"))

	engine.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    1,
				"message": "not found",
				"data":    nil,
			})
			return
		}

		c.File(indexPath)
	})
}

func resolveFrontendDir() string {
	candidates := []string{
		filepath.Clean(filepath.Join("..", "seu-oj-frontend")),
		filepath.Clean(filepath.Join(".", "seu-oj-frontend")),
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "..", "seu-oj-frontend"),
			filepath.Join(cwd, "seu-oj-frontend"),
		)
	}

	for _, candidate := range candidates {
		indexPath := filepath.Join(candidate, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return candidate
		}
	}

	return filepath.Clean(filepath.Join("..", "seu-oj-frontend"))
}
