package dialog

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
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

func TestComposeBuildsAndRestoresThreadItems(t *testing.T) {
	t.Parallel()

	compose := NewCompose(ui.NewTheme(), nil, nil, "first").SetThreadEnabled(true)
	compose.mediaIDs = []string{"media-1"}
	compose.poll = &north.CreatePoll{Options: []string{"yes", "no"}, DurationMinutes: 60}
	compose.addThreadItem()
	if len(compose.thread) != 1 || compose.thread[0].Text != "first" || len(compose.thread[0].MediaIDs) != 1 || compose.thread[0].Poll == nil {
		t.Fatalf("thread = %#v", compose.thread)
	}
	if compose.input.Widget.Value() != "" || len(compose.mediaIDs) != 0 || compose.poll != nil {
		t.Fatalf("current item was not cleared: text=%q media=%#v poll=%#v", compose.input.Widget.Value(), compose.mediaIDs, compose.poll)
	}
	compose.undoThreadItem()
	if len(compose.thread) != 0 || compose.input.Widget.Value() != "first" || compose.poll == nil {
		t.Fatalf("restored composer = thread %#v, text %q, poll %#v", compose.thread, compose.input.Widget.Value(), compose.poll)
	}
}

func TestComposeAcceptsPollForm(t *testing.T) {
	t.Parallel()

	compose := NewCompose(ui.NewTheme(), nil, nil, "question")
	compose.pollForm = true
	command, handled := compose.updatePostOptions(nil, modal.Result[FormResult]{Value: FormResult{
		Values: map[string]string{
			"option1":  "yes",
			"option2":  "no",
			"option3":  "",
			"option4":  "",
			"duration": "60",
		},
		Toggles: map[string]bool{},
	}})
	if !handled || command != nil || compose.poll == nil {
		t.Fatalf("poll result = handled %t, command %#v, poll %#v", handled, command, compose.poll)
	}
	if compose.poll.DurationMinutes != 60 || len(compose.poll.Options) != 2 || compose.poll.Options[0] != "yes" {
		t.Fatalf("poll = %#v", compose.poll)
	}
}
