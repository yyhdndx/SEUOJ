-- Minimal performance indexes for current high-frequency queries.
-- Keep indexes tied to concrete WHERE / ORDER BY patterns; avoid low-value single-column indexes.

DELIMITER //

DROP PROCEDURE IF EXISTS add_index_if_missing//

CREATE PROCEDURE add_index_if_missing(
    IN p_table_name VARCHAR(64),
    IN p_index_name VARCHAR(64),
    IN p_statement TEXT
)
BEGIN
    DECLARE index_count INT;

    SELECT COUNT(1) INTO index_count
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = p_table_name
      AND index_name = p_index_name;

    IF index_count = 0 THEN
        SET @sql = p_statement;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END//

DELIMITER ;

-- GET /api/announcements: ORDER BY is_pinned DESC, id DESC
CALL add_index_if_missing('announcements', 'idx_announcements_pinned_id', 'CREATE INDEX idx_announcements_pinned_id ON announcements(is_pinned, id)');

-- GET /api/problems: WHERE visible = 1 ORDER BY id DESC
CALL add_index_if_missing('problems', 'idx_problems_visible_id', 'CREATE INDEX idx_problems_visible_id ON problems(visible, id)');

-- Problem detail / judge worker: WHERE problem_id = ? ORDER BY sort_order, id
CALL add_index_if_missing('problem_testcases', 'idx_problem_testcases_problem_sort', 'CREATE INDEX idx_problem_testcases_problem_sort ON problem_testcases(problem_id, sort_order, id)');

-- Problem solutions: WHERE problem_id = ? [AND visibility = ?]
CALL add_index_if_missing('problem_solutions', 'idx_problem_solutions_problem_visibility', 'CREATE INDEX idx_problem_solutions_problem_visibility ON problem_solutions(problem_id, visibility)');

-- Contest list: WHERE is_public = ? plus time-window filters, ORDER BY start_time DESC, id DESC
CALL add_index_if_missing('contests', 'idx_contests_public_start', 'CREATE INDEX idx_contests_public_start ON contests(is_public, start_time, id)');

-- Contest problem list / ranklist headers: WHERE contest_id = ? ORDER BY display_order, id
CALL add_index_if_missing('contest_problems', 'idx_contest_problems_contest_order', 'CREATE INDEX idx_contest_problems_contest_order ON contest_problems(contest_id, display_order, id)');
CALL add_index_if_missing('contest_problems', 'uk_contest_problem_code', 'CREATE UNIQUE INDEX uk_contest_problem_code ON contest_problems(contest_id, problem_code)');
CALL add_index_if_missing('contest_problems', 'uk_contest_problem_id', 'CREATE UNIQUE INDEX uk_contest_problem_id ON contest_problems(contest_id, problem_id)');

-- Contest registration checks and participant counts: WHERE contest_id = ? [AND user_id = ?]
CALL add_index_if_missing('contest_registrations', 'uk_contest_user', 'CREATE UNIQUE INDEX uk_contest_user ON contest_registrations(contest_id, user_id)');

-- Contest announcements: WHERE contest_id = ? ORDER BY is_pinned DESC, id DESC
CALL add_index_if_missing('contest_announcements', 'idx_contest_announcements_contest_pinned', 'CREATE INDEX idx_contest_announcements_contest_pinned ON contest_announcements(contest_id, is_pinned, id)');

-- Forum scoped lists and topic replies.
CALL add_index_if_missing('forum_topics', 'idx_forum_topics_scope', 'CREATE INDEX idx_forum_topics_scope ON forum_topics(scope_type, scope_id)');
CALL add_index_if_missing('forum_replies', 'idx_forum_replies_topic_id', 'CREATE INDEX idx_forum_replies_topic_id ON forum_replies(topic_id)');

-- Submission lists, stats, ranklists, and recent activity.
CALL add_index_if_missing('submissions', 'idx_submissions_user_id', 'CREATE INDEX idx_submissions_user_id ON submissions(user_id)');
CALL add_index_if_missing('submissions', 'idx_submissions_user_status', 'CREATE INDEX idx_submissions_user_status ON submissions(user_id, status)');
CALL add_index_if_missing('submissions', 'idx_submissions_user_created', 'CREATE INDEX idx_submissions_user_created ON submissions(user_id, created_at)');
CALL add_index_if_missing('submissions', 'idx_submissions_problem_status', 'CREATE INDEX idx_submissions_problem_status ON submissions(problem_id, status)');
CALL add_index_if_missing('submissions', 'idx_submissions_contest_created', 'CREATE INDEX idx_submissions_contest_created ON submissions(contest_id, created_at)');
CALL add_index_if_missing('submissions', 'idx_submissions_status', 'CREATE INDEX idx_submissions_status ON submissions(status)');

-- Submission detail: WHERE submission_id = ? ORDER BY id
CALL add_index_if_missing('submission_results', 'idx_submission_results_submission_id', 'CREATE INDEX idx_submission_results_submission_id ON submission_results(submission_id)');

DROP PROCEDURE IF EXISTS add_index_if_missing;
