package search

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const debounceDelay = 300 * time.Millisecond

type debounceMsg struct {
	target *Screen
	seq    uint64
	query  string
}

type Screen struct {
	reactea.BasicComponent

	theme   ui.Theme
	input   *reactea.ReactifiedWidget[textinput.Model]
	feed    *feed.Feed
	seq     uint64
	pending bool
}

func New(api feed.API, theme ui.Theme, query string) *Screen {
	return NewWithImages(api, theme, query, nil)
}

func NewWithImages(api feed.API, theme ui.Theme, query string, images *termimage.Renderer) *Screen {
	model := textinput.New()
	model.Prompt = ""
	model.Placeholder = "Search north"
	model.SetValue(query)
	model.SetVirtualCursor(false)
	model.Focus()
	input := reactea.ReactifyWidget(model).OnResize(func(model textinput.Model, width, _ int) textinput.Model {
		model.SetWidth(max(1, width))

		return model
	})

	return &Screen{theme: theme, input: input, feed: feed.NewSearchWithImages(api, theme, images)}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	commands := []tea.Cmd{s.input.Init(s.inputCtx(ctx)), s.input.Widget.Focus()}
	if query := strings.TrimSpace(s.input.Widget.Value()); query != "" {
		s.pending = true
		commands = append(commands, s.debounce(query, 0))
	}

	return tea.Batch(commands...)
}

func (s *Screen) Query() string { return s.input.Widget.Value() }

func (s *Screen) debounce(query string, delay time.Duration) tea.Cmd {
	seq := s.seq
	if delay == 0 {
		return func() tea.Msg { return debounceMsg{target: s, seq: seq, query: query} }
	}

	return tea.Tick(delay, func(time.Time) tea.Msg {
		return debounceMsg{target: s, seq: seq, query: query}
	})
}

func (s *Screen) inputCtx(ctx *reactea.Ctx) *reactea.Ctx {
	return ctx.Inset(6, 1, max(1, ctx.Width()-6), 1)
}

func (s *Screen) feedCtx(ctx *reactea.Ctx) *reactea.Ctx {
	return ctx.Inset(0, pageheader.Height, ctx.Width(), max(0, ctx.Height()-pageheader.Height))
}

var _ reactea.Component = (*Screen)(nil)
