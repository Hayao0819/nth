package list

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const itemHeight = 3

func (s *Members) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	room := s.room(height)
	title := ui.SafeInline(s.list.Name)
	if title == "" {
		title = "List"
	}
	right := s.theme.Dim.Render(fmt.Sprintf("%d members", len(s.users)))
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-len(title)-8)))
	}
	header := pageheader.Render(s.theme, title+" · Members", right, width)
	body := s.renderItems(width, room)
	footer := "j/k move · Enter/u profile"
	if s.list.OwnedByViewer {
		footer += " · a add · x remove"
	}
	if s.next != nil {
		footer += " · L more"
	}
	footer += " · Esc/← back"

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Members) renderItems(width, room int) string {
	if len(s.users) == 0 {
		message := "No members"
		switch {
		case s.loading:
			message = "Loading members…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}

		return ui.Fit(lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message), width, room)
	}
	lines := make([]string, 0, room)
	for index := s.top; index < len(s.users) && len(lines)+itemHeight <= room; index++ {
		lines = append(lines, strings.Split(s.renderUser(s.users[index], width, index == s.selected), "\n")...)
	}
	if s.more && len(lines) < room {
		lines = append(lines, s.theme.Dim.Render("  Loading more…"))
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (s *Members) renderUser(user north.User, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	name := ui.SafeInline(user.Name)
	if name == "" {
		name = "@" + ui.SafeInline(user.Handle)
	}
	first := marker + "  " + s.theme.Name.Render(name) + "  " + s.theme.Handle.Render("@"+strings.TrimPrefix(ui.SafeInline(user.Handle), "@"))
	meta := fmt.Sprintf("   %d followers · %d following", user.FollowerCount, user.FollowingCount)

	return ui.Clip(first, width) + "\n" + s.theme.Dim.Render(ui.Clip(meta, width)) + "\n" + s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))
}

func (s *Members) room(height int) int {
	return max(0, height-pageheader.Height-1)
}
