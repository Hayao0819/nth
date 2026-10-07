package saved

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const savedItemHeight = 4

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	tabs := []string{"Drafts", "Scheduled", "Undo queue"}
	for index := range tabs {
		if tab(index) == s.tab {
			tabs[index] = s.theme.Active.Bold(true).Render(tabs[index])
		} else {
			tabs[index] = s.theme.Dim.Render(tabs[index])
		}
	}
	right := strings.Join(tabs, "   ")
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-14)))
	} else if s.loading {
		right = s.theme.Dim.Render("Loading…")
	}
	header := pageheader.Render(s.theme, "Saved posts", right, width)
	room := s.room(height)
	body := s.renderItems(width, room)
	footer := "Tab switch · j/k move · Enter/e edit · c create · d delete"
	if s.tab == pendingTab {
		footer = "Tab switch · j/k move · c create · s send now · u undo"
	}
	footer += " · Esc/← back"

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Screen) renderItems(width, room int) string {
	if s.count() == 0 {
		message := "Nothing saved"
		switch {
		case s.loading:
			message = "Loading…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}

		return ui.Fit(lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message), width, room)
	}
	lines := make([]string, 0, room)
	for index := s.top; index < s.count() && len(lines)+savedItemHeight <= room; index++ {
		lines = append(lines, strings.Split(s.renderItem(index, width, index == s.selected), "\n")...)
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (s *Screen) renderItem(index, width int, selected bool) string {
	draft := s.draftAt(index)
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	preview := "[empty post]"
	if len(draft.Items) > 0 {
		preview = strings.TrimSpace(draft.Items[0].Text)
		if preview == "" && len(draft.Items[0].Media) > 0 {
			preview = "[media]"
		}
	}
	first := marker + "  " + ui.Clip(ui.SafeInline(preview), max(1, width-3))
	meta := savedTime(draft.UpdatedAt)
	if len(draft.Items) > 1 {
		meta += fmt.Sprintf(" · %d posts", len(draft.Items))
	}
	status := "Draft"
	switch s.tab {
	case scheduledTab:
		item := s.scheduled[index]
		status = "Scheduled for " + item.ScheduledAt.Local().Format("2006-01-02 15:04")
		if item.ScheduleError != nil && *item.ScheduleError != "" {
			status = "Failed · " + ui.SafeInline(*item.ScheduleError)
		}
	case pendingTab:
		item := s.pending[index]
		status = "Undo until " + item.UndoUntil.Local().Format("15:04:05")
	}

	return strings.Join([]string{
		first,
		"   " + s.theme.Dim.Render(ui.Clip(status, max(1, width-3))),
		"   " + s.theme.Dim.Render(ui.Clip(meta, max(1, width-3))),
		s.theme.Dim.Render(strings.Repeat("─", max(0, width-2))),
	}, "\n")
}

func savedTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return "Updated " + value.Local().Format("2006-01-02 15:04")
}

func (s *Screen) draftAt(index int) north.Draft {
	switch s.tab {
	case scheduledTab:
		return s.scheduled[index].Draft
	case pendingTab:
		return s.pending[index].Draft
	default:
		return s.drafts[index]
	}
}

func (s *Screen) count() int {
	switch s.tab {
	case scheduledTab:
		return len(s.scheduled)
	case pendingTab:
		return len(s.pending)
	default:
		return len(s.drafts)
	}
}

func (s *Screen) room(height int) int {
	return max(0, height-pageheader.Height-1)
}
