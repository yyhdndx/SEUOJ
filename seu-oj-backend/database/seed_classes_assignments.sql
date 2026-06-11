-- Idempotent demo data for the teaching/class assignment scenario.
--
-- Creates:
--   - one demo teacher account
--   - one demo admin account
--   - three demo student accounts
--   - three active classes
--   - class members for each class
--   - class-visible playlists with problems
--   - homework/exam assignments for every demo class
--   - a few submissions so progress dashboards are not empty
--
-- Demo account password for all inserted users: 123456
--
-- Recommended order:
--   1) Run user/problem seeds first if your database is mostly empty.
--   2) Run this file.
--
-- Problem binding:
--   This file uses display_id 1003-1007 when present. If those problems are
--   missing, corresponding playlist problem rows and demo submissions are skipped.

START TRANSACTION;

SET @demo_password_hash = '$2b$10$lCnJ6mfOG.QCyRQxgl/cDeXvVgWtkyeNduvXqTihKkwdThjin75X.';

-- Demo users.
INSERT INTO users (username, userid, password_hash, role, status, created_at, updated_at)
SELECT 'demo_teacher', 'T-DEMO-2026', @demo_password_hash, 'teacher', 'active', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'demo_teacher' OR userid = 'T-DEMO-2026');

INSERT INTO users (username, userid, password_hash, role, status, created_at, updated_at)
SELECT 'demo_alice', 'S-DEMO-001', @demo_password_hash, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'demo_alice' OR userid = 'S-DEMO-001');

INSERT INTO users (username, userid, password_hash, role, status, created_at, updated_at)
SELECT 'demo_bob', 'S-DEMO-002', @demo_password_hash, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'demo_bob' OR userid = 'S-DEMO-002');

INSERT INTO users (username, userid, password_hash, role, status, created_at, updated_at)
SELECT 'demo_cindy', 'S-DEMO-003', @demo_password_hash, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'demo_cindy' OR userid = 'S-DEMO-003');

INSERT INTO users (username, userid, password_hash, role, status, created_at, updated_at)
SELECT 'demo_admin', 'A-DEMO-2026', @demo_password_hash, 'admin', 'active', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'demo_admin' OR userid = 'A-DEMO-2026');

UPDATE users
SET role = 'teacher', status = 'active'
WHERE username = 'demo_teacher';

UPDATE users
SET role = 'student', status = 'active'
WHERE username IN ('demo_alice', 'demo_bob', 'demo_cindy');

SET @teacher_id = (SELECT id FROM users WHERE username = 'demo_teacher' LIMIT 1);
SET @alice_id = (SELECT id FROM users WHERE username = 'demo_alice' LIMIT 1);
SET @bob_id = (SELECT id FROM users WHERE username = 'demo_bob' LIMIT 1);
SET @cindy_id = (SELECT id FROM users WHERE username = 'demo_cindy' LIMIT 1);

-- Demo classes.
INSERT INTO classes (name, description, join_code, teacher_id, status, created_at, updated_at)
SELECT
  'Demo C Programming 2026 Spring',
  'Demo class: basic programming homework and progress tracking.',
  'DEMO-C-2026',
  @teacher_id,
  'active',
  NOW(),
  NOW()
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM classes WHERE join_code = 'DEMO-C-2026');

INSERT INTO classes (name, description, join_code, teacher_id, status, created_at, updated_at)
SELECT
  'Demo Data Structure Lab',
  'Demo class: arrays, strings, and simple data structure exercises.',
  'DEMO-DS-2026',
  @teacher_id,
  'active',
  NOW(),
  NOW()
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM classes WHERE join_code = 'DEMO-DS-2026');

INSERT INTO classes (name, description, join_code, teacher_id, status, created_at, updated_at)
SELECT
  'Demo Algorithm Sprint',
  'Demo class: short-cycle algorithm assignments and exam review.',
  'DEMO-ALG-2026',
  @teacher_id,
  'active',
  NOW(),
  NOW()
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM classes WHERE join_code = 'DEMO-ALG-2026');

UPDATE classes
SET teacher_id = @teacher_id, status = 'active', updated_at = NOW()
WHERE join_code IN ('DEMO-C-2026', 'DEMO-DS-2026', 'DEMO-ALG-2026')
  AND @teacher_id IS NOT NULL;

