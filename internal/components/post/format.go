package post

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
)

func RenderPoll(poll *north.Poll, width int, theme ui.Theme) []string {
	total := poll.TotalVotes
	if total <= 0 {
		for _, option := range poll.Options {
			total += option.VoteCount
		}
	}

	lines := make([]string, 0, len(poll.Options)+1)
	for _, option := range poll.Options {
		selected := poll.ViewerOptionID != nil && *poll.ViewerOptionID == option.ID
		marker := "○"
		if selected {
			marker = "●"
		}
		percent := option.Percent
		if percent == 0 && total > 0 && option.VoteCount > 0 {
			percent = float64(option.VoteCount) * 100 / float64(total)
		}
		line := ui.Sides(marker+" "+ui.SafeInline(option.Label), fmt.Sprintf("%.0f%%", percent), width)
		if selected {
			line = theme.Active.Render(line)
		}
		lines = append(lines, "  "+line)
	}
	lines = append(lines, "  "+theme.Dim.Render(fmt.Sprintf("%d votes", total)))

	return lines
}

func RenderBox(title string, body []string, width int) []string {
	width = max(4, width)
	label := "─ " + ui.SafeInline(title) + " "
	label = ui.Clip(label, width-2)
	top := "╭" + label + strings.Repeat("─", max(0, width-2-lipgloss.Width(label))) + "╮"
	lines := []string{top}
	for _, line := range body {
		lines = append(lines, "│"+ui.Left(" "+line, width-2)+"│")
	}
	lines = append(lines, "╰"+strings.Repeat("─", width-2)+"╯")

	return lines
}

func metricLabel(icon string, count int) string {
	if count == 0 {
		return icon
	}

	return fmt.Sprintf("%s %d", icon, count)
}

func RelativeTime(when, now time.Time) string {
	if when.IsZero() {
		return "unknown"
	}
	delta := now.Sub(when)
	if delta < 0 {
		delta = 0
	}
	switch {
	case delta < time.Minute:
		return "now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh", int(delta.Hours()))
	case delta < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(delta.Hours()/24))
	default:
		return when.Local().Format("2006-01-02")
	}
}
