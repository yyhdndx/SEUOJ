package repository

import (
	"strings"

	"seu-oj-backend/internal/model"

	"gorm.io/gorm"
)

type ProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) *ProblemRepository {
	return &ProblemRepository{db: db}
}

func (r *ProblemRepository) Create(tx *gorm.DB, problem *model.Problem) error {
	return tx.Create(problem).Error
}

func (r *ProblemRepository) Update(tx *gorm.DB, problem *model.Problem) error {
	return tx.Save(problem).Error
}

func (r *ProblemRepository) Delete(tx *gorm.DB, id uint64) error {
	return tx.Delete(&model.Problem{}, id).Error
}

func (r *ProblemRepository) GetByID(id uint64) (*model.Problem, error) {
	var problem model.Problem
	if err := r.db.First(&problem, id).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *ProblemRepository) ListVisible(page, pageSize int, keyword string) ([]model.Problem, int64, error) {
	var problems []model.Problem
	var total int64

	query := r.db.Model(&model.Problem{}).Where("visible = ?", true)
	if trimmedKeyword := strings.TrimSpace(keyword); trimmedKeyword != "" {
		query = query.Where("title LIKE ?", "%"+trimmedKeyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&problems).Error
	return problems, total, err
}

func (r *ProblemRepository) List(page, pageSize int, keyword string, includeHidden bool) ([]model.Problem, int64, error) {
	var problems []model.Problem
	var total int64

	query := r.db.Model(&model.Problem{})
	if !includeHidden {
		query = query.Where("visible = ?", true)
	}
	if trimmedKeyword := strings.TrimSpace(keyword); trimmedKeyword != "" {
		query = query.Where("title LIKE ?", "%"+trimmedKeyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&problems).Error
	return problems, total, err
}