SET @class_c = (SELECT id FROM classes WHERE join_code = 'DEMO-C-2026' LIMIT 1);
SET @class_ds = (SELECT id FROM classes WHERE join_code = 'DEMO-DS-2026' LIMIT 1);
SET @class_alg = (SELECT id FROM classes WHERE join_code = 'DEMO-ALG-2026' LIMIT 1);

-- Class members.
INSERT INTO class_members (class_id, user_id, role, status, created_at, updated_at)
SELECT @class_c, @alice_id, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE @class_c IS NOT NULL AND @alice_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM class_members WHERE class_id = @class_c AND user_id = @alice_id);

INSERT INTO class_members (class_id, user_id, role, status, created_at, updated_at)
SELECT @class_c, @bob_id, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE @class_c IS NOT NULL AND @bob_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM class_members WHERE class_id = @class_c AND user_id = @bob_id);

INSERT INTO class_members (class_id, user_id, role, status, created_at, updated_at)
SELECT @class_ds, @alice_id, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE @class_ds IS NOT NULL AND @alice_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM class_members WHERE class_id = @class_ds AND user_id = @alice_id);

INSERT INTO class_members (class_id, user_id, role, status, created_at, updated_at)
SELECT @class_ds, @cindy_id, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE @class_ds IS NOT NULL AND @cindy_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM class_members WHERE class_id = @class_ds AND user_id = @cindy_id);

INSERT INTO class_members (class_id, user_id, role, status, created_at, updated_at)
SELECT @class_alg, @bob_id, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE @class_alg IS NOT NULL AND @bob_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM class_members WHERE class_id = @class_alg AND user_id = @bob_id);

INSERT INTO class_members (class_id, user_id, role, status, created_at, updated_at)
SELECT @class_alg, @cindy_id, 'student', 'active', NOW(), NOW()
FROM DUAL
WHERE @class_alg IS NOT NULL AND @cindy_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM class_members WHERE class_id = @class_alg AND user_id = @cindy_id);

UPDATE class_members
SET status = 'active', updated_at = NOW()
WHERE class_id IN (@class_c, @class_ds, @class_alg)
  AND user_id IN (@alice_id, @bob_id, @cindy_id);

-- Problem ids used by demo playlists. These are created by seed_more_problems.sql.
SET @p_max = (SELECT id FROM problems WHERE display_id = '1003' LIMIT 1);
SET @p_even = (SELECT id FROM problems WHERE display_id = '1004' LIMIT 1);
SET @p_sum = (SELECT id FROM problems WHERE display_id = '1005' LIMIT 1);
SET @p_sort = (SELECT id FROM problems WHERE display_id = '1006' LIMIT 1);
SET @p_reverse = (SELECT id FROM problems WHERE display_id = '1007' LIMIT 1);

-- Class playlists.
INSERT INTO playlists (title, description, visibility, created_by, created_at, updated_at)
SELECT
  'Demo C Class Homework 01',
  'Class playlist for the C programming demo class.',
  'class',
  @teacher_id,
  NOW(),
  NOW()
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM playlists WHERE title = 'Demo C Class Homework 01');

INSERT INTO playlists (title, description, visibility, created_by, created_at, updated_at)
SELECT
  'Demo Data Structure Homework 01',
  'Class playlist for arrays and strings.',
  'class',
  @teacher_id,
  NOW(),
  NOW()
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM playlists WHERE title = 'Demo Data Structure Homework 01');

INSERT INTO playlists (title, description, visibility, created_by, created_at, updated_at)
SELECT
  'Demo Algorithm Sprint Exam',
  'Class playlist for a short exam-style assignment.',
  'class',
  @teacher_id,
  NOW(),
  NOW()
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM playlists WHERE title = 'Demo Algorithm Sprint Exam');

SET @pl_c = (SELECT id FROM playlists WHERE title = 'Demo C Class Homework 01' LIMIT 1);
SET @pl_ds = (SELECT id FROM playlists WHERE title = 'Demo Data Structure Homework 01' LIMIT 1);
SET @pl_alg = (SELECT id FROM playlists WHERE title = 'Demo Algorithm Sprint Exam' LIMIT 1);

UPDATE playlists
SET visibility = 'class', created_by = @teacher_id, updated_at = NOW()
WHERE id IN (@pl_c, @pl_ds, @pl_alg)
  AND @teacher_id IS NOT NULL;

DELETE FROM playlist_problems WHERE playlist_id IN (@pl_c, @pl_ds, @pl_alg);

INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_c, @p_max, 1, NOW() FROM DUAL WHERE @pl_c IS NOT NULL AND @p_max IS NOT NULL;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_c, @p_even, 2, NOW() FROM DUAL WHERE @pl_c IS NOT NULL AND @p_even IS NOT NULL;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_c, @p_sum, 3, NOW() FROM DUAL WHERE @pl_c IS NOT NULL AND @p_sum IS NOT NULL;

INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_ds, @p_sum, 1, NOW() FROM DUAL WHERE @pl_ds IS NOT NULL AND @p_sum IS NOT NULL;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_ds, @p_sort, 2, NOW() FROM DUAL WHERE @pl_ds IS NOT NULL AND @p_sort IS NOT NULL;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_ds, @p_reverse, 3, NOW() FROM DUAL WHERE @pl_ds IS NOT NULL AND @p_reverse IS NOT NULL;

INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_alg, @p_max, 1, NOW() FROM DUAL WHERE @pl_alg IS NOT NULL AND @p_max IS NOT NULL;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_alg, @p_sort, 2, NOW() FROM DUAL WHERE @pl_alg IS NOT NULL AND @p_sort IS NOT NULL;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order, created_at)
SELECT @pl_alg, @p_reverse, 3, NOW() FROM DUAL WHERE @pl_alg IS NOT NULL AND @p_reverse IS NOT NULL;

-- Assignments: every demo class has at least one teacher-created assignment.
INSERT INTO assignments (class_id, playlist_id, title, description, type, start_at, due_at, created_by, created_at, updated_at)
SELECT
  @class_c,
  @pl_c,
  'Week 1 Basic Programming Homework',
  'Teacher demo assignment: input/output, branches, and simple loops.',
  'homework',
  DATE_SUB(NOW(), INTERVAL 2 DAY),
  DATE_ADD(NOW(), INTERVAL 5 DAY),
  @teacher_id,
  NOW(),
  NOW()
FROM DUAL
WHERE @class_c IS NOT NULL AND @pl_c IS NOT NULL AND @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM assignments WHERE class_id = @class_c AND title = 'Week 1 Basic Programming Homework'
  );

INSERT INTO assignments (class_id, playlist_id, title, description, type, start_at, due_at, created_by, created_at, updated_at)
SELECT
  @class_ds,
  @pl_ds,
  'Arrays and Strings Lab',
  'Teacher demo assignment: list processing and string operations.',
  'homework',
  DATE_SUB(NOW(), INTERVAL 1 DAY),
  DATE_ADD(NOW(), INTERVAL 7 DAY),
  @teacher_id,
  NOW(),
  NOW()
FROM DUAL
WHERE @class_ds IS NOT NULL AND @pl_ds IS NOT NULL AND @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM assignments WHERE class_id = @class_ds AND title = 'Arrays and Strings Lab'
  );

INSERT INTO assignments (class_id, playlist_id, title, description, type, start_at, due_at, created_by, created_at, updated_at)
SELECT
  @class_alg,
  @pl_alg,
  'Algorithm Sprint Mock Exam',
  'Teacher demo assignment: timed practice with three problems.',
  'exam',
  DATE_SUB(NOW(), INTERVAL 3 HOUR),
  DATE_ADD(NOW(), INTERVAL 2 DAY),
  @teacher_id,
  NOW(),
  NOW()
FROM DUAL
WHERE @class_alg IS NOT NULL AND @pl_alg IS NOT NULL AND @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM assignments WHERE class_id = @class_alg AND title = 'Algorithm Sprint Mock Exam'
  );

UPDATE assignments
SET updated_at = NOW()
WHERE title IN (
  'Week 1 Basic Programming Homework',
  'Arrays and Strings Lab',
  'Algorithm Sprint Mock Exam'
);

-- Demo submissions for progress pages. They are normal practice submissions.
INSERT INTO submissions (user_id, problem_id, contest_id, is_practice, language, code, status, passed_count, total_count, runtime_ms, memory_kb, compile_info, error_message, created_at, judged_at)
SELECT @alice_id, @p_max, NULL, 0, 'cpp', '// seed_classes_assignments: demo_alice p_max', 'Accepted', 3, 3, 12, 2048, '', '', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY)
FROM DUAL
WHERE @alice_id IS NOT NULL AND @p_max IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM submissions WHERE user_id = @alice_id AND problem_id = @p_max AND code LIKE '%seed_classes_assignments:%');

