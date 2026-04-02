package repository

import (
	"seu-oj-backend/internal/model"

	"gorm.io/gorm"
)

type ProblemTestcaseRepository struct {
	db *gorm.DB
}

func NewProblemTestcaseRepository(db *gorm.DB) *ProblemTestcaseRepository {
	return &ProblemTestcaseRepository{db: db}
}

func (r *ProblemTestcaseRepository) BatchCreate(tx *gorm.DB, testcases []model.ProblemTestcase) error {
	return tx.Create(&testcases).Error
}

func (r *ProblemTestcaseRepository) ReplaceByProblemID(tx *gorm.DB, problemID uint64, testcases []model.ProblemTestcase) error {
	if err := tx.Where("problem_id = ?", problemID).Delete(&model.ProblemTestcase{}).Error; err != nil {
		return err
	}
	if len(testcases) == 0 {
		return nil
	}
	return tx.Create(&testcases).Error
}

func (r *ProblemTestcaseRepository) ListByProblemID(problemID uint64) ([]model.ProblemTestcase, error) {
	var testcases []model.ProblemTestcase
	err := r.db.
		Where("problem_id = ?", problemID).
		Order("sort_order ASC, id ASC").
		Find(&testcases).Error
	return testcases, err
}
