package observability

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestTimingObserverBuildsServerTimingHeader(t *testing.T) {
	observer := &TimingObserver{}
	observer.AddDesc(" db query ", 12*time.Millisecond, "read \"quoted\"\nvalue")
	observer.Add("cache-hit", 0)
	observer.Add("", time.Second)

	header := observer.HeaderValue()
	if !strings.Contains(header, "dbquery;dur=12;desc=\"read 'quoted' value\"") {
		t.Fatalf("expected sanitized db segment, got %q", header)
	}
	if !strings.Contains(header, "cache-hit") {
		t.Fatalf("expected cache segment, got %q", header)
	}
	if strings.Contains(header, " db query ") {
		t.Fatalf("header was not sanitized: %q", header)
	}
}

func TestTimingObserverContextRoundTrip(t *testing.T) {
	observer := &TimingObserver{}
	ctx := WithTimingObserver(context.Background(), observer)

	if FromContext(ctx) != observer {
		t.Fatal("expected observer from context")
	}
	if FromContext(context.Background()) != nil {
		t.Fatal("expected nil observer without context value")
	}
}

func TestSanitizersDropUnsafeCharacters(t *testing.T) {
	if got := sanitizeToken(" db/@query;1 "); got != "dbquery1" {
		t.Fatalf("unexpected sanitized token %q", got)
	}
	if got := sanitizeDesc(" a\r\n\"b\" "); got != "a  'b'" {
		t.Fatalf("unexpected sanitized desc %q", got)
	}
}
