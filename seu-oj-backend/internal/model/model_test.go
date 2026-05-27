package model

import "testing"

func TestTableNames(t *testing.T) {
	tests := map[string]string{
		Announcement{}.TableName():        "announcements",
		AuditLog{}.TableName():            "audit_logs",
		Classroom{}.TableName():           "classes",
		ClassMember{}.TableName():         "class_members",
		Contest{}.TableName():             "contests",
		ContestAnnouncement{}.TableName(): "contest_announcements",
		ContestProblem{}.TableName():      "contest_problems",
		ContestRegistration{}.TableName(): "contest_registrations",
		ForumReply{}.TableName():          "forum_replies",
		ForumTopic{}.TableName():          "forum_topics",
		ForumTopicFavorite{}.TableName():  "forum_topic_favorites",
		ForumTopicLike{}.TableName():      "forum_topic_likes",
		Playlist{}.TableName():            "playlists",
		PlaylistProblem{}.TableName():     "playlist_problems",
		Problem{}.TableName():             "problems",
		ProblemSolution{}.TableName():     "problem_solutions",
		ProblemTestcase{}.TableName():     "problem_testcases",
		Submission{}.TableName():          "submissions",
		SubmissionResult{}.TableName():    "submission_results",
		User{}.TableName():                "users",
		Assignment{}.TableName():          "assignments",
	}
	for got, want := range tests {
		if got != want {
			t.Fatalf("expected table %q, got %q", want, got)
		}
	}
}
