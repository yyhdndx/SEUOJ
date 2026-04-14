package service

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/queue"
)

type StatsService struct {
	db         *gorm.DB
	judgeQueue *queue.JudgeQueue
}

func NewStatsService(db *gorm.DB, judgeQueue *queue.JudgeQueue) *StatsService {
	return &StatsService{db: db, judgeQueue: judgeQueue}
}

func (s *StatsService) Overview() (*dto.OverviewStatsResponse, error) {
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

func (s *StatsService) My(userID uint64) (*dto.UserStatsResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	resp := &dto.UserStatsResponse{
		UserID:   user.ID,
		Username: user.Username,
	}
	if err := s.db.Model(&model.Submission{}).Where("user_id = ?", userID).Count(&resp.SubmissionsTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("user_id = ? AND status = ?", userID, "Accepted").Count(&resp.AcceptedSubmissions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("user_id = ? AND status = ?", userID, "Pending").Count(&resp.PendingSubmissions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Where("user_id = ? AND status = ?", userID, "Running").Count(&resp.RunningSubmissions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Submission{}).Distinct("problem_id").Where("user_id = ? AND status = ?", userID, "Accepted").Count(&resp.AcceptedProblems).Error; err != nil {
		return nil, err
	}

	var avg sql.NullFloat64
	if err := s.db.Model(&model.Submission{}).Where("user_id = ? AND runtime_ms IS NOT NULL", userID).Select("AVG(runtime_ms)").Scan(&avg).Error; err != nil {
		return nil, err
	}
	if avg.Valid {
		resp.AverageRuntimeMS = &avg.Float64
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
