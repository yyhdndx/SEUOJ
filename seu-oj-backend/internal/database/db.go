package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"seu-oj-backend/internal/config"
	"seu-oj-backend/internal/model"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&model.Announcement{},
		&model.Contest{},
		&model.ContestProblem{},
		&model.ContestRegistration{},
		&model.ContestAnnouncement{},
		&model.Playlist{},
		&model.PlaylistProblem{},
		&model.Classroom{},
		&model.ClassMember{},
		&model.Assignment{},
		&model.ProblemSolution{},
		&model.ForumTopic{},
		&model.ForumReply{},
	); err != nil {
		return nil, err
	}
	if !db.Migrator().HasColumn(&model.Submission{}, "contest_id") {
		if err := db.Exec("ALTER TABLE submissions ADD COLUMN contest_id BIGINT UNSIGNED NULL AFTER problem_id").Error; err != nil {
			return nil, err
		}
	}
	if !db.Migrator().HasColumn(&model.Submission{}, "is_practice") {
		if err := db.Exec("ALTER TABLE submissions ADD COLUMN is_practice TINYINT(1) NOT NULL DEFAULT 0 AFTER contest_id").Error; err != nil {
			return nil, err
		}
	}
	if !db.Migrator().HasColumn(&model.ForumTopic{}, "is_pinned") {
		if err := db.Exec("ALTER TABLE forum_topics ADD COLUMN is_pinned TINYINT(1) NOT NULL DEFAULT 0 AFTER reply_count").Error; err != nil {
			return nil, err
		}
	}
	if !db.Migrator().HasColumn(&model.ForumTopic{}, "is_locked") {
		if err := db.Exec("ALTER TABLE forum_topics ADD COLUMN is_locked TINYINT(1) NOT NULL DEFAULT 0 AFTER is_pinned").Error; err != nil {
			return nil, err
		}
	}
	return db, nil
}
