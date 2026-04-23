package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/cache"
	"seu-oj-backend/internal/dto"
)

type RanklistService struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewRanklistService(db *gorm.DB, cacheStore *cache.Cache) *RanklistService {
	return &RanklistService{db: db, cache: cacheStore}
}

func (s *RanklistService) List(page, pageSize int) (*dto.RanklistResponse, error) {
	key := fmt.Sprintf("cache:ranklist:list:p%d:s%d", page, pageSize)
	return cache.GetOrSet(context.Background(), s.cache, key, 20*time.Second, func() (*dto.RanklistResponse, error) {
		return s.list(page, pageSize)
	})
}

func (s *RanklistService) list(page, pageSize int) (*dto.RanklistResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	if err := s.db.Table("users").Where("status = ?", "active").Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows := make([]struct {
		UserID              uint64       `gorm:"column:user_id"`
		Username            string       `gorm:"column:username"`
		StudentID           string       `gorm:"column:userid"`
		SolvedCount         int64        `gorm:"column:solved_count"`
		AcceptedSubmissions int64        `gorm:"column:accepted_submissions"`
		TotalSubmissions    int64        `gorm:"column:total_submissions"`
		LastAcceptedAt      sql.NullTime `gorm:"column:last_accepted_at"`
	}, 0)

	query := `
SELECT
  users.id AS user_id,
  users.username AS username,
  users.userid AS userid,
  COUNT(DISTINCT CASE WHEN submissions.status = 'Accepted' THEN submissions.problem_id END) AS solved_count,
  SUM(CASE WHEN submissions.status = 'Accepted' THEN 1 ELSE 0 END) AS accepted_submissions,
  COUNT(submissions.id) AS total_submissions,
  MAX(CASE WHEN submissions.status = 'Accepted' THEN submissions.created_at END) AS last_accepted_at
FROM users
LEFT JOIN submissions ON submissions.user_id = users.id
WHERE users.status = 'active'
GROUP BY users.id, users.username, users.userid
ORDER BY solved_count DESC, accepted_submissions DESC, total_submissions ASC, last_accepted_at ASC, users.id ASC
LIMIT ? OFFSET ?
`

	if err := s.db.Raw(query, pageSize, offset).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]dto.RanklistItem, 0, len(rows))
	for i, row := range rows {
		var lastAcceptedAt *time.Time
		if row.LastAcceptedAt.Valid {
			t := row.LastAcceptedAt.Time
			lastAcceptedAt = &t
		}
		items = append(items, dto.RanklistItem{
			Rank:                offset + i + 1,
			UserID:              row.UserID,
			Username:            row.Username,
			StudentID:           row.StudentID,
			SolvedCount:         row.SolvedCount,
			AcceptedSubmissions: row.AcceptedSubmissions,
			TotalSubmissions:    row.TotalSubmissions,
			LastAcceptedAt:      lastAcceptedAt,
		})
	}

	return &dto.RanklistResponse{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}
