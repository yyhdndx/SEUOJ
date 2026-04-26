package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"seu-oj-backend/internal/api"
	"seu-oj-backend/internal/cache"
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
	engine.Use(middleware.Timing())
	cacheStore := cache.New(redisClient)

	authService := service.NewAuthService(db, cfg.Auth.JWTSecret)
	authHandler := api.NewAuthHandler(authService)
	userService := service.NewUserService(db)
	userHandler := api.NewUserHandler(userService)
	announcementService := service.NewAnnouncementService(db, cacheStore)
	announcementHandler := api.NewAnnouncementHandler(announcementService)
	ranklistService := service.NewRanklistService(db, cacheStore)
	ranklistHandler := api.NewRanklistHandler(ranklistService)
	forumService := service.NewForumService(db, cacheStore)
	forumHandler := api.NewForumHandler(forumService)
	problemRepo := repository.NewProblemRepository(db)
	problemTestcaseRepo := repository.NewProblemTestcaseRepository(db)
	problemService := service.NewProblemService(db, problemRepo, problemTestcaseRepo, cacheStore)
	problemHandler := api.NewProblemHandler(problemService)
	contestService := service.NewContestService(db, problemRepo, problemTestcaseRepo, cacheStore)
	contestHandler := api.NewContestHandler(contestService)
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
	submissionService := service.NewSubmissionService(db, submissionRepo, submissionResultRepo, problemRepo, problemTestcaseRepo, judgeQueue, sandboxRunner, contestService, cacheStore)
	submissionHandler := api.NewSubmissionHandler(submissionService)
	statsService := service.NewStatsService(db, judgeQueue, cacheStore)
	statsHandler := api.NewStatsHandler(statsService)
	teachingService := service.NewTeachingService(db)
	teachingHandler := api.NewTeachingHandler(teachingService)

	apiGroup := engine.Group("/api")
	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "ok"}})
	})

	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", middleware.JWTAuth(cfg.Auth.JWTSecret), authHandler.Me)
		authGroup.PUT("/profile", middleware.JWTAuth(cfg.Auth.JWTSecret), authHandler.UpdateProfile)
		authGroup.PUT("/password", middleware.JWTAuth(cfg.Auth.JWTSecret), authHandler.ChangePassword)
	}

	statsGroup := apiGroup.Group("/stats")
	{
		statsGroup.GET("/overview", statsHandler.Overview)
		statsGroup.GET("/me", middleware.JWTAuth(cfg.Auth.JWTSecret), statsHandler.Me)
		statsGroup.GET("/admin", middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireAdmin(), statsHandler.Admin)
	}

	announcementGroup := apiGroup.Group("/announcements")
	{
		announcementGroup.GET("", announcementHandler.List)
		announcementGroup.GET("/:id", announcementHandler.Detail)
	}

	apiGroup.GET("/ranklist", ranklistHandler.List)

	forumGroup := apiGroup.Group("/forum")
	{
		forumGroup.GET("/topics", forumHandler.ListTopics)
		forumGroup.GET("/topics/:id", forumHandler.TopicDetail)
	}

	contestGroup := apiGroup.Group("/contests")
	{
		contestGroup.GET("", contestHandler.List)
		contestGroup.GET("/:id", contestHandler.Detail)
		contestGroup.GET("/:id/ranklist", contestHandler.Ranklist)
		contestGroup.GET("/:id/announcements", contestHandler.AnnouncementList)
		contestGroup.GET("/:id/announcements/:announcement_id", contestHandler.AnnouncementDetail)
		contestGroup.GET("/:id/me", middleware.JWTAuth(cfg.Auth.JWTSecret), contestHandler.Me)
		contestGroup.POST("/:id/register", middleware.JWTAuth(cfg.Auth.JWTSecret), contestHandler.Register)
		contestGroup.GET("/:id/problems", middleware.JWTAuth(cfg.Auth.JWTSecret), contestHandler.Problems)
		contestGroup.GET("/:id/problems/:problem_id", middleware.JWTAuth(cfg.Auth.JWTSecret), contestHandler.ProblemDetail)
	}

	playlistGroup := apiGroup.Group("/playlists")
	{
		playlistGroup.GET("", teachingHandler.ListPlaylists)
		playlistGroup.GET("/:id", teachingHandler.PlaylistDetail)
	}

	classGroup := apiGroup.Group("/classes")
	classGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
	{
		classGroup.GET("/my", teachingHandler.MyClasses)
		classGroup.POST("/join", teachingHandler.JoinClass)
		classGroup.GET("/:id", teachingHandler.ClassDetail)
	}

	assignmentGroup := apiGroup.Group("/assignments")
	assignmentGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
	{
		assignmentGroup.GET("/:id", teachingHandler.AssignmentDetail)
	}

	problemGroup := apiGroup.Group("/problems")
	{
		problemGroup.GET("", problemHandler.List)
		problemGroup.GET("/:id/stats", problemHandler.Stats)
		problemGroup.GET("/:id/solutions", problemHandler.SolutionList)
		problemGroup.GET("/:id", problemHandler.Detail)
	}

	forumAuthGroup := apiGroup.Group("/forum")
	forumAuthGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
	{
		forumAuthGroup.POST("/topics", forumHandler.CreateTopic)
		forumAuthGroup.PUT("/topics/:id", forumHandler.UpdateTopic)
		forumAuthGroup.DELETE("/topics/:id", forumHandler.DeleteTopic)
		forumAuthGroup.POST("/topics/:id/replies", forumHandler.CreateReply)
		forumAuthGroup.PUT("/replies/:id", forumHandler.UpdateReply)
		forumAuthGroup.DELETE("/replies/:id", forumHandler.DeleteReply)
	}

	submissionGroup := apiGroup.Group("/submissions")
	submissionGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
	{
		submissionGroup.POST("/run", submissionHandler.Run)
		submissionGroup.POST("", submissionHandler.Create)
		submissionGroup.GET("/my", submissionHandler.ListMy)
		submissionGroup.GET("/:id", submissionHandler.Detail)
	}

	adminContestGroup := apiGroup.Group("/admin/contests")
	adminContestGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireAdmin())
	{
		adminContestGroup.GET("", contestHandler.AdminList)
		adminContestGroup.POST("", contestHandler.Create)
		adminContestGroup.GET("/:id", contestHandler.AdminDetail)
		adminContestGroup.GET("/:id/ranklist", contestHandler.AdminRanklist)
		adminContestGroup.GET("/:id/announcements", contestHandler.AdminAnnouncementList)
		adminContestGroup.GET("/:id/announcements/:announcement_id", contestHandler.AdminAnnouncementDetail)
		adminContestGroup.POST("/:id/announcements", contestHandler.CreateAnnouncement)
		adminContestGroup.PUT("/:id/announcements/:announcement_id", contestHandler.UpdateAnnouncement)
		adminContestGroup.DELETE("/:id/announcements/:announcement_id", contestHandler.DeleteAnnouncement)
		adminContestGroup.PUT("/:id", contestHandler.Update)
		adminContestGroup.DELETE("/:id", contestHandler.Delete)
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

	adminSubmissionGroup := apiGroup.Group("/admin/submissions")
	adminSubmissionGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireAdmin())
	{
		adminSubmissionGroup.GET("", submissionHandler.AdminList)
		adminSubmissionGroup.POST("/:id/rejudge", submissionHandler.Rejudge)
	}

	adminUserGroup := apiGroup.Group("/admin/users")
	adminUserGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireAdmin())
	{
		adminUserGroup.GET("", userHandler.AdminList)
		adminUserGroup.PUT("/:id", userHandler.AdminUpdate)
	}

	teacherGroup := apiGroup.Group("/teacher")
	teacherGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireTeacherOrAdmin())
	{
		teacherGroup.GET("/playlists", teachingHandler.TeacherPlaylists)
		teacherGroup.GET("/playlists/:id", teachingHandler.TeacherPlaylistDetail)
		teacherGroup.POST("/playlists", teachingHandler.CreatePlaylist)
		teacherGroup.PUT("/playlists/:id", teachingHandler.UpdatePlaylist)
		teacherGroup.DELETE("/playlists/:id", teachingHandler.DeletePlaylist)
		teacherGroup.GET("/problems/:id/solutions", problemHandler.TeacherSolutionList)
		teacherGroup.POST("/problems/:id/solutions", problemHandler.CreateSolution)
		teacherGroup.PUT("/problems/:id/solutions/:solution_id", problemHandler.UpdateSolution)
		teacherGroup.DELETE("/problems/:id/solutions/:solution_id", problemHandler.DeleteSolution)
		teacherGroup.GET("/classes", teachingHandler.TeacherClasses)
		teacherGroup.POST("/classes", teachingHandler.CreateClass)
		teacherGroup.GET("/classes/:id", teachingHandler.TeacherClassDetail)
		teacherGroup.GET("/classes/:id/analytics", teachingHandler.TeacherClassAnalytics)
		teacherGroup.PUT("/classes/:id", teachingHandler.UpdateClass)
		teacherGroup.GET("/classes/:id/members", teachingHandler.ClassMembers)
		teacherGroup.PUT("/classes/:id/members/:user_id", teachingHandler.UpdateClassMember)
		teacherGroup.GET("/classes/:id/assignments", teachingHandler.TeacherAssignments)
		teacherGroup.POST("/classes/:id/assignments", teachingHandler.CreateAssignment)
		teacherGroup.GET("/assignments/:id", teachingHandler.TeacherAssignmentOverview)
		teacherGroup.PUT("/assignments/:id", teachingHandler.UpdateAssignment)
		teacherGroup.DELETE("/assignments/:id", teachingHandler.DeleteAssignment)
	}

	adminAnnouncementGroup := apiGroup.Group("/admin/announcements")
	adminAnnouncementGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret), middleware.RequireAdmin())
	{
		adminAnnouncementGroup.POST("", announcementHandler.Create)
		adminAnnouncementGroup.PUT("/:id", announcementHandler.Update)
		adminAnnouncementGroup.DELETE("/:id", announcementHandler.Delete)
	}

	registerStaticRoutes(engine)
	return engine
}

