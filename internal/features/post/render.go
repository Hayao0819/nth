package post

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (d *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	innerWidth, room := d.layout(width, height)
	lines := d.contentLines(innerWidth)
	offset := clampOffset(d.offset, len(lines), room)
	end := min(len(lines), offset+room)

	profileHint := "u profile"
	if len(postcomponent.LinkedUsers(d.post, d.ancestors...)) > 1 {
		profileHint = "u/U profiles"
	}
	if len(d.ancestors) > 0 {
		profileHint = "p parent · " + profileHint
	}
	if target := d.post.DisplayPost(); target != nil && canVote(target.Poll) && d.polls != nil {
		profileHint = "v vote · " + profileHint
	}
	if d.activityEnabled {
		profileHint = "V/I/T activity · " + profileHint
	}
	if d.historyEnabled {
		profileHint = "H history · " + profileHint
	}
	titleRight := d.theme.Dim.Render(profileHint)
	if d.notice != "" {
		titleRight = d.theme.Dim.Render(ui.Clip(d.notice, max(1, innerWidth-8)))
	} else if len(lines) > room {
		titleRight = d.theme.Dim.Render(fmt.Sprintf("%d–%d / %d", offset+1, end, len(lines)))
	}
	title := "Post"
	if replyParentID(d.post) != "" || len(d.ancestors) > 0 {
		title = "Conversation"
	}
	titleRight += " · " + d.theme.Dim.Render("Esc/← back")
	header := pageheader.Render(d.theme, title, titleRight, innerWidth)
	body := ""
	if offset < end {
		body = strings.Join(lines[offset:end], "\n")
	}
	sections := []string{header}
	if room > 0 {
		sections = append(sections, ui.Fit(body, innerWidth, room))
	}
	content := strings.Join(sections, "\n")

	return ui.Fit(content, width, height)
}

func (d *Screen) layout(width, height int) (int, int) {
	innerWidth := max(0, width)
	innerHeight := max(0, height)

	return innerWidth, max(0, innerHeight-pageheader.Height)
}

func clampOffset(offset, lines, room int) int {
	return min(max(0, offset), max(0, lines-room))
}

func (d *Screen) contentLines(width int) []string {
	lines, _, _ := d.content(width)

	return lines
}

func (d *Screen) content(width int) ([]string, []userHit, []postHit) {
	return d.buildContent(width, true)
}

