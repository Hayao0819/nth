package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/charmbracelet/x/ansi"
)

func TestSafeText(t *testing.T) {
	t.Parallel()

	got := SafeText("hello\x1b[31m red\x1b[0m\x00\rworld")
	if strings.ContainsRune(got, '\x1b') || strings.ContainsRune(got, '\x00') {
		t.Fatalf("SafeText retained control bytes: %q", got)
	}
	if got != "hello red\nworld" {
		t.Errorf("SafeText = %q", got)
	}
	if got := SafeInline("name\nwith\ttabs"); got != "name with tabs" {
		t.Errorf("SafeInline = %q", got)
	}
}

func TestSafeTextKeepsHalfwidthKanaWidthStable(t *testing.T) {
	t.Parallel()

	source := "ｶﾞｯﾂﾎﾟｰｽﾞ"
	got := SafeText(source)
	if width, want := ansi.StringWidth(got), ansi.StringWidthWc(source); width != want {
		t.Fatalf("SafeText width = %d, want %d: %q", width, want, got)
	}
	for _, line := range WrappedLines(source+source, 6) {
		if width := ansi.StringWidthWc(line); width > 6 {
			t.Errorf("wrapped line width = %d: %q", width, line)
		}
	}
}

func TestWrappedLinesBreaksUnspacedText(t *testing.T) {
	t.Parallel()

	lines := WrappedLines(strings.Repeat("日本語", 10), 8)
	if len(lines) < 2 {
		t.Fatalf("WrappedLines did not wrap: %#v", lines)
	}
}

func TestTextAtUsesTerminalCellWidths(t *testing.T) {
	t.Parallel()

	line := "日本  enter save"
	if !TextAt(line, 8, "enter save") || TextAt(line, 5, "enter save") || TextAt(line, 18, "enter save") {
		t.Fatalf("TextAt did not match the rendered label in %q", line)
	}
}

func TestFriendlyError(t *testing.T) {
	t.Parallel()

	err := &north.APIError{StatusCode: 429, Errors: []north.ErrorDetail{{Code: 88, Message: "Rate limit exceeded"}}}
	if got := FriendlyError(err); got != "Rate limit reached" {
		t.Errorf("FriendlyError = %q", got)
	}
	authError := &north.APIError{StatusCode: 401, Errors: []north.ErrorDetail{{Code: 32}}}
	if got := FriendlyError(authError); !strings.Contains(got, "nth setup") {
		t.Errorf("authentication error = %q", got)
	}
	if got := FriendlyError(errors.New("offline\x1b[31m\nbad")); got != "offline bad" {
		t.Errorf("plain FriendlyError = %q", got)
	}
}
