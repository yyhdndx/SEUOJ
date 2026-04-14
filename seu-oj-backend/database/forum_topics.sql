CREATE TABLE IF NOT EXISTS forum_topics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    scope_type ENUM('general', 'problem', 'contest') NOT NULL DEFAULT 'general',
    scope_id BIGINT UNSIGNED NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    reply_count INT NOT NULL DEFAULT 0,
    is_pinned TINYINT(1) NOT NULL DEFAULT 0,
    is_locked TINYINT(1) NOT NULL DEFAULT 0,
    last_reply_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_forum_topics_scope (scope_type, scope_id),
    INDEX idx_forum_topics_author_id (author_id),
    INDEX idx_forum_topics_pinned (is_pinned),
    INDEX idx_forum_topics_last_reply_at (last_reply_at),
    CONSTRAINT fk_forum_topics_author FOREIGN KEY (author_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
