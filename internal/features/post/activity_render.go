package post

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Activity) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	room := s.room(height)
	right := ""
	if s.loading {
		right = s.theme.Dim.Render("Loading…")
	}
	header := pageheader.Render(s.theme, s.title(), right, width)
	body := s.renderItems(width, room)
	footer := "j/k move · Enter open · Esc/← back"
	if s.kind == navigation.PostQuotes {
		footer = "j/k move · Enter open · r/t/l/Q/b actions · Esc/← back"
	}
	if s.next != nil {
		footer += " · L more"
	}

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Activity) renderItems(width, room int) string {
	if s.itemCount() == 0 {
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
	used := 0
	for index := s.top; index < s.itemCount(); index++ {
		item := s.renderItem(index, width, index == s.selected)
		height := lipgloss.Height(item)
		if used+height > room {
			break
		}
		lines = append(lines, item)
		used += height
	}
	if s.more && used < room {
		lines = append(lines, s.theme.Dim.Render("  Loading more…"))
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (s *Activity) renderItem(index, width int, selected bool) string {
	switch s.kind {
	case navigation.PostQuotes:
		return postcomponent.RenderCardWithImages(s.posts[index], width, selected, s.theme, time.Now(), s.images)
	case navigation.PostHistory:
		return s.renderVersion(s.versions[index], width, selected)
	default:
		return s.renderUser(s.users[index], width, selected)
	}
}

func (s *Activity) renderUser(user north.User, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	name := ui.SafeInline(user.Name)
	if name == "" {
		name = "@" + ui.SafeInline(user.Handle)
	}
	prefix := marker + "  "
	if user.AvatarURL != nil {
		if avatar := s.images.View(*user.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows); avatar != "" {
			prefix += avatar + " "
		}
	}
	identity := prefix + s.theme.Name.Render(name) + "  " + s.theme.Handle.Render("@"+strings.TrimPrefix(ui.SafeInline(user.Handle), "@"))
	meta := fmt.Sprintf("   %d followers · %d following", user.FollowerCount, user.FollowingCount)
	divider := s.theme.Dim.Render(strings.Repeat("─", max(0, width-2)))

	return ui.Clip(identity, width) + "\n" + s.theme.Dim.Render(ui.Clip(meta, width)) + "\n" + divider
}

func (s *Activity) renderVersion(version north.PostVersion, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render(">")
	}
	when := version.CreatedAt.Local().Format("15:04 · 2006-01-02")
	if version.Current {
		when += " · current"
	}
	lines := []string{marker + "  " + s.theme.Dim.Render(when)}
	for _, line := range ui.WrappedLines(version.Text, max(1, width-3)) {
		lines = append(lines, "   "+line)
	}
	if len(version.Media) > 0 {
		lines = append(lines, "   "+s.theme.Dim.Render(fmt.Sprintf("%d media", len(version.Media))))
	}
	lines = append(lines, s.theme.Dim.Render(strings.Repeat("─", max(0, width-2))))

	return strings.Join(lines, "\n")
}

func (s *Activity) title() string {
	switch s.kind {
	case navigation.PostLikes:
		return "Likes"
	case navigation.PostReposts:
		return "Reposts"
	case navigation.PostHistory:
		return "Edit history"
	default:
		return "Quotes"
	}
}

func (s *Activity) room(height int) int {
	return max(0, height-pageheader.Height-1)
}
