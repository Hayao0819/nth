package dialog

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
	"github.com/charmbracelet/x/ansi"
)

func TestTextLength(t *testing.T) {
	t.Parallel()

	tests := map[string]int{
		"north": 5,
		"日本":    4,
		"A界":    3,
		"a\nb":  3,
	}
	for input, want := range tests {
		if got := TextLength(input); got != want {
			t.Errorf("TextLength(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestComposeCursorLineUsesDialogBackground(t *testing.T) {
	t.Parallel()

	compose := NewCompose(ui.NewTheme(), nil, nil, "")
	if background := compose.input.Widget.Styles().Focused.CursorLine.GetBackground(); background != (lipgloss.NoColor{}) {
		t.Fatalf("cursor line background = %v", background)
	}

	program := reactea.New(compose, reactea.WithSize(60, 14))
	program.Start()
	blackBackground := ansi.Style{}.BackgroundColor(lipgloss.Color("0")).String()
	frames := []string{program.View().Content}
	testkit.SendKeys(program, "x")
	frames = append(frames, program.View().Content)
	for _, frame := range frames {
		if strings.Contains(frame, blackBackground) {
			t.Fatalf("compose view contains the textarea's black cursor-line background: %q", frame)
		}
	}
}

func TestComposeCountMatchesTheSubmittedValue(t *testing.T) {
	t.Parallel()

	value := strings.Repeat("a", 280) + " "
	program := reactea.New(NewCompose(ui.NewTheme(), nil, nil, value), reactea.WithSize(72, 16))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "281/280") {
		t.Fatalf("compose counter ignored submitted whitespace:\n%s", plain)
	}
}

func TestComposeConcealsItsReplyPreview(t *testing.T) {
	t.Parallel()

	reply := &north.Post{
		Text:         "SHOULD_NOT_RENDER",
		HiddenReason: north.HiddenReason("MUTED"),
		Author:       north.User{Name: "Muted", Handle: "muted"},
	}
	program := reactea.New(NewCompose(ui.NewTheme(), reply, nil, ""), reactea.WithSize(72, 16))
	program.Start()
	plain := testkit.Plain(program)
	if strings.Contains(plain, "SHOULD_NOT_RENDER") || !strings.Contains(plain, "Hidden: MUTED") {
		t.Fatalf("hidden reply leaked through the composer:\n%s", plain)
	}
}
