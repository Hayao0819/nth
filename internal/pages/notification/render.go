package notification

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	notificationdomain "github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (d *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	innerWidth := d.innerWidth(width)
	room := d.room(height)
	lines := make([]string, 0, room)
	used := 0
	for index := d.top; index < len(d.items); index++ {
		card := d.renderItem(d.items[index], innerWidth, index == d.selected)
		cardHeight := lipgloss.Height(card)
		if used+cardHeight > room {
			break
		}
		lines = append(lines, card)
		used += cardHeight
	}
	if len(d.items) == 0 {
		message := "No notifications"
		switch {
		case d.loading:
			message = "Loading notifications…"
		case d.err != nil:
			message = ui.FriendlyError(d.err) + "\n\nPress . to try again"
		}
		lines = []string{lipgloss.Place(innerWidth, max(1, room), lipgloss.Center, lipgloss.Center, message)}
	}
	if d.loadingMore && used < room {
		lines = append(lines, d.theme.Dim.Render("  Loading more…"))
	}

	right := ""
	if d.notice != "" {
		right = d.theme.Warn.Render(ui.Clip(d.notice, max(1, innerWidth-8)))
	}
	header := pageheader.Render(d.theme, "Notifications", right, innerWidth)
	footer := "j/k move · Enter open · u profile · Esc/← back"
	if d.nextCursor != nil {
		footer += " · L more"
	}
	body := ui.Fit(strings.Join(lines, "\n"), innerWidth, room)
	footer = d.theme.Dim.Render(ui.Clip(footer, innerWidth))
	return ui.Fit(header+"\n"+body+"\n"+footer, width, height)
}

func (d *Screen) room(height int) int {
	return max(0, height-pageheader.Height-1)
}

func (d *Screen) innerWidth(width int) int {
	return max(0, width)
}

func (d *Screen) bodyTop() int {
	return pageheader.Height
}

func (d *Screen) renderItem(item notificationdomain.Item, width int, selected bool) string {
	actor := "Someone"
	if len(item.Actors) > 0 {
		actor = ui.SafeInline(item.Actors[0].Name)
		if actor == "" && strings.TrimSpace(item.Actors[0].Handle) != "" {
			actor = "@" + ui.SafeInline(item.Actors[0].Handle)
		} else if actor == "" {
			actor = "Someone"
		}
	}
	count := max(item.ActorCount, len(item.Actors))
	if count > 1 {
		actor += fmt.Sprintf(" +%d", count-1)
	}
	action := notificationLabel(item.Kind)
	when := relativeTime(item.CreatedAt, time.Now())
	marker := "○"
	if !item.Read {
		marker = d.theme.Active.Render("●")
	}
	if selected {
		marker = d.theme.Active.Render(">")
	}
	avatar := ""
	if len(item.Actors) > 0 && item.Actors[0].AvatarURL != nil {
		avatar = d.theme.Accent.Render("●") + " "
		if rendered := d.images.View(*item.Actors[0].AvatarURL, termimage.AvatarColumns, termimage.AvatarRows); rendered != "" {
			avatar = rendered
		}
		avatar += " "
	}
	lines := []string{marker + "  " + avatar + d.theme.Name.Render(actor) + " " + d.theme.Heading.Render(action)}
	if when != "" {
		lines[0] = ui.Sides(lines[0], d.theme.Dim.Render(when), width)
	} else {
		lines[0] = ui.Clip(lines[0], width)
	}
	if item.Post != nil {
		target := item.Post.DisplayPost()
		preview := ""
		if postcomponent.Concealed(target) {
			preview = postcomponent.ConcealmentLabel(target)
		} else if target != nil {
			preview = target.Text
			if preview == "" && len(target.Media) > 0 {
				preview = "[media]"
			}
		}
		previewLines := ui.WrappedLines(preview, max(1, width-4))
		for _, line := range previewLines[:min(2, len(previewLines))] {
			lines = append(lines, "   "+line)
		}
	}
	lines = append(lines, d.theme.Dim.Render(strings.Repeat("─", max(0, width-2))))

	return strings.Join(lines, "\n")
}

func notificationLabel(kind notificationdomain.Kind) string {
	switch kind {
	case notificationdomain.Follow:
		return "followed you"
	case notificationdomain.Like:
		return "liked your post"
	case notificationdomain.Repost:
		return "reposted your post"
	case notificationdomain.Post:
		return "posted"
	case notificationdomain.Reply:
		return "replied to you"
	case notificationdomain.Quote:
		return "quoted your post"
	case notificationdomain.Mention:
		return "mentioned you"
	default:
		return "sent a notification"
	}
}

func relativeTime(when, now time.Time) string {
	if when.IsZero() {
		return ""
	}
	delta := max(time.Duration(0), now.Sub(when))
	switch {
	case delta < time.Minute:
		return "now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh", int(delta.Hours()))
	default:
		return fmt.Sprintf("%dd", int(delta.Hours()/24))
	}
}
