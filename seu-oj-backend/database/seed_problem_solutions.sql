CREATE TABLE IF NOT EXISTS problem_solutions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    problem_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    content LONGTEXT NOT NULL,
    visibility VARCHAR(32) NOT NULL DEFAULT 'public',
    author_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_problem_solutions_problem_id (problem_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO problem_solutions (problem_id, title, content, visibility, author_id)
SELECT 3, 'Official Solution', '# Idea\n\nRead two integers `a` and `b`, then print `a + b`.\n\n## Steps\n\n1. Read `a` and `b`.\n2. Compute `a + b`.\n3. Output the result.\n\n## C++\n\n```cpp\n#include <iostream>\nusing namespace std;\n\nint main() {\n    int a, b;\n    cin >> a >> b;\n    cout << a + b << "\\n";\n    return 0;\n}\n```', 'public', 1
WHERE NOT EXISTS (
    SELECT 1 FROM problem_solutions WHERE problem_id = 3 AND title = 'Official Solution'
);
