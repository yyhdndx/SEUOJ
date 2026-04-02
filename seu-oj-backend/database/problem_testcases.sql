CREATE TABLE problem_testcases (
                                   id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
                                   problem_id BIGINT UNSIGNED NOT NULL,
                                   case_type ENUM('sample', 'hidden') NOT NULL DEFAULT 'hidden',
                                   input_data MEDIUMTEXT NOT NULL,
                                   output_data MEDIUMTEXT NOT NULL,
                                   score INT NOT NULL DEFAULT 0,
                                   sort_order INT NOT NULL DEFAULT 0,
                                   is_active TINYINT(1) NOT NULL DEFAULT 1,
                                   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                   CONSTRAINT fk_problem_testcases_problem_id
                                       FOREIGN KEY (problem_id) REFERENCES problems(id)
                                           ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;