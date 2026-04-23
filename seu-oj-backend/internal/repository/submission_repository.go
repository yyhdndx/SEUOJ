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

func (r *SubmissionRepository) ListByUserID(userID uint64, page, pageSize int, problemID *uint64, contestID *uint64, status *string) ([]model.Submission, int64, error) {
	return r.List(page, pageSize, &userID, problemID, contestID, status)
}

func (r *SubmissionRepository) ListRecentByUserID(userID uint64, page, pageSize int, problemID *uint64, contestID *uint64, status *string) ([]model.Submission, error) {
	submissions, _, err := r.list(page, pageSize, &userID, problemID, contestID, status, false)
	return submissions, err
}

func (r *SubmissionRepository) List(page, pageSize int, userID *uint64, problemID *uint64, contestID *uint64, status *string) ([]model.Submission, int64, error) {
	return r.list(page, pageSize, userID, problemID, contestID, status, true)
}

func (r *SubmissionRepository) list(page, pageSize int, userID *uint64, problemID *uint64, contestID *uint64, status *string, countTotal bool) ([]model.Submission, int64, error) {
	var submissions []model.Submission
	var total int64

	query := r.db.Model(&model.Submission{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if problemID != nil {
		query = query.Where("problem_id = ?", *problemID)
	}
	if contestID != nil {
		query = query.Where("contest_id = ?", *contestID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	if countTotal {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error
	return submissions, total, err
}

func (r *SubmissionRepository) Update(tx *gorm.DB, submission *model.Submission) error {
	return tx.Save(submission).Error
}
