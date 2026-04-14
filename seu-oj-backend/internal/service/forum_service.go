package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

var (
	ErrForumTopicNotFound = errors.New("forum topic not found")
	ErrForumReplyNotFound = errors.New("forum reply not found")
	ErrForumForbidden     = errors.New("forum access denied")
	ErrForumLocked        = errors.New("forum topic locked")
)

type ForumService struct{ db *gorm.DB }

func NewForumService(db *gorm.DB) *ForumService { return &ForumService{db: db} }

func (s *ForumService) ListTopics(page, pageSize int, keyword, scopeType string, scopeID *uint64) (*dto.ForumTopicListResponse, error) {
	query := s.db.Table("forum_topics t").
		Select("t.id, t.title, LEFT(t.content, 160) AS content_preview, t.scope_type, t.scope_id, t.author_id, u.username AS author_name, t.reply_count, t.is_pinned, t.is_locked, t.last_reply_at, t.created_at, t.updated_at").
		Joins("JOIN users u ON u.id = t.author_id")
	if scopeType = strings.TrimSpace(scopeType); scopeType != "" {
		query = query.Where("t.scope_type = ?", scopeType)
	}
	if scopeID != nil {
		query = query.Where("t.scope_id = ?", *scopeID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("t.title LIKE ? OR t.content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []dto.ForumTopicListItem
	if err := query.Order("t.is_pinned DESC, COALESCE(t.last_reply_at, t.created_at) DESC, t.id DESC").Offset((page-1)*pageSize).Limit(pageSize).Scan(&list).Error; err != nil {
		return nil, err
	}
	return &dto.ForumTopicListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ForumService) GetTopicDetail(id uint64) (*dto.ForumTopicDetailResponse, error) {
	var topic dto.ForumTopicDetailResponse
	if err := s.db.Table("forum_topics t").
		Select("t.id, t.title, t.content, t.scope_type, t.scope_id, t.author_id, u.username AS author_name, t.reply_count, t.is_pinned, t.is_locked, t.last_reply_at, t.created_at, t.updated_at").
		Joins("JOIN users u ON u.id = t.author_id").
		Where("t.id = ?", id).
		Take(&topic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForumTopicNotFound
		}
		return nil, err
	}
	var replies []dto.ForumReplyResponse
	if err := s.db.Table("forum_replies r").
		Select("r.id, r.topic_id, r.content, r.author_id, u.username AS author_name, r.created_at, r.updated_at").
		Joins("JOIN users u ON u.id = r.author_id").
		Where("r.topic_id = ?", id).
		Order("r.id ASC").
		Scan(&replies).Error; err != nil {
		return nil, err
	}
	topic.Replies = replies
	return &topic, nil
}

func (s *ForumService) CreateTopic(userID uint64, req dto.CreateForumTopicRequest) (*dto.ForumTopicDetailResponse, error) {
	if req.ScopeType != "general" && req.ScopeID == nil {
		return nil, ErrForumForbidden
	}
	if req.ScopeType == "problem" && req.ScopeID != nil {
		var count int64
		if err := s.db.Model(&model.Problem{}).Where("id = ?", *req.ScopeID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, ErrProblemNotFound
		}
	}
	if req.ScopeType == "contest" && req.ScopeID != nil {
		var count int64
		if err := s.db.Model(&model.Contest{}).Where("id = ?", *req.ScopeID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, ErrContestNotFound
		}
	}
	topic := model.ForumTopic{Title: strings.TrimSpace(req.Title), Content: req.Content, ScopeType: req.ScopeType, ScopeID: req.ScopeID, AuthorID: userID}
	if err := s.db.Create(&topic).Error; err != nil {
		return nil, err
	}
	return s.GetTopicDetail(topic.ID)
}

func (s *ForumService) UpdateTopic(userID uint64, role string, topicID uint64, req dto.UpdateForumTopicRequest) (*dto.ForumTopicDetailResponse, error) {
	var topic model.ForumTopic
	if err := s.db.First(&topic, topicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForumTopicNotFound
		}
		return nil, err
	}
	if role != "admin" && topic.AuthorID != userID && role != "teacher" {
		return nil, ErrForumForbidden
	}
	topic.Title = strings.TrimSpace(req.Title)
	topic.Content = req.Content
	if role == "admin" || role == "teacher" {
		if req.IsPinned != nil {
			topic.IsPinned = *req.IsPinned
		}
		if req.IsLocked != nil {
			topic.IsLocked = *req.IsLocked
		}
	}
	if err := s.db.Save(&topic).Error; err != nil {
		return nil, err
	}
	return s.GetTopicDetail(topic.ID)
}

func (s *ForumService) DeleteTopic(userID uint64, role string, topicID uint64) error {
	var topic model.ForumTopic
	if err := s.db.First(&topic, topicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrForumTopicNotFound
		}
		return err
	}
	if role != "admin" && role != "teacher" && topic.AuthorID != userID {
		return ErrForumForbidden
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("topic_id = ?", topicID).Delete(&model.ForumReply{}).Error; err != nil {
			return err
		}
		return tx.Delete(&topic).Error
	})
}

