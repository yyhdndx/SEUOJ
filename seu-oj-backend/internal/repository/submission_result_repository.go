package repository

import (
	"seu-oj-backend/internal/model"

	"gorm.io/gorm"
)

type SubmissionResultRepository struct {
	db *gorm.DB
}

func NewSubmissionResultRepository(db *gorm.DB) *SubmissionResultRepository {
	return &SubmissionResultRepository{db: db}
}

func (r *SubmissionResultRepository) ListBySubmissionID(submissionID uint64) ([]model.SubmissionResult, error) {
	var results []model.SubmissionResult
	err := r.db.Where("submission_id = ?", submissionID).Order("id ASC").Find(&results).Error
	return results, err
}

func (r *SubmissionResultRepository) ReplaceBySubmissionID(tx *gorm.DB, submissionID uint64, results []model.SubmissionResult) error {
	if err := tx.Where("submission_id = ?", submissionID).Delete(&model.SubmissionResult{}).Error; err != nil {
		return err
	}
	if len(results) == 0 {
		return nil
	}
	return tx.Create(&results).Error
}
