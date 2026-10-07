package user

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Connections) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	room := s.room(height)
	title := "Followers"
	if s.following {
		title = "Following"
	}
	right := ""
	if s.loading {
		right = s.theme.Dim.Render("Loading…")
	}
	header := pageheader.Render(s.theme, title, right, width)
	body := s.renderItems(width, room)
	footer := "j/k move · Enter open · Esc/← back"
	if s.next != nil {
		footer += " · L more"
	}

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Connections) renderItems(width, room int) string {
	if len(s.items) == 0 {
		message := "No accounts"
		switch {
		case s.loading:
			message = "Loading accounts…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}

		return ui.Fit(lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message), width, room)
	}
	lines := make([]string, 0, room)
	used := 0
	for index := s.top; index < len(s.items); index++ {
		card := s.renderUser(s.items[index], width, index == s.selected)
		cardHeight := lipgloss.Height(card)
		if used+cardHeight > room {
			break
		}
		lines = append(lines, card)
		used += cardHeight
	}
	if s.more && used < room {
		lines = append(lines, s.theme.Dim.Render("  Loading more…"))
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (s *Connections) renderUser(user north.User, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	name := ui.SafeInline(user.Name)
	if name == "" {
		name = "@" + ui.SafeInline(user.Handle)
	}
	identity := marker + "  " + s.theme.Name.Render(name) + "  " + s.theme.Handle.Render("@"+strings.TrimPrefix(ui.SafeInline(user.Handle), "@"))
	meta := fmt.Sprintf("   %d followers · %d following", user.FollowerCount, user.FollowingCount)
	if user.Bio != nil && strings.TrimSpace(*user.Bio) != "" {
		meta = "   " + ui.SafeInline(*user.Bio)
	}
	divider := s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))

	return ui.Clip(identity, width) + "\n" + s.theme.Dim.Render(ui.Clip(meta, width)) + "\n" + divider
}
