CREATE TABLE submission_results (
                                    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
                                    submission_id BIGINT UNSIGNED NOT NULL,
                                    testcase_id BIGINT UNSIGNED NOT NULL,
                                    status ENUM(
                                        'Accepted',
                                        'Wrong Answer',
                                        'Runtime Error',
                                        'Time Limit Exceeded',
                                        'System Error'
                                        ) NOT NULL,
                                    runtime_ms INT DEFAULT NULL,
                                    memory_kb INT DEFAULT NULL,
                                    error_message TEXT,
                                    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                    CONSTRAINT fk_submission_results_submission_id
                                        FOREIGN KEY (submission_id) REFERENCES submissions(id)
                                            ON DELETE CASCADE,
                                    CONSTRAINT fk_submission_results_testcase_id
                                        FOREIGN KEY (testcase_id) REFERENCES problem_testcases(id)
                                            ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;