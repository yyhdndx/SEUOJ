package repository

import (
	"seu-oj-backend/internal/model"

	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

func (r *SubmissionRepository) GetByID(id uint64) (*model.Submission, error) {
	var submission model.Submission
	if err := r.db.First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) ListByUserID(userID uint64, page, pageSize int, problemID *uint64, status *string) ([]model.Submission, int64, error) {
	var submissions []model.Submission
	var total int64

	query := r.db.Model(&model.Submission{}).Where("user_id = ?", userID)
	if problemID != nil {
		query = query.Where("problem_id = ?", *problemID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error
	return submissions, total, err
}

func (r *SubmissionRepository) Update(tx *gorm.DB, submission *model.Submission) error {
	return tx.Save(submission).Error
}
