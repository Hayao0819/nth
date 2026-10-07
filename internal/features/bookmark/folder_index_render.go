package bookmark

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const folderHeight = 3

func (s *FolderIndex) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	room := s.room(height)
	lines := make([]string, 0, room)
	for index := s.top; index < len(s.folders) && len(lines)+folderHeight <= room; index++ {
		lines = append(lines, strings.Split(s.renderFolder(s.folders[index], width, index == s.selected), "\n")...)
	}
	if len(s.folders) == 0 {
		message := "No bookmark folders"
		switch {
		case s.loading:
			message = "Loading folders…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}
		lines = []string{lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message)}
	}
	right := ""
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-12)))
	}
	header := pageheader.Render(s.theme, "Bookmark folders", right, width)
	body := ui.Fit(strings.Join(lines, "\n"), width, room)
	footer := s.theme.Dim.Render(ui.Clip("j/k move · Enter open · n new · e rename · d delete · Esc/← back", width))

	return ui.Fit(header+"\n"+body+"\n"+footer, width, height)
}

func (s *FolderIndex) renderFolder(folder bookmarkdomain.BookmarkFolder, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	name := marker + "  " + s.theme.Name.Render(ui.SafeInline(folder.Name))
	if folder.Color != "" {
		name = ui.Sides(name, s.theme.Dim.Render(ui.SafeInline(folder.Color)), width)
	}
	count := fmt.Sprintf("   %d bookmarks", folder.Count)
	divider := s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))

	return strings.Join([]string{name, s.theme.Dim.Render(count), divider}, "\n")
}

func (s *FolderIndex) room(height int) int {
	return max(0, height-pageheader.Height-1)
}

func (s *FolderIndex) move(delta, room int) {
	if len(s.folders) == 0 {
		return
	}
	s.selected = min(max(0, s.selected+delta), len(s.folders)-1)
	s.ensureVisible(room)
}

func (s *FolderIndex) ensureVisible(room int) {
	if s.selected < s.top {
		s.top = s.selected
	}
	visible := max(1, room/folderHeight)
	if s.selected >= s.top+visible {
		s.top = s.selected - visible + 1
	}
}

func (s *FolderIndex) clampSelection() {
	if len(s.folders) == 0 {
		s.selected, s.top = 0, 0

		return
	}
	s.selected = min(s.selected, len(s.folders)-1)
	s.top = min(s.top, s.selected)
}
