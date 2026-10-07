package diagnostic

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const requestTimeout = 15 * time.Second

type resultState uint8

const (
	resultPending resultState = iota
	resultPassed
	resultFailed
)

type result struct {
	name     string
	state    resultState
	summary  string
	err      error
	duration time.Duration
}

type resultMsg struct {
	index    int
	summary  string
	err      error
	duration time.Duration
}

type Screen struct {
	reactea.BasicComponent

	theme  ui.Theme
	checks []check
	items  []result
}

func New(api API) *Screen {
	list := checks(api)
	items := make([]result, len(list))
	for index, check := range list {
		items[index] = result{name: check.name}
	}

	return &Screen{theme: ui.NewTheme(), checks: list, items: items}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	commands := make([]tea.Cmd, len(s.checks))
	for index, check := range s.checks {
		commands[index] = runCheck(ctx.Context(), index, check)
	}

	return tea.Batch(commands...)
}

func (s *Screen) Update(_ *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if reactea.Key(msg, "q", "esc", "enter", "ctrl+c") {
		return tea.Quit
	}
	completed, ok := msg.(resultMsg)
	if !ok || completed.index < 0 || completed.index >= len(s.items) {
		return nil
	}
	item := &s.items[completed.index]
	item.summary = completed.summary
	item.err = completed.err
	item.duration = completed.duration
	if completed.err != nil {
		item.state = resultFailed
	} else {
		item.state = resultPassed
	}

	return nil
}

func runCheck(ctx context.Context, index int, check check) tea.Cmd {
	return func() tea.Msg {
		started := time.Now()
		requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		summary, err := check.run(requestCtx)

		return resultMsg{index: index, summary: summary, err: err, duration: time.Since(started)}
	}
}

func Run(ctx context.Context, api API) error {
	program := reactea.New(
		New(api),
		reactea.WithAltScreen(),
		reactea.WithWindowTitle("nth test"),
	)
	defer program.Scope().Close()
	if err := program.Run(tea.WithContext(ctx)); err != nil {
		return fmt.Errorf("run API test: %w", err)
	}

	return nil
}

var _ reactea.Component = (*Screen)(nil)
