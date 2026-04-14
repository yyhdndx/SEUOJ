SET NAMES utf8mb4;

INSERT INTO forum_topics (title, content, scope_type, scope_id, author_id, reply_count, is_pinned, is_locked, last_reply_at, created_at, updated_at)
SELECT 'Welcome to SEU OJ Forum', 'Use this forum for general discussion about training, bug reports, and problem solving ideas.', 'general', NULL, 4, 0, 1, 0, NULL, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM forum_topics WHERE title = 'Welcome to SEU OJ Forum');

INSERT INTO forum_topics (title, content, scope_type, scope_id, author_id, reply_count, is_pinned, is_locked, last_reply_at, created_at, updated_at)
SELECT 'How should I solve Problem 1002?', 'I can finish the A plus B task, but I want to discuss input edge cases and output formatting here.', 'problem', 3, 1, 0, 0, 0, NULL, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM forum_topics WHERE title = 'How should I solve Problem 1002?');

INSERT INTO forum_topics (title, content, scope_type, scope_id, author_id, reply_count, is_pinned, is_locked, last_reply_at, created_at, updated_at)
SELECT 'Contest warmup strategy thread', 'Share practice ideas, contest pacing tips, and any clarification requests for the warmup contest.', 'contest', 1, 2, 0, 0, 0, NULL, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM forum_topics WHERE title = 'Contest warmup strategy thread');

INSERT INTO forum_replies (topic_id, content, author_id, created_at, updated_at)
SELECT t.id, 'Use standard input carefully and remember to end the output with a newline.', 4, NOW(), NOW()
FROM forum_topics t
WHERE t.title = 'How should I solve Problem 1002?'
  AND NOT EXISTS (
    SELECT 1 FROM forum_replies r WHERE r.topic_id = t.id AND r.author_id = 4
  );

INSERT INTO forum_replies (topic_id, content, author_id, created_at, updated_at)
SELECT t.id, 'For the warmup contest, I suggest solving A quickly and then checking all sample cases before moving on.', 1, NOW(), NOW()
FROM forum_topics t
WHERE t.title = 'Contest warmup strategy thread'
  AND NOT EXISTS (
    SELECT 1 FROM forum_replies r WHERE r.topic_id = t.id AND r.author_id = 1
  );

UPDATE forum_topics t
LEFT JOIN (
    SELECT topic_id, COUNT(*) AS reply_count, MAX(created_at) AS last_reply_at
    FROM forum_replies
    GROUP BY topic_id
) r ON r.topic_id = t.id
SET t.reply_count = COALESCE(r.reply_count, 0),
    t.last_reply_at = r.last_reply_at;

UPDATE forum_topics SET is_pinned = 1 WHERE title = 'Welcome to SEU OJ Forum';
