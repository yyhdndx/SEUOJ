CREATE TABLE contest_announcements (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    contest_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    content MEDIUMTEXT NOT NULL,
    is_pinned TINYINT(1) NOT NULL DEFAULT 0,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_contest_announcements_contest_id FOREIGN KEY (contest_id) REFERENCES contests(id),
    INDEX idx_contest_announcements_contest_id (contest_id),
    INDEX idx_contest_announcements_pinned (contest_id, is_pinned, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
