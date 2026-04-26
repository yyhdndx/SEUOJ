package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
)

type StatsService struct {
	db         *gorm.DB
	judgeQueue *queue.JudgeQueue
	cache      *cache.Cache
}

func NewStatsService(db *gorm.DB, judgeQueue *queue.JudgeQueue, cacheStore *cache.Cache) *StatsService {
	return &StatsService{db: db, judgeQueue: judgeQueue, cache: cacheStore}
}

func (s *StatsService) Overview(ctx context.Context) (*dto.OverviewStatsResponse, error) {
	return cache.GetOrSet(ctx, s.cache, "cache:stats:overview", 60*time.Second, s.overview)
}

func (s *StatsService) overview() (*dto.OverviewStatsResponse, error) {
	resp := &dto.OverviewStatsResponse{}
	if err := s.db.Model(&model.Problem{}).Count(&resp.ProblemsTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Problem{}).Where("visible = ?", true).Count(&resp.VisibleProblems).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.User{}).Count(&resp.UsersTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Count(&resp.SubmissionsTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("status = ?", "Accepted").Count(&resp.AcceptedSubmissions).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *StatsService) My(ctx context.Context, userID uint64) (*dto.UserStatsResponse, error) {
	key := fmt.Sprintf("cache:stats:my:%d", userID)
	return cache.GetOrSet(ctx, s.cache, key, 60*time.Second, func() (*dto.UserStatsResponse, error) {
		return s.my(userID)
	})
}

func (s *StatsService) my(userID uint64) (*dto.UserStatsResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	resp := &dto.UserStatsResponse{
		UserID:   user.ID,
		Username: user.Username,
	}

	var summary struct {
		SubmissionsTotal    int64           `gorm:"column:submissions_total"`
		AcceptedSubmissions int64           `gorm:"column:accepted_submissions"`
		PendingSubmissions  int64           `gorm:"column:pending_submissions"`
		RunningSubmissions  int64           `gorm:"column:running_submissions"`
		AcceptedProblems    int64           `gorm:"column:accepted_problems"`
		AverageRuntimeMS    sql.NullFloat64 `gorm:"column:average_runtime_ms"`
	}
	if err := s.db.Model(&model.Submission{}).
		Select(`
			COUNT(*) AS submissions_total,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS accepted_submissions,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS pending_submissions,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS running_submissions,
			COUNT(DISTINCT CASE WHEN status = ? THEN problem_id END) AS accepted_problems,
			AVG(runtime_ms) AS average_runtime_ms
		`, "Accepted", "Pending", "Running", "Accepted").
		Where("user_id = ?", userID).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	resp.SubmissionsTotal = summary.SubmissionsTotal
	resp.AcceptedSubmissions = summary.AcceptedSubmissions
	resp.PendingSubmissions = summary.PendingSubmissions
	resp.RunningSubmissions = summary.RunningSubmissions
	resp.AcceptedProblems = summary.AcceptedProblems
	if summary.AverageRuntimeMS.Valid {
		resp.AverageRuntimeMS = &summary.AverageRuntimeMS.Float64
	}

	var statusRows []dto.CountItem
	if err := s.db.Model(&model.Submission{}).
		Select("status as name, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Order("count DESC").
		Scan(&statusRows).Error; err != nil {
		return nil, err
	}
	resp.StatusBreakdown = statusRows

	var languageRows []dto.CountItem
	if err := s.db.Model(&model.Submission{}).
		Select("language as name, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("language").
		Order("count DESC").
		Scan(&languageRows).Error; err != nil {
		return nil, err
	}
	resp.LanguageBreakdown = languageRows

	var recentRows []dto.RecentActivityItem
	if err := s.db.Model(&model.Submission{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, COUNT(*) as count").
		Where("user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 14 DAY)", userID).
		Group("DATE_FORMAT(created_at, '%Y-%m-%d')").
		Order("date ASC").
		Scan(&recentRows).Error; err != nil {
		return nil, err
	}
	resp.RecentActivity = recentRows

	return resp, nil
}

func (s *StatsService) Admin(role string) (*dto.AdminStatsResponse, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}

	resp := &dto.AdminStatsResponse{}
	if err := s.db.Model(&model.User{}).Count(&resp.UsersTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Problem{}).Count(&resp.ProblemsTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Problem{}).Where("visible = ?", false).Count(&resp.HiddenProblems).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Count(&resp.SubmissionsTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("status = ?", "Pending").Count(&resp.PendingSubmissions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("status = ?", "Running").Count(&resp.RunningSubmissions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("status = ?", "System Error").Count(&resp.SystemErrors).Error; err != nil {
		return nil, err
	}
	queueLength, err := s.judgeQueue.Length(context.Background())
	if err != nil {
		return nil, err
	}
	resp.QueueLength = queueLength

	var statusRows []dto.CountItem
	if err := s.db.Model(&model.Submission{}).
		Select("status as name, COUNT(*) as count").
		Group("status").
		Order("count DESC").
		Scan(&statusRows).Error; err != nil {
		return nil, err
	}
	resp.StatusBreakdown = statusRows

	return resp, nil
}
