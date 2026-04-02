CREATE TABLE submissions (
                             id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
                             user_id BIGINT UNSIGNED NOT NULL,
                             problem_id BIGINT UNSIGNED NOT NULL,
                             language ENUM('cpp','c','java','python3','go','rust') NOT NULL DEFAULT 'cpp',
                             code MEDIUMTEXT NOT NULL,
                             status ENUM(
                                 'Pending',
                                 'Running',
                                 'Accepted',
                                 'Wrong Answer',
                                 'Compile Error',
                                 'Runtime Error',
                                 'Time Limit Exceeded',
                                 'Memory Limit Exceeded',
                                 'System Error'
                                 ) NOT NULL DEFAULT 'Pending',
                             passed_count INT NOT NULL DEFAULT 0,
                             total_count INT NOT NULL DEFAULT 0,
                             runtime_ms INT DEFAULT NULL,
                             memory_kb INT DEFAULT NULL,
                             compile_info TEXT,
                             error_message TEXT,
                             created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             judged_at DATETIME DEFAULT NULL,
                             CONSTRAINT fk_submissions_user_id
                                 FOREIGN KEY (user_id) REFERENCES users(id),
                             CONSTRAINT fk_submissions_problem_id
                                 FOREIGN KEY (problem_id) REFERENCES problems(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;