INSERT INTO submissions (user_id, problem_id, contest_id, is_practice, language, code, status, passed_count, total_count, runtime_ms, memory_kb, compile_info, error_message, created_at, judged_at)
SELECT @alice_id, @p_even, NULL, 0, 'cpp', '// seed_classes_assignments: demo_alice p_even', 'Accepted', 3, 3, 10, 2048, '', '', DATE_SUB(NOW(), INTERVAL 22 HOUR), DATE_SUB(NOW(), INTERVAL 22 HOUR)
FROM DUAL
WHERE @alice_id IS NOT NULL AND @p_even IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM submissions WHERE user_id = @alice_id AND problem_id = @p_even AND code LIKE '%seed_classes_assignments:%');

INSERT INTO submissions (user_id, problem_id, contest_id, is_practice, language, code, status, passed_count, total_count, runtime_ms, memory_kb, compile_info, error_message, created_at, judged_at)
SELECT @bob_id, @p_max, NULL, 0, 'cpp', '// seed_classes_assignments: demo_bob p_max', 'Wrong Answer', 1, 3, 8, 2048, '', 'demo wrong answer', DATE_SUB(NOW(), INTERVAL 20 HOUR), DATE_SUB(NOW(), INTERVAL 20 HOUR)
FROM DUAL
WHERE @bob_id IS NOT NULL AND @p_max IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM submissions WHERE user_id = @bob_id AND problem_id = @p_max AND code LIKE '%seed_classes_assignments:%');

INSERT INTO submissions (user_id, problem_id, contest_id, is_practice, language, code, status, passed_count, total_count, runtime_ms, memory_kb, compile_info, error_message, created_at, judged_at)
SELECT @cindy_id, @p_sum, NULL, 0, 'python3', '# seed_classes_assignments: demo_cindy p_sum', 'Accepted', 3, 3, 20, 4096, '', '', DATE_SUB(NOW(), INTERVAL 18 HOUR), DATE_SUB(NOW(), INTERVAL 18 HOUR)
FROM DUAL
WHERE @cindy_id IS NOT NULL AND @p_sum IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM submissions WHERE user_id = @cindy_id AND problem_id = @p_sum AND code LIKE '%seed_classes_assignments:%');

INSERT INTO submissions (user_id, problem_id, contest_id, is_practice, language, code, status, passed_count, total_count, runtime_ms, memory_kb, compile_info, error_message, created_at, judged_at)
SELECT @bob_id, @p_sort, NULL, 0, 'cpp', '// seed_classes_assignments: demo_bob p_sort', 'Accepted', 3, 3, 16, 2048, '', '', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR)
FROM DUAL
WHERE @bob_id IS NOT NULL AND @p_sort IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM submissions WHERE user_id = @bob_id AND problem_id = @p_sort AND code LIKE '%seed_classes_assignments:%');

INSERT INTO submissions (user_id, problem_id, contest_id, is_practice, language, code, status, passed_count, total_count, runtime_ms, memory_kb, compile_info, error_message, created_at, judged_at)
SELECT @cindy_id, @p_reverse, NULL, 0, 'python3', '# seed_classes_assignments: demo_cindy p_reverse', 'Wrong Answer', 2, 3, 18, 4096, '', 'demo wrong answer', DATE_SUB(NOW(), INTERVAL 5 HOUR), DATE_SUB(NOW(), INTERVAL 5 HOUR)
FROM DUAL
WHERE @cindy_id IS NOT NULL AND @p_reverse IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM submissions WHERE user_id = @cindy_id AND problem_id = @p_reverse AND code LIKE '%seed_classes_assignments:%');

COMMIT;

SELECT id, username, userid, role, status
FROM users
WHERE username IN ('demo_teacher', 'demo_alice', 'demo_bob', 'demo_cindy')
ORDER BY role DESC, id;

SELECT c.id, c.name, c.join_code, u.username AS teacher_name, c.status
FROM classes c
JOIN users u ON u.id = c.teacher_id
WHERE c.join_code IN ('DEMO-C-2026', 'DEMO-DS-2026', 'DEMO-ALG-2026')
ORDER BY c.id;

SELECT c.name AS class_name, a.title AS assignment_title, a.type, p.title AS playlist_title, a.start_at, a.due_at
FROM assignments a
JOIN classes c ON c.id = a.class_id
JOIN playlists p ON p.id = a.playlist_id
WHERE c.join_code IN ('DEMO-C-2026', 'DEMO-DS-2026', 'DEMO-ALG-2026')
ORDER BY c.id, a.id;
