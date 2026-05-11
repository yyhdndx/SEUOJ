-- Idempotent demo playlists (safe to re-run).
-- Requirements:
--   - At least one user with role teacher or admin (sets @teacher_id).
--   - Rows in problems for display_id 1003-1007 (run seed_more_problems.sql if needed).
--
-- All user-facing strings in this file are ASCII-only so latin1 / utf8mb4 columns and
-- mixed server collations do not raise ERROR 1267 or ERROR 1366.
--
-- Failure cases to check if something still errors:
--   1) No teacher/admin -> @teacher_id NULL -> no new playlist rows; problem inserts skipped
--      (every INSERT ... SELECT includes AND @pl_* IS NOT NULL).
--   2) Missing problem display_id -> that INSERT inserts 0 rows; playlist may have fewer problems.
--   3) Wrong ENUM for visibility -> alter table if your schema differs from public/private/class.

START TRANSACTION;

SET @teacher_id = (
  SELECT id FROM users
  WHERE role IN ('teacher', 'admin')
  ORDER BY FIELD(role, 'teacher', 'admin'), id
  LIMIT 1
);

-- 1) public
INSERT INTO playlists (title, description, visibility, created_by)
SELECT
  'Demo Basic Syntax Training',
  'Public demo: IO and control flow. SEU OJ seed data.',
  'public',
  @teacher_id
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM playlists WHERE title = 'Demo Basic Syntax Training'
  );

SET @pl_basic = (
  SELECT id FROM playlists WHERE title = 'Demo Basic Syntax Training' LIMIT 1
);

UPDATE playlists
SET
  description = 'Public demo: IO and control flow. SEU OJ seed data.',
  visibility = 'public'
WHERE id = @pl_basic AND @pl_basic IS NOT NULL;

DELETE FROM playlist_problems WHERE playlist_id = @pl_basic;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_basic, id, 1 FROM problems WHERE display_id = '1003' AND @pl_basic IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_basic, id, 2 FROM problems WHERE display_id = '1004' AND @pl_basic IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_basic, id, 3 FROM problems WHERE display_id = '1005' AND @pl_basic IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_basic, id, 4 FROM problems WHERE display_id = '1006' AND @pl_basic IS NOT NULL LIMIT 1;

-- 2) public
INSERT INTO playlists (title, description, visibility, created_by)
SELECT
  'Demo Loops and Arrays',
  'Public demo: multi-problem progress and continue practice.',
  'public',
  @teacher_id
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM playlists WHERE title = 'Demo Loops and Arrays'
  );

SET @pl_loop = (
  SELECT id FROM playlists WHERE title = 'Demo Loops and Arrays' LIMIT 1
);

UPDATE playlists
SET
  description = 'Public demo: multi-problem progress and continue practice.',
  visibility = 'public'
WHERE id = @pl_loop AND @pl_loop IS NOT NULL;

DELETE FROM playlist_problems WHERE playlist_id = @pl_loop;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_loop, id, 1 FROM problems WHERE display_id = '1004' AND @pl_loop IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_loop, id, 2 FROM problems WHERE display_id = '1005' AND @pl_loop IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_loop, id, 3 FROM problems WHERE display_id = '1006' AND @pl_loop IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_loop, id, 4 FROM problems WHERE display_id = '1007' AND @pl_loop IS NOT NULL LIMIT 1;

-- 3) public
INSERT INTO playlists (title, description, visibility, created_by)
SELECT
  'Demo Strings Practice',
  'Public demo: longer list and difficulty labels from problems.difficulty.',
  'public',
  @teacher_id
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM playlists WHERE title = 'Demo Strings Practice'
  );

SET @pl_str = (
  SELECT id FROM playlists WHERE title = 'Demo Strings Practice' LIMIT 1
);

UPDATE playlists
SET
  description = 'Public demo: longer list and difficulty labels from problems.difficulty.',
  visibility = 'public'
WHERE id = @pl_str AND @pl_str IS NOT NULL;

DELETE FROM playlist_problems WHERE playlist_id = @pl_str;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_str, id, 1 FROM problems WHERE display_id = '1003' AND @pl_str IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_str, id, 2 FROM problems WHERE display_id = '1005' AND @pl_str IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_str, id, 3 FROM problems WHERE display_id = '1006' AND @pl_str IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_str, id, 4 FROM problems WHERE display_id = '1007' AND @pl_str IS NOT NULL LIMIT 1;

-- 4) class visibility (assignments)
INSERT INTO playlists (title, description, visibility, created_by)
SELECT
  'Demo Class Homework Playlist',
  'Class visibility demo for teacher assignments.',
  'class',
  @teacher_id
FROM DUAL
WHERE @teacher_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM playlists WHERE title = 'Demo Class Homework Playlist'
  );

SET @pl_class = (
  SELECT id FROM playlists WHERE title = 'Demo Class Homework Playlist' LIMIT 1
);

UPDATE playlists
SET
  description = 'Class visibility demo for teacher assignments.',
  visibility = 'class'
WHERE id = @pl_class AND @pl_class IS NOT NULL;

DELETE FROM playlist_problems WHERE playlist_id = @pl_class;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_class, id, 1 FROM problems WHERE display_id = '1003' AND @pl_class IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_class, id, 2 FROM problems WHERE display_id = '1004' AND @pl_class IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_class, id, 3 FROM problems WHERE display_id = '1005' AND @pl_class IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_class, id, 4 FROM problems WHERE display_id = '1006' AND @pl_class IS NOT NULL LIMIT 1;
INSERT INTO playlist_problems (playlist_id, problem_id, display_order)
SELECT @pl_class, id, 5 FROM problems WHERE display_id = '1007' AND @pl_class IS NOT NULL LIMIT 1;

COMMIT;

SELECT id, title, visibility, created_by FROM playlists
WHERE title IN (
  'Demo Basic Syntax Training',
  'Demo Loops and Arrays',
  'Demo Strings Practice',
  'Demo Class Homework Playlist'
)
ORDER BY id;

SELECT p.title AS playlist_title, pp.display_order, pr.display_id, pr.title AS problem_title
FROM playlist_problems pp
JOIN playlists p ON p.id = pp.playlist_id
JOIN problems pr ON pr.id = pp.problem_id
WHERE p.title IN (
  'Demo Basic Syntax Training',
  'Demo Loops and Arrays',
  'Demo Strings Practice',
  'Demo Class Homework Playlist'
)
ORDER BY p.id, pp.display_order;
