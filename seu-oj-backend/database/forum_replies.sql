CREATE TABLE IF NOT EXISTS forum_replies (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    topic_id BIGINT UNSIGNED NOT NULL,
    content TEXT NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_forum_replies_topic_id (topic_id),
    INDEX idx_forum_replies_author_id (author_id),
    CONSTRAINT fk_forum_replies_topic FOREIGN KEY (topic_id) REFERENCES forum_topics(id) ON DELETE CASCADE,
    CONSTRAINT fk_forum_replies_author FOREIGN KEY (author_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
