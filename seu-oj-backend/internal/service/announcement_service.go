package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

var ErrAnnouncementNotFound = errors.New("announcement not found")

type AnnouncementService struct {
	db *gorm.DB
}

func NewAnnouncementService(db *gorm.DB) *AnnouncementService {
	return &AnnouncementService{db: db}
}

func (s *AnnouncementService) List(page, pageSize int) (*dto.AnnouncementListResponse, error) {
	var list []model.Announcement
	var total int64
	query := s.db.Model(&model.Announcement{})
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("is_pinned DESC, id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return toAnnouncementListResponse(list, total, page, pageSize), nil
}

func (s *AnnouncementService) GetByID(id uint64) (*dto.AnnouncementItem, error) {
	var item model.Announcement
	if err := s.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAnnouncementNotFound
		}
		return nil, err
	}
	resp := toAnnouncementItem(item)
	return &resp, nil
}

func (s *AnnouncementService) Create(userID uint64, role string, req dto.CreateAnnouncementRequest) (uint64, error) {
	if role != "admin" {
		return 0, ErrPermissionDenied
	}
	item := model.Announcement{
		Title:     strings.TrimSpace(req.Title),
		Content:   req.Content,
		IsPinned:  req.IsPinned,
		CreatedBy: userID,
	}
	if err := s.db.Create(&item).Error; err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *AnnouncementService) Update(role string, id uint64, req dto.CreateAnnouncementRequest) error {
	if role != "admin" {
		return ErrPermissionDenied
	}
	var item model.Announcement
	if err := s.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAnnouncementNotFound
		}
		return err
	}
	item.Title = strings.TrimSpace(req.Title)
	item.Content = req.Content
	item.IsPinned = req.IsPinned
	return s.db.Save(&item).Error
}

func (s *AnnouncementService) Delete(role string, id uint64) error {
	if role != "admin" {
		return ErrPermissionDenied
	}
	return s.db.Delete(&model.Announcement{}, id).Error
}

func toAnnouncementListResponse(list []model.Announcement, total int64, page, pageSize int) *dto.AnnouncementListResponse {
	items := make([]dto.AnnouncementItem, 0, len(list))
	for _, item := range list {
		items = append(items, toAnnouncementItem(item))
	}
	return &dto.AnnouncementListResponse{List: items, Total: total, Page: page, PageSize: pageSize}
}

func toAnnouncementItem(item model.Announcement) dto.AnnouncementItem {
	return dto.AnnouncementItem{
		ID:        item.ID,
		Title:     item.Title,
		Content:   item.Content,
		IsPinned:  item.IsPinned,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
