package auth

import (
	"testing"
	"time"
)

func TestDaysUntilReadsMilliseconds(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expiry := now.Add(30*24*time.Hour + time.Hour).UnixMilli()
	if got := daysUntil(expiry, now); got != 30 {
		t.Fatalf("daysUntil = %d, want 30", got)
	}
}

func TestSpacedCode(t *testing.T) {
	if got := spacedCode("ABCD2345"); got != "A B C D - 2 3 4 5" {
		t.Fatalf("spacedCode = %q", got)
	}
}
