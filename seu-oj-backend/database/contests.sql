CREATE TABLE contests (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    description LONGTEXT,
    rule_type ENUM('acm') NOT NULL DEFAULT 'acm',
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    is_public TINYINT(1) NOT NULL DEFAULT 1,
    allow_practice TINYINT(1) NOT NULL DEFAULT 0,
    ranklist_freeze_at DATETIME DEFAULT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_contests_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_contests_time_window (start_time, end_time),
    INDEX idx_contests_is_public (is_public)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
