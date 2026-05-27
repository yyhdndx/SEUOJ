package dto

import (
	"encoding/json"
	"testing"
)

func TestDTOJSONTags(t *testing.T) {
	payload, err := json.Marshal(CreateProblemResponse{ProblemID: 10})
	if err != nil {
		t.Fatalf("marshal create problem response: %v", err)
	}
	if string(payload) != `{"problem_id":10}` {
		t.Fatalf("unexpected create problem json %s", payload)
	}

	var query ProblemListQuery
	if query.Page != 0 || query.PageSize != 0 || query.Keyword != "" {
		t.Fatalf("unexpected zero-value query: %+v", query)
	}
}
