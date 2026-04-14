CREATE TABLE contest_problems (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    contest_id BIGINT UNSIGNED NOT NULL,
    problem_id BIGINT UNSIGNED NOT NULL,
    problem_code VARCHAR(16) NOT NULL,
    display_order INT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_contest_problems_contest_id FOREIGN KEY (contest_id) REFERENCES contests(id),
    CONSTRAINT fk_contest_problems_problem_id FOREIGN KEY (problem_id) REFERENCES problems(id),
    CONSTRAINT uk_contest_problem_code UNIQUE (contest_id, problem_code),
    CONSTRAINT uk_contest_problem_id UNIQUE (contest_id, problem_id),
    INDEX idx_contest_problems_order (contest_id, display_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
