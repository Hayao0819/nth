package account

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

const accountItemHeight = 3

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	tabs := []string{"Requests", "Blocked", "Muted"}
	if s.keywords != nil {
		tabs = append(tabs, "Words")
	}
	for index := range tabs {
		if tab(index) == s.tab {
			tabs[index] = s.theme.Active.Bold(true).Render(tabs[index])
		} else {
			tabs[index] = s.theme.Dim.Render(tabs[index])
		}
	}
	right := strings.Join(tabs, "  ")
	if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-16)))
	} else if s.loading {
		right = s.theme.Dim.Render("Loading…")
	}
	header := pageheader.Render(s.theme, "Privacy", right, width)
	room := s.room(height)
	body := s.renderItems(width, room)
	footer := "Tab switch · j/k move"
	switch s.tab {
	case requestsTab:
		footer += " · Enter profile · y accept · x reject"
	case blockedTab:
		footer += " · Enter profile · x unblock"
	case mutedTab:
		footer += " · Enter profile · x unmute"
	case keywordsTab:
		footer += " · c add · x delete"
	}
	if s.next != nil {
		footer += " · L more"
	}
	footer += " · Esc/← back"

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Screen) renderItems(width, room int) string {
	if s.count() == 0 {
		message := "Nothing here"
		switch {
		case s.loading:
			message = "Loading…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}

		return ui.Fit(lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message), width, room)
	}
	lines := make([]string, 0, room)
	for index := s.top; index < s.count() && len(lines)+accountItemHeight <= room; index++ {
		lines = append(lines, strings.Split(s.renderItem(index, width, index == s.selected), "\n")...)
	}
	if s.more && len(lines) < room {
		lines = append(lines, s.theme.Dim.Render("  Loading more…"))
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (s *Screen) renderItem(index, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	if s.tab == keywordsTab {
		word := s.words[index]
		targets := make([]string, 0, 2)
		if word.Home {
			targets = append(targets, "home")
		}
		if word.Notifications {
			targets = append(targets, "notifications")
		}
		until := "forever"
		if word.ExpiresAt != nil {
			until = "until " + word.ExpiresAt.Local().Format("2006-01-02 15:04")
		}

		return ui.Clip(marker+"  "+s.theme.Name.Render(ui.SafeInline(word.Word)), width) + "\n" +
			s.theme.Dim.Render(ui.Clip("   "+strings.Join(targets, " · ")+" · "+until, width)) + "\n" +
			s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))
	}
	user := s.users[index]
	name := ui.SafeInline(user.Name)
	if name == "" {
		name = "@" + ui.SafeInline(user.Handle)
	}
	first := marker + "  " + s.theme.Name.Render(name) + "  " + s.theme.Handle.Render("@"+strings.TrimPrefix(ui.SafeInline(user.Handle), "@"))
	meta := fmt.Sprintf("   %d followers · %d following", user.FollowerCount, user.FollowingCount)

	return ui.Clip(first, width) + "\n" + s.theme.Dim.Render(ui.Clip(meta, width)) + "\n" + s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))
}

func (s *Screen) count() int {
	if s.tab == keywordsTab {
		return len(s.words)
	}

	return len(s.users)
}

func (s *Screen) room(height int) int {
	return max(0, height-pageheader.Height-1)
}
