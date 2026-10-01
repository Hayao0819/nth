package sessiondialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/domain/session"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/charmbracelet/x/ansi"
)

type refreshDoneMsg struct {
	target *Screen
	err    error
}

type Screen struct {
	reactea.BasicComponent

	theme      ui.Theme
	request    *session.RefreshRequest
	refreshing bool
	problem    string
}

func New(theme ui.Theme, request *session.RefreshRequest) *Screen {
	return &Screen{theme: theme, request: request}
}

func (s *Screen) Init(ctx *reactea.Ctx) tea.Cmd {
	ctx.OnDestroy(func() {
		s.request.Complete(session.ErrRefreshCanceled)
	})

	return nil
}

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if done, ok := msg.(refreshDoneMsg); ok {
		if done.target != s {
			return nil
		}
		s.refreshing = false
		if done.err != nil {
			s.problem = ui.FriendlyError(done.err)

			return nil
		}
		s.request.Complete(nil)

		return modal.Dismiss(ctx)
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside {
			lines := strings.Split(ansi.Strip(s.Render(ctx)), "\n")
			if y >= 0 && y < len(lines) {
				line := lines[y]
				if ui.TextAt(line, x, "enter retry") {
					return s.retry(ctx)
				}
				if ui.TextAt(line, x, "q quit") {
					s.request.Complete(session.ErrRefreshCanceled)

					return tea.Quit
				}
			}
		}
	}

	switch {
	case reactea.Key(msg, "enter", "."):
		return s.retry(ctx)
	case reactea.Key(msg, "q"):
		s.request.Complete(session.ErrRefreshCanceled)

		return tea.Quit
	}

	return nil
}

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	if width <= 0 || height <= 0 {
		return ""
	}
	if width < 38 || height < 10 {
		message := strings.Join([]string{
			"nth session",
			"",
			"Browser session expired",
			"resize to at least 38×10",
			"",
			"q quit",
		}, "\n")

		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
	}

	inner := min(width-4, 72)
	header := ui.Sides(
		s.theme.Brand.Render("nth session"),
		s.theme.Dim.Render("Browser authentication"),
		inner,
	)
	lines := []string{
		header,
		"",
		s.theme.Heading.Render("Browser session expired"),
		s.theme.Dim.Render("Sign in to north.rip in the selected browser profile."),
		s.theme.Dim.Render("Then return here and retry reading its cookies."),
		"",
		s.theme.Dim.Render("Browser        ") + s.theme.Heading.Render(ui.SafeInline(s.request.Profile())),
	}
	if s.problem != "" {
		lines = append(lines, "", s.theme.Bad.Render("! "+ui.SafeInline(s.problem)))
	}
	lines = append(lines, "")
	footer := "enter retry   q quit"
	if s.refreshing {
		footer = "refreshing browser session…   q quit"
	}
	lines = append(lines, s.theme.Dim.Render(footer))
	for index, line := range lines {
		lines[index] = ui.Clip(line, inner)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func (s *Screen) retry(ctx *reactea.Ctx) tea.Cmd {
	if s.refreshing {
		return nil
	}
	s.refreshing = true
	s.problem = ""

	return func() tea.Msg {
		return refreshDoneMsg{target: s, err: s.request.Refresh(ctx.Context())}
	}
}

var _ reactea.Component = (*Screen)(nil)