func (d *Screen) buildContent(width int, includeReplies bool) ([]string, []userHit, []postHit) {
	if width <= 0 {
		return nil, nil, nil
	}
	target := d.post.DisplayPost()
	if target == nil {
		target = &d.post
	}

	thread, hits, postHits := d.threadContent(width)
	lines := make([]string, 0, len(thread)+24)
	lines = append(lines, thread...)
	if d.post.RepostOf != nil {
		label := ui.SafeInline(d.post.Author.Name) + " @" + ui.SafeInline(d.post.Author.Handle)
		if hit, ok := makeUserHit(d.post.Author, len(lines), lipgloss.Width("↻ Reposted by "), label, width); ok {
			hits = append(hits, hit)
		}
		lines = append(lines, d.theme.Dim.Render("↻ Reposted by "+ui.SafeInline(d.post.Author.Name)+" @"+ui.SafeInline(d.post.Author.Handle)), "")
	}

	badges := ""
	if target.Author.Verified {
		badges += " ✓"
	}
	if target.Author.Protected {
		badges += " 🔒"
	}
	name := ui.SafeInline(target.Author.Name) + badges
	if hit, ok := makeUserHit(target.Author, len(lines), termimage.AvatarColumns+1, name, width); ok {
		hits = append(hits, hit)
	}
	if hit, ok := makeUserHit(target.Author, len(lines)+1, termimage.AvatarColumns+1, "@"+ui.SafeInline(target.Author.Handle), width); ok {
		hits = append(hits, hit)
	}
	avatar := d.theme.Accent.Render("●") + " "
	if target.Author.AvatarURL != nil {
		if rendered := d.images.View(*target.Author.AvatarURL, termimage.AvatarColumns, termimage.AvatarRows); rendered != "" {
			avatar = rendered
		}
	}
	lines = append(lines,
		avatar+" "+d.theme.Name.Render(name),
		d.theme.Handle.Render(strings.Repeat(" ", termimage.AvatarColumns+1)+"@"+ui.SafeInline(target.Author.Handle)),
	)
	if target.InReplyToHandle != nil && *target.InReplyToHandle != "" {
		replyUser := north.User{Handle: strings.TrimPrefix(*target.InReplyToHandle, "@")}
		label := "@" + ui.SafeInline(*target.InReplyToHandle)
		if hit, ok := makeUserHit(replyUser, len(lines), lipgloss.Width("  Replying to "), label, width); ok {
			hits = append(hits, hit)
		}
		lines = append(lines, d.theme.Dim.Render("  Replying to @"+ui.SafeInline(*target.InReplyToHandle)))
	}
	lines = append(lines, "")

	concealed := postcomponent.Concealed(target)
	switch {
	case target.Deleted:
		lines = append(lines, d.theme.Dim.Render("Post deleted"))
	case target.HiddenReason != "":
		lines = append(lines, d.theme.Warn.Render("Hidden: "+ui.SafeInline(string(target.HiddenReason))))
	case target.Unavailable:
		lines = append(lines, d.theme.Dim.Render("This post is unavailable"))
	default:
		lines = append(lines, ui.WrappedLines(target.Text, width)...)
		if target.Text == "" && len(target.Media) == 0 {
			lines = append(lines, d.theme.Dim.Render("No text"))
		}
	}

	if !concealed {
		for _, media := range target.Media {
			body := make([]string, 0, 3)
			columns, rows := DetailMediaSize(width)
			rendered := ""
			if media.Kind == north.MediaPhoto || media.Kind == north.MediaGIF {
				rendered = d.images.View(postcomponent.MediaPreviewURL(media), columns, rows)
				if rendered != "" {
					body = append(body, strings.Split(rendered, "\n")...)
				}
			}
			if media.AltText != nil && *media.AltText != "" {
				body = append(body, ui.WrappedLines("Alt: "+*media.AltText, max(1, width-4))...)
			}
			if rendered == "" && media.URL != "" {
				body = append(body, ui.Clip(ui.SafeInline(media.URL), max(1, width-4)))
			}
			lines = append(lines, "")
			for _, line := range postcomponent.RenderBox(postcomponent.MediaLabel(media), body, width) {
				lines = append(lines, d.theme.Dim.Render(line))
			}
		}
	}
	if !concealed && target.Quoted != nil {
		quoted := target.Quoted.DisplayPost()
		body := ui.WrappedLines(quoted.Text, max(1, width-4))
		if postcomponent.Concealed(quoted) {
			body = ui.WrappedLines(postcomponent.ConcealmentLabel(quoted), max(1, width-4))
		}
		title := ui.SafeInline(quoted.Author.Name) + " @" + ui.SafeInline(quoted.Author.Handle)
		lines = append(lines, "")
		quoteTop := len(lines)
		if hit, ok := makeUserHit(quoted.Author, len(lines), 3, title, width); ok {
			hits = append(hits, hit)
		}
		for _, line := range postcomponent.RenderBox(title, body, width) {
			lines = append(lines, d.theme.Dim.Render(line))
		}
		if !postcomponent.Concealed(quoted) {
			postHits = append(postHits, postHit{post: *quoted, top: quoteTop, bottom: len(lines) - 1, right: width})
		}
	} else if !concealed && target.QuotedUnavailable {
		lines = append(lines, "", d.theme.Dim.Render("Quoted post unavailable"))
	}
	if !concealed && target.Poll != nil {
		lines = append(lines, "")
		lines = append(lines, postcomponent.RenderPoll(target.Poll, max(1, width-2), d.theme)...)
		if !target.Poll.EndsAt.IsZero() {
			lines = append(lines, "  "+d.theme.Dim.Render("Ends "+target.Poll.EndsAt.Local().Format("2006-01-02 15:04")))
		}
	}
	lines = append(lines, "")
	lines = append(lines, d.metadataLines(width)...)
	lines = append(lines,
		d.theme.Dim.Render(strings.Repeat("─", width)),
		d.footer(width),
	)
	if includeReplies {
		lines, hits, postHits = d.appendReplyContent(lines, hits, postHits, width)
	}

	return lines, hits, postHits
}
