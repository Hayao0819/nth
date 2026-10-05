package diagnostic

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	if width <= 0 || height <= 0 {
		return ""
	}
	if width < 48 || height < 18 {
		message := strings.Join([]string{
			"nth test",
			"",
			fmt.Sprintf("terminal is %d×%d", width, height),
			"resize to at least 48×18",
			"",
			"q close",
		}, "\n")

		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
	}

	panelStyle := s.theme.Box.Padding(1, 2)
	panelWidth := min(width-4, 78)
	panelHeight := 16
	contentWidth := panelWidth - panelStyle.GetHorizontalFrameSize()
	contentHeight := panelHeight - panelStyle.GetVerticalFrameSize()
	passed, failed, completed := s.counts()
	header := ui.Sides(
		s.theme.Brand.Render("nth test"),
		s.theme.Dim.Render(fmt.Sprintf("%d / %d", completed, len(s.items))),
		contentWidth,
	)
	lines := []string{
		header,
		"",
		s.theme.Heading.Render("Live API check") + "  " + s.theme.Dim.Render("Read-only requests to north"),
		"",
	}
	for _, item := range s.items {
		lines = append(lines, s.resultLine(item, contentWidth))
	}
	lines = append(lines, "", ui.Sides(
		s.footerStatus(passed, failed, completed),
		s.theme.Dim.Render("q close"),
		contentWidth,
	))
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}
	for index, line := range lines {
		lines[index] = ui.Clip(line, contentWidth)
	}

	panel := panelStyle.Width(panelWidth).Height(panelHeight).Render(strings.Join(lines, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, panel)
}

func (s *Screen) resultLine(item result, width int) string {
	marker := s.theme.Dim.Render("·")
	summary := "Requesting…"
	duration := ""
	switch item.state {
	case resultPassed:
		marker = s.theme.Active.Render("✓")
		summary = item.summary
		duration = elapsed(item.duration)
	case resultFailed:
		marker = s.theme.Bad.Render("×")
		summary = ui.FriendlyError(item.err)
		duration = elapsed(item.duration)
	}
	prefix := marker + " " + ui.Left(item.name, 15)
	timeWidth := 0
	if duration != "" {
		timeWidth = lipgloss.Width(duration) + 2
	}
	summaryWidth := max(1, width-lipgloss.Width(prefix)-timeWidth)
	line := prefix + ui.Left(summary, summaryWidth)
	if duration != "" {
		line += "  " + s.theme.Dim.Render(duration)
	}

	return line
}

func (s *Screen) counts() (int, int, int) {
	passed := 0
	failed := 0
	for _, item := range s.items {
		switch item.state {
		case resultPassed:
			passed++
		case resultFailed:
			failed++
		}
	}

	return passed, failed, passed + failed
}

func (s *Screen) footerStatus(passed, failed, completed int) string {
	if completed < len(s.items) {
		return s.theme.Dim.Render(fmt.Sprintf("Checking… %d remaining", len(s.items)-completed))
	}
	if failed == 0 {
		return s.theme.Active.Render(fmt.Sprintf("All %d checks passed", passed))
	}

	return s.theme.Heading.Render(fmt.Sprintf("%d passed · %d failed", passed, failed))
}

func elapsed(duration time.Duration) string {
	if duration < time.Millisecond {
		return "<1ms"
	}

	return duration.Round(time.Millisecond).String()
}
