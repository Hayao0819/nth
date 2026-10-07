package list

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const entryHeight = 4

func (s *Index) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	room := s.room(height)
	lines := make([]string, 0, room)
	for index := s.top; index < len(s.entries) && len(lines)+entryHeight <= room; index++ {
		lines = append(lines, strings.Split(s.renderEntry(s.entries[index], width, index == s.selected), "\n")...)
	}
	if len(s.entries) == 0 {
		message := "No lists"
		switch {
		case s.loading:
			message = "Loading lists…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}
		lines = []string{lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message)}
	}

	right := ""
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-12)))
	}
	header := pageheader.Render(s.theme, "Lists", right, width)
	footerText := "j/k move · Enter open · u owner · f follow · p pin"
	if s.editor != nil {
		footerText += " · c create"
	}
	footer := s.theme.Dim.Render(ui.Clip(footerText+" · Esc/← back", width))
	body := ui.Fit(strings.Join(lines, "\n"), width, room)

	return ui.Fit(header+"\n"+body+"\n"+footer, width, height)
}

func (s *Index) renderEntry(entry entry, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	title := marker + "  " + s.theme.Name.Render(ui.SafeInline(entry.item.Name))
	badges := make([]string, 0, 2)
	if entry.item.PinnedByViewer {
		badges = append(badges, "pinned")
	}
	if entry.item.Private {
		badges = append(badges, "private")
	}
	if len(badges) > 0 {
		title = ui.Sides(title, s.theme.Dim.Render(strings.Join(badges, " · ")), width)
	} else {
		title = ui.Clip(title, width)
	}
	description := ""
	if entry.item.Description != nil {
		description = ui.SafeInline(*entry.item.Description)
	}
	if description == "" {
		description = "by @" + ui.SafeInline(entry.item.Owner.Handle)
	}
	meta := fmt.Sprintf("%s · %d members · %d followers", entry.section, entry.item.MemberCount, entry.item.FollowerCount)
	if entry.item.FollowedByViewer && !entry.item.OwnedByViewer {
		meta += " · following"
	}
	divider := s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))

	return strings.Join([]string{
		title,
		"   " + s.theme.Dim.Render(ui.Clip(description, max(1, width-3))),
		"   " + s.theme.Dim.Render(ui.Clip(meta, max(1, width-3))),
		divider,
	}, "\n")
}

func (s *Index) room(height int) int {
	return max(0, height-pageheader.Height-1)
}

func (s *Index) move(delta, room int) {
	if len(s.entries) == 0 {
		return
	}
	s.selected = min(max(0, s.selected+delta), len(s.entries)-1)
	s.ensureVisible(room)
}

func (s *Index) ensureVisible(room int) {
	if s.selected < s.top {
		s.top = s.selected
	}
	visible := max(1, room/entryHeight)
	if s.selected >= s.top+visible {
		s.top = s.selected - visible + 1
	}
}

func (s *Index) entryAt(row int) (int, bool) {
	if row < 0 {
		return 0, false
	}
	index := s.top + row/entryHeight
	if index < s.top || index >= len(s.entries) {
		return 0, false
	}

	return index, true
}
