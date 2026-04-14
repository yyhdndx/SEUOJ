# CREATE INDEX idx_problems_created_by ON problems(created_by);
# CREATE INDEX idx_problems_visible ON problems(visible);
#
# CREATE INDEX idx_problem_testcases_problem_id ON problem_testcases(problem_id);
# CREATE INDEX idx_problem_testcases_case_type ON problem_testcases(case_type);
#
# CREATE INDEX idx_contests_time_window ON contests(start_time, end_time);
# CREATE INDEX idx_contests_is_public ON contests(is_public);
# CREATE UNIQUE INDEX uk_contest_problem_code ON contest_problems(contest_id, problem_code);
# CREATE UNIQUE INDEX uk_contest_problem_id ON contest_problems(contest_id, problem_id);
# CREATE UNIQUE INDEX uk_contest_registration ON contest_registrations(contest_id, user_id);
#
# CREATE INDEX idx_submissions_user_id ON submissions(user_id);
# CREATE INDEX idx_submissions_problem_id ON submissions(problem_id);
# CREATE INDEX idx_submissions_contest_id ON submissions(contest_id);
# CREATE INDEX idx_submissions_status ON submissions(status);
# CREATE INDEX idx_submissions_created_at ON submissions(created_at);
#
# CREATE INDEX idx_submission_results_submission_id ON submission_results(submission_id);
# CREATE INDEX idx_submission_results_testcase_id ON submission_results(testcase_id);

select * from contests;
