package post

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
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
	lines = append(lines, "  "+ui.Columns([]string{replyText, repostText, likeText, quoteText}, max(1, contentWidth-2)))
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

func CardCursorRow(post north.Post) int {
	if post.RepostOf != nil {
		return 1
	}

	return 0
}

func CardUserAt(post north.Post, x, row, width int) (north.User, bool) {
	return CardUserAtWithImages(post, x, row, width, nil)
}

func CardUserAtWithImages(post north.Post, x, row, width int, images *termimage.Renderer) (north.User, bool) {
	if width <= 0 || x < 0 || x >= width || row < 0 {
		return north.User{}, false
	}
	target := post.DisplayPost()
	if target == nil {
		return north.User{}, false
	}

	currentRow := 0
	if post.RepostOf != nil {
		left := 6
		if row == currentRow {
			if user, ok := cardUserHit(post.Author, x, left, "@"+ui.SafeInline(post.Author.Handle), width); ok {
				return user, true
			}
		}
		currentRow++
	}

	badges := ""
	if target.Author.Verified {
		badges += " ✓"
	}
	if target.Author.Protected {
		badges += " 🔒"
	}
	label := ui.SafeInline(target.Author.Name) + badges + " @" + ui.SafeInline(target.Author.Handle)
	if row == currentRow {
		if user, ok := cardUserHit(target.Author, x, 3+termimage.AvatarColumns, label, width); ok {
			return user, true
		}
	}
	currentRow++

	if target.InReplyToHandle != nil && *target.InReplyToHandle != "" {
		left := 4 + lipgloss.Width("Replying to ")
		if row == currentRow {
			user := north.User{Handle: *target.InReplyToHandle}
			if user, ok := cardUserHit(user, x, left, "@"+ui.SafeInline(*target.InReplyToHandle), width); ok {
				return user, true
			}
		}
		currentRow++
	}
	concealed := Concealed(target)
	if target.Deleted || target.HiddenReason != "" || target.Unavailable {
		currentRow++
	}
	if !concealed {
		body := ui.WrappedLines(target.Text, max(1, width-4))
		currentRow += len(body)
		if len(body) == 0 && len(target.Media) == 0 {
			currentRow++
		}
		for _, media := range target.Media {
			bodyRows := len(cardMediaBody(media, max(1, width-2), images))
			currentRow += bodyRows + 2
		}
	}

	if !concealed && target.Quoted != nil {
		quoted := target.Quoted.DisplayPost()
		label := ui.SafeInline(quoted.Author.Name) + " @" + ui.SafeInline(quoted.Author.Handle)
		if row == currentRow {
			if user, ok := cardUserHit(quoted.Author, x, 7, label, width); ok {
				return user, true
			}
		}
	}

	return north.User{}, false
}

// CardQuotedPostAt reports whether a position belongs to the quoted post.
// Check CardUserAt first so the quoted author's label remains a user link.
func CardQuotedPostAt(post north.Post, x, row, width int) (north.Post, bool) {
	return CardQuotedPostAtWithImages(post, x, row, width, nil)
}

func CardQuotedPostAtWithImages(post north.Post, x, row, width int, images *termimage.Renderer) (north.Post, bool) {
	quoted, first, last, ok := cardQuotedPostRows(post, width, images)
	if !ok || x < 4 || x >= width || row < first || row > last {
		return north.Post{}, false
	}

	return quoted, true
}

func cardQuotedPostRows(post north.Post, width int, images *termimage.Renderer) (north.Post, int, int, bool) {
	if width <= 0 {
		return north.Post{}, 0, 0, false
	}
	target := post.DisplayPost()
	if target == nil || Concealed(target) || target.Quoted == nil {
		return north.Post{}, 0, 0, false
	}
	quoted := target.Quoted.DisplayPost()
	if Concealed(quoted) {
		return north.Post{}, 0, 0, false
	}

	currentRow := 0
	if post.RepostOf != nil {
		currentRow++
	}
	currentRow++
	if target.InReplyToHandle != nil && *target.InReplyToHandle != "" {
		currentRow++
	}
	body := ui.WrappedLines(target.Text, max(1, width-4))
	currentRow += len(body)
	if len(body) == 0 && len(target.Media) == 0 {
		currentRow++
	}
	for _, media := range target.Media {
		bodyRows := len(cardMediaBody(media, max(1, width-2), images))
		currentRow += bodyRows + 2
	}

	quoteWidth := max(4, width-4)
	quoteBody := ui.WrappedLines(quoted.Text, max(1, quoteWidth-4))
	quoteHeight := len(quoteBody) + 2

	return *quoted, currentRow, currentRow + quoteHeight - 1, true
}

