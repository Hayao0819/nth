package post

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/navigation"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/charmbracelet/x/ansi"
)

func (d *Screen) actionRow(width int) int {
	if width <= 0 {
		return -1
	}

	lines, _, _ := d.buildContent(width, false)

	return len(lines) - 1
}

func (d *Screen) metadataLines(width int) []string {
	if width <= 0 {
		return nil
	}
	target := d.post.DisplayPost()
	if target == nil {
		target = &d.post
	}

	when := "Unknown time"
	if !target.CreatedAt.IsZero() {
		when = target.CreatedAt.Local().Format("15:04 · 2006-01-02")
	}
	if target.Source != "" {
		when += " · " + ui.SafeInline(target.Source)
	}
	lines := make([]string, 0, 4)
	for _, line := range ui.WrappedLines(when, width) {
		lines = append(lines, d.theme.Dim.Render(line))
	}
	if target.EditedAt != nil {
		lines = append(lines, d.theme.Dim.Render("Edited "+target.EditedAt.Local().Format("15:04 · 2006-01-02")))
	}
	metrics := []string{
		fmt.Sprintf("%d Replies", target.ReplyCount),
		fmt.Sprintf("%d Reposts", target.RepostCount),
		fmt.Sprintf("%d Quotes", target.QuoteCount),
		fmt.Sprintf("%d Likes", target.LikeCount),
	}
	joined := strings.Join(metrics, "  ·  ")
	if lipgloss.Width(joined) <= width {
		lines = append(lines, d.theme.Dim.Render(joined))
	} else {
		lines = append(lines,
			d.theme.Dim.Render(ui.Columns(metrics[:2], width)),
			d.theme.Dim.Render(ui.Columns(metrics[2:], width)),
		)
	}

	return lines
}

func (d *Screen) activityAtPosition(x, row, width int) (navigation.PostActivity, bool) {
	if row < 0 || width <= 0 {
		return 0, false
	}
	lines := d.contentLines(width)
	if row >= len(lines) {
		return 0, false
	}
	plain := ansi.Strip(lines[row])
	if d.historyEnabled && strings.HasPrefix(plain, "Edited ") {
		return navigation.PostHistory, true
	}
	if !d.activityEnabled {
		return 0, false
	}
	if strings.Contains(plain, "Replies") && strings.Contains(plain, "Reposts") && strings.Contains(plain, "Quotes") && strings.Contains(plain, "Likes") {
		switch ui.ColumnAt(x, width, 4) {
		case 1:
			return navigation.PostReposts, true
		case 2:
			return navigation.PostQuotes, true
		case 3:
			return navigation.PostLikes, true
		}
	}
	if strings.Contains(plain, "Replies") && strings.Contains(plain, "Reposts") {
		if ui.ColumnAt(x, width, 2) == 1 {
			return navigation.PostReposts, true
		}
	}
	if strings.Contains(plain, "Quotes") && strings.Contains(plain, "Likes") {
		if ui.ColumnAt(x, width, 2) == 0 {
			return navigation.PostQuotes, true
		}

		return navigation.PostLikes, true
	}

	return 0, false
}

func (d *Screen) footer(width int) string {
	target := d.post.DisplayPost()
	if target == nil {
		target = &d.post
	}
	if !postcomponent.CanInteract(&d.post) {
		return d.theme.Dim.Render(ui.Left("Actions unavailable", width))
	}
	actions := d.actions()
	labels := make([]string, 0, len(actions))
	for _, action := range actions {
		switch action {
		case postcomponent.Reply:
			labels = append(labels, "r  Reply")
		case postcomponent.Repost:
			if target.Reposted {
				labels = append(labels, "t  Undo")
			} else {
				labels = append(labels, "t  Repost")
			}
		case postcomponent.Like:
			if target.Liked {
				labels = append(labels, "l  Unlike")
			} else {
				labels = append(labels, "l  Like")
			}
		case postcomponent.Quote:
			labels = append(labels, "Q  Quote")
		case postcomponent.Bookmark:
			if target.Bookmarked {
				labels = append(labels, "b  Remove")
			} else {
				labels = append(labels, "b  Bookmark")
			}
		case postcomponent.Edit:
			labels = append(labels, "e  Edit")
		case postcomponent.Delete:
			labels = append(labels, "d  Delete")
		}
	}

	styled := make([]string, 0, len(labels))
	for index, label := range labels {
		action := actions[index]
		if action == postcomponent.Repost && target.Reposted || action == postcomponent.Like && target.Liked || action == postcomponent.Bookmark && target.Bookmarked {
			label = d.theme.Active.Render(label)
		} else if action == postcomponent.Delete {
			label = d.theme.Bad.Render(label)
		} else {
			label = d.theme.Heading.Render(label)
		}
		styled = append(styled, label)
	}

	return ui.Columns(styled, width)
}