func registerStaticRoutes(engine *gin.Engine) {
	frontendDir := resolveFrontendDir()
	indexPath := filepath.Join(frontendDir, "index.html")
	appJSPath := filepath.Join(frontendDir, "app.js")
	stylesPath := filepath.Join(frontendDir, "styles.css")
	jsDir := filepath.Join(frontendDir, "js")
	cssDir := filepath.Join(frontendDir, "css")
	codeMirrorDir := filepath.Join(frontendDir, "CodeMirror")

	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	if _, err := os.Stat(jsDir); err == nil {
		engine.StaticFS("/js", gin.Dir(jsDir, false))
	}
	if _, err := os.Stat(cssDir); err == nil {
		engine.StaticFS("/css", gin.Dir(cssDir, false))
	}
	if _, err := os.Stat(codeMirrorDir); err == nil {
		engine.StaticFS("/CodeMirror", gin.Dir(codeMirrorDir, false))
	}

	engine.GET("/", func(c *gin.Context) {
		serveFrontendFile(c, indexPath)
	})
	engine.GET("/index.html", func(c *gin.Context) {
		serveFrontendFile(c, indexPath)
	})
	engine.GET("/app.js", func(c *gin.Context) {
		serveFrontendFile(c, appJSPath)
	})
	engine.GET("/styles.css", func(c *gin.Context) {
		serveFrontendFile(c, stylesPath)
	})
	engine.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	engine.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "not found", "data": nil})
			return
		}
		serveFrontendFile(c, indexPath)
	})
}

func serveFrontendFile(c *gin.Context, path string) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.File(path)
}

func resolveFrontendDir() string {
	candidates := []string{filepath.Clean(filepath.Join("..", "seu-oj-frontend")), filepath.Clean(filepath.Join(".", "seu-oj-frontend"))}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "..", "seu-oj-frontend"), filepath.Join(cwd, "seu-oj-frontend"))
	}
	for _, candidate := range candidates {
		indexPath := filepath.Join(candidate, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return candidate
		}
	}
	return filepath.Clean(filepath.Join("..", "seu-oj-frontend"))
}