func CardMediaSize(width int) (int, int) {
	return max(1, min(48, width-8)), 6
}

func cardMediaBody(media north.Media, width int, images *termimage.Renderer) []string {
	var body []string
	if media.Kind == north.MediaPhoto || media.Kind == north.MediaGIF {
		columns, rows := CardMediaSize(width + 2)
		if rendered := images.View(MediaPreviewURL(media), columns, rows); rendered != "" {
			body = append(body, strings.Split(rendered, "\n")...)
		}
	}
	if media.AltText != nil && *media.AltText != "" {
		body = append(body, ui.WrappedLines(*media.AltText, max(1, width-6))...)
	}

	return body
}

func cardUserHit(user north.User, x, left int, label string, limit int) (north.User, bool) {
	user.Handle = strings.TrimPrefix(strings.TrimSpace(user.Handle), "@")
	if user.Handle == "" || !hitText(x, left, label, limit) {
		return north.User{}, false
	}

	return user, true
}

func hitText(x, left int, label string, limit int) bool {
	return x >= left && x < min(limit, left+lipgloss.Width(label))
}

func MediaLabel(media north.Media) string {
	kind := "Media"
	switch media.Kind {
	case north.MediaPhoto:
		kind = "Photo"
	case north.MediaGIF:
		kind = "GIF"
	case north.MediaVideo:
		kind = "Video"
	}
	detail := "▣ " + kind
	if media.Width > 0 && media.Height > 0 {
		detail += fmt.Sprintf(" · %d×%d", media.Width, media.Height)
	}
	if media.Sensitive {
		detail += " · sensitive"
	}
	if media.Warning != nil {
		detail += " · " + strings.ToLower(ui.SafeInline(string(*media.Warning)))
	}

	return detail
}

func RenderPoll(poll *north.Poll, width int, theme ui.Theme) []string {
	total := 0
	for _, option := range poll.Options {
		total += option.Votes
	}

	lines := make([]string, 0, len(poll.Options)+1)
	for index, option := range poll.Options {
		marker := "○"
		if poll.Voted != nil && *poll.Voted == index {
			marker = "●"
		}
		percent := 0
		if total > 0 {
			percent = option.Votes * 100 / total
		}
		line := ui.Sides(marker+" "+ui.SafeInline(option.Label), fmt.Sprintf("%d%%", percent), width)
		if poll.Voted != nil && *poll.Voted == index {
			line = theme.Active.Render(line)
		}
		lines = append(lines, "  "+line)
	}
	lines = append(lines, "  "+theme.Dim.Render(fmt.Sprintf("%d votes", total)))

	return lines
}

func RenderBox(title string, body []string, width int) []string {
	width = max(4, width)
	label := "─ " + ui.SafeInline(title) + " "
	label = ui.Clip(label, width-2)
	top := "╭" + label + strings.Repeat("─", max(0, width-2-lipgloss.Width(label))) + "╮"
	lines := []string{top}
	for _, line := range body {
		lines = append(lines, "│"+ui.Left(" "+line, width-2)+"│")
	}
	lines = append(lines, "╰"+strings.Repeat("─", width-2)+"╯")

	return lines
}

func metricLabel(icon string, count int) string {
	if count == 0 {
		return icon
	}

	return fmt.Sprintf("%s %d", icon, count)
}

func RelativeTime(when, now time.Time) string {
	if when.IsZero() {
		return "unknown"
	}
	delta := now.Sub(when)
	if delta < 0 {
		delta = 0
	}
	switch {
	case delta < time.Minute:
		return "now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh", int(delta.Hours()))
	case delta < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(delta.Hours()/24))
	default:
		return when.Local().Format("2006-01-02")
	}
}
