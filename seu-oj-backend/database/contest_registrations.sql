CREATE TABLE contest_registrations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    contest_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_contest_registrations_contest_id FOREIGN KEY (contest_id) REFERENCES contests(id),
    CONSTRAINT fk_contest_registrations_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uk_contest_registration UNIQUE (contest_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
