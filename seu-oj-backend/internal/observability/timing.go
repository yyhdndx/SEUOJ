package observability

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type timingEntry struct {
	Name string
	Dur  time.Duration
	Desc string
}

type TimingObserver struct {
	mu      sync.Mutex
	entries []timingEntry
}

func (t *TimingObserver) Add(name string, dur time.Duration) {
	t.AddDesc(name, dur, "")
}

func (t *TimingObserver) AddDesc(name string, dur time.Duration, desc string) {
	if name == "" {
		return
	}
	t.mu.Lock()
	t.entries = append(t.entries, timingEntry{Name: name, Dur: dur, Desc: desc})
	t.mu.Unlock()
}

func (t *TimingObserver) HeaderValue() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.entries) == 0 {
		return ""
	}
	parts := make([]string, 0, len(t.entries))
	for _, entry := range t.entries {
		token := sanitizeToken(entry.Name)
		if token == "" {
			continue
		}
		seg := token
		if entry.Dur > 0 {
			seg += fmt.Sprintf(";dur=%d", entry.Dur.Milliseconds())
		}
		if entry.Desc != "" {
			seg += fmt.Sprintf(";desc=%q", sanitizeDesc(entry.Desc))
		}
		parts = append(parts, seg)
	}
	return strings.Join(parts, ", ")
}

type ctxKey string

const timingKey ctxKey = "seuoj_timing_observer"

func WithTimingObserver(ctx context.Context, observer *TimingObserver) context.Context {
	return context.WithValue(ctx, timingKey, observer)
}

func FromContext(ctx context.Context) *TimingObserver {
	raw := ctx.Value(timingKey)
	observer, _ := raw.(*TimingObserver)
	return observer
}

func sanitizeToken(value string) string {
	// Server-Timing tokens are limited; keep it conservative.
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z':
			b.WriteRune(ch)
		case ch >= 'A' && ch <= 'Z':
			b.WriteRune(ch)
		case ch >= '0' && ch <= '9':
			b.WriteRune(ch)
		case ch == '_' || ch == '-' || ch == '.':
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func sanitizeDesc(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	// Prevent newlines/quotes from breaking the header.
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\"", "'")
	return value
}
