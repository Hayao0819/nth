package post

import (
	"strings"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
)

func Concealed(post *north.Post) bool {
	return post == nil || post.Deleted || post.Unavailable || post.HiddenReason != ""
}

func ConcealmentLabel(post *north.Post) string {
	switch {
	case post == nil || post.Unavailable:
		return "Post unavailable"
	case post.Deleted:
		return "Post deleted"
	case post.HiddenReason != "":
		return "Hidden: " + string(post.HiddenReason)
	default:
		return ""
	}
}

func RenderCard(post north.Post, width int, selected bool, theme ui.Theme, now time.Time) string {
	return RenderCardWithImages(post, width, selected, theme, now, nil)
}

func RenderCardWithImages(post north.Post, width int, selected bool, theme ui.Theme, now time.Time, images *termimage.Renderer) string {
	if width <= 0 {
		return ""
	}
	contentWidth := max(1, width-2)
	target := post.DisplayPost()
	if target == nil {
		target = &post
	}

	lines := make([]string, 0, 16)
	if post.RepostOf != nil {
		lines = append(lines, theme.Dim.Render("  ↻ @"+ui.SafeInline(post.Author.Handle)+" reposted"))
	}
	cursorLine := len(lines)

	badges := ""
	if target.Author.Verified {
		badges += " ✓"
	}
	if target.Author.Protected {
		badges += " 🔒"
	}
	when := RelativeTime(target.CreatedAt, now)
	if target.EditedAt != nil {
		when += " · edited"
	}
	avatar := theme.Accent.Render("●") + " "
	if target.Author.AvatarURL != nil {
		if rendered := images.View(*target.Author.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows); rendered != "" {
			avatar = rendered
		}
	}
	identity := avatar + " " + theme.Name.Render(ui.SafeInline(target.Author.Name)+badges) + " " +
		theme.Handle.Render("@"+ui.SafeInline(target.Author.Handle))
	header := ui.Sides(identity, theme.Dim.Render(when), contentWidth)
	lines = append(lines, header)

	if target.InReplyToHandle != nil && *target.InReplyToHandle != "" {
		lines = append(lines, "  "+theme.Dim.Render("Replying to @"+ui.SafeInline(*target.InReplyToHandle)))
	}
	concealed := Concealed(target)
	if target.Deleted {
		lines = append(lines, "  "+theme.Dim.Render("Post deleted"))
	} else if target.HiddenReason != "" {
		lines = append(lines, "  "+theme.Warn.Render("Hidden: "+ui.SafeInline(string(target.HiddenReason))))
	} else if target.Unavailable {
		lines = append(lines, "  "+theme.Dim.Render("Post unavailable"))
	}
	if !concealed {
		body := ui.WrappedLines(target.Text, max(1, contentWidth-2))
		for _, line := range body {
			lines = append(lines, "  "+line)
		}
		if len(body) == 0 && len(target.Media) == 0 {
			lines = append(lines, "  "+theme.Dim.Render("No text"))
		}
	}

	if !concealed {
		for _, media := range target.Media {
			body := cardMediaBody(media, contentWidth, images)
			for _, line := range RenderBox(MediaLabel(media), body, max(4, contentWidth-2)) {
				lines = append(lines, "  "+theme.Dim.Render(line))
			}
		}
	}

	if !concealed && target.Quoted != nil {
		quoted := target.Quoted.DisplayPost()
		quoteWidth := max(4, contentWidth-2)
		title := ui.SafeInline(quoted.Author.Name) + " @" + ui.SafeInline(quoted.Author.Handle) + " · " + RelativeTime(quoted.CreatedAt, now)
		body := ui.WrappedLines(quoted.Text, max(1, quoteWidth-4))
		if Concealed(quoted) {
			body = ui.WrappedLines(ConcealmentLabel(quoted), max(1, quoteWidth-4))
		}
		for _, line := range RenderBox(title, body, quoteWidth) {
			lines = append(lines, "  "+theme.Dim.Render(line))
		}
	} else if !concealed && target.QuotedUnavailable {
		lines = append(lines, "  "+theme.Dim.Render("Quoted post unavailable"))
	}

	if !concealed && target.Poll != nil {
		lines = append(lines, RenderPoll(target.Poll, max(1, contentWidth-2), theme)...)
	}

	replyText := theme.Dim.Render(metricLabel("↩", target.ReplyCount))
	repostText := metricLabel("↻", target.RepostCount)
	if target.Reposted {
		repostText = theme.Active.Render(repostText)
	} else {
		repostText = theme.Dim.Render(repostText)
	}
	likeText := metricLabel("♡", target.LikeCount)
	if target.Liked {
		likeText = theme.Active.Render(metricLabel("♥", target.LikeCount))
	} else {
		likeText = theme.Dim.Render(likeText)
	}
	quoteText := theme.Dim.Render(metricLabel("❝", target.QuoteCount))
	bookmarkText := theme.Dim.Render("♧")
	if target.Bookmarked {
		bookmarkText = theme.Active.Render("♣")
	}
	lines = append(lines, "  "+ui.Columns([]string{replyText, repostText, likeText, quoteText, bookmarkText}, max(1, contentWidth-2)))
	lines = append(lines, "  "+theme.Dim.Render(strings.Repeat("─", max(0, contentWidth-4))))

	for index := range lines {
		marker := " "
		if selected && index == cursorLine {
			marker = theme.Active.Render(">")
		}
		lines[index] = marker + " " + ui.Left(lines[index], contentWidth)
	}

	return strings.Join(lines, "\n")
}