func (s *ForumService) CreateReply(userID uint64, role string, topicID uint64, req dto.CreateForumReplyRequest) (*dto.ForumReplyResponse, error) {
	var topic model.ForumTopic
	if err := s.db.First(&topic, topicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForumTopicNotFound
		}
		return nil, err
	}
	if topic.IsLocked && role != "admin" && role != "teacher" {
		return nil, ErrForumLocked
	}
	reply := model.ForumReply{TopicID: topicID, Content: req.Content, AuthorID: userID}
	now := time.Now()
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&reply).Error; err != nil {
			return err
		}
		return tx.Model(&model.ForumTopic{}).Where("id = ?", topicID).Updates(map[string]any{"reply_count": gorm.Expr("reply_count + 1"), "last_reply_at": &now}).Error
	}); err != nil {
		return nil, err
	}
	var result dto.ForumReplyResponse
	if err := s.db.Table("forum_replies r").
		Select("r.id, r.topic_id, r.content, r.author_id, u.username AS author_name, r.created_at, r.updated_at").
		Joins("JOIN users u ON u.id = r.author_id").
		Where("r.id = ?", reply.ID).
		Take(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ForumService) UpdateReply(userID uint64, role string, replyID uint64, req dto.CreateForumReplyRequest) (*dto.ForumReplyResponse, error) {
	var reply model.ForumReply
	if err := s.db.First(&reply, replyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrForumReplyNotFound
		}
		return nil, err
	}
	if role != "admin" && role != "teacher" && reply.AuthorID != userID {
		return nil, ErrForumForbidden
	}
	reply.Content = req.Content
	if err := s.db.Save(&reply).Error; err != nil {
		return nil, err
	}
	var result dto.ForumReplyResponse
	if err := s.db.Table("forum_replies r").
		Select("r.id, r.topic_id, r.content, r.author_id, u.username AS author_name, r.created_at, r.updated_at").
		Joins("JOIN users u ON u.id = r.author_id").
		Where("r.id = ?", reply.ID).
		Take(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ForumService) DeleteReply(userID uint64, role string, replyID uint64) error {
	var reply model.ForumReply
	if err := s.db.First(&reply, replyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrForumReplyNotFound
		}
		return err
	}
	if role != "admin" && role != "teacher" && reply.AuthorID != userID {
		return ErrForumForbidden
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&reply).Error; err != nil {
			return err
		}
		updates := map[string]any{"reply_count": gorm.Expr("GREATEST(reply_count - 1, 0)")}
		var latest time.Time
		if err := tx.Model(&model.ForumReply{}).Where("topic_id = ?", reply.TopicID).Select("MAX(created_at)").Scan(&latest).Error; err == nil && !latest.IsZero() {
			updates["last_reply_at"] = latest
		} else {
			updates["last_reply_at"] = nil
		}
		return tx.Model(&model.ForumTopic{}).Where("id = ?", reply.TopicID).Updates(updates).Error
	})
}
