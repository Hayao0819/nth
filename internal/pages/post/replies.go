package postpage

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	conversationdomain "github.com/Hayao0819/nth/internal/domain/conversation"
	"github.com/Hayao0819/nth/internal/support/collection"
	"github.com/Hayao0819/nth/internal/ui"
)

type conversationLoadedMsg struct {
	target *Screen
	page   conversationdomain.Page
	resp   *north.Response
	err    error
	more   bool
}

func (m conversationLoadedMsg) Response() *north.Response { return m.resp }

func (d *Screen) loadConversation(ctx context.Context, more bool) tea.Cmd {
	if d.conversation == nil || d.loadingConversation {
		return nil
	}
	id := conversationPostID(d.post)
	if id == "" {
		return nil
	}
	cursor := ""
	if more {
		if d.nextReplyCursor == nil || strings.TrimSpace(*d.nextReplyCursor) == "" {
			return nil
		}
		cursor = *d.nextReplyCursor
	}
	d.loadingConversation = true
	d.loadingMoreReplies = more
	d.replyErr = nil

	return func() tea.Msg {
		page, response, err := d.conversation.PostConversation(ctx, id, cursor)

		return conversationLoadedMsg{target: d, page: page, resp: response, err: err, more: more}
	}
}

func conversationPostID(post north.Post) string {
	target := post.DisplayPost()
	if target != nil && strings.TrimSpace(target.ID) != "" {
		return strings.TrimSpace(target.ID)
	}

	return strings.TrimSpace(post.ID)
}

func appendUniqueReplies(current, additional []north.Post) []north.Post {
	return collection.AppendUniqueBy(current, additional, conversationPostID)
}

func (d *Screen) appendReplyContent(lines []string, users []userHit, posts []postHit, width int) ([]string, []userHit, []postHit) {
	lines = append(lines, "", d.theme.Section.Render("Replies"), "")
	switch {
	case d.conversation == nil:
		message := "No replies yet"
		if target := d.post.DisplayPost(); target != nil && target.ReplyCount > 0 {
			message = "Browser authentication is required to load replies"
		}
		lines = append(lines, d.theme.Dim.Render(message))
	case d.loadingConversation && len(d.replies) == 0:
		lines = append(lines, d.theme.Dim.Render("Loading replies…"))
	case d.replyErr != nil && len(d.replies) == 0:
		message := "Replies unavailable: " + ui.FriendlyError(d.replyErr)
		for _, line := range ui.WrappedLines(message, width) {
			lines = append(lines, d.theme.Warn.Render(line))
		}
		lines = append(lines, d.theme.Dim.Render("Press . to try again"))
	case len(d.replies) == 0:
		lines = append(lines, d.theme.Dim.Render("No replies yet"))
	}

	for index, reply := range d.replies {
		top := len(lines)
		card := postcomponent.RenderCardWithImages(
			reply,
			width,
			d.replyFocused && index == d.replyIndex,
			d.theme,
			time.Now(),
			d.images,
		)
		cardLines := strings.Split(card, "\n")
		lines = append(lines, cardLines...)
		posts = append(posts, postHit{
			post: reply, top: top, bottom: len(lines) - 1, right: width, card: true, index: index,
		})
	}

	switch {
	case d.loadingMoreReplies:
		lines = append(lines, d.theme.Dim.Render("  Loading more replies…"))
	case d.replyErr != nil && len(d.replies) > 0:
		lines = append(lines, d.theme.Warn.Render("  More replies unavailable: "+ui.FriendlyError(d.replyErr)))
		lines = append(lines, d.theme.Dim.Render("  Press . to try again"))
	case d.nextReplyCursor != nil && strings.TrimSpace(*d.nextReplyCursor) != "":
		lines = append(lines, d.theme.Dim.Render("  L  Load more replies"))
	}

	return lines, users, posts
}

func (d *Screen) selectedReply() *north.Post {
	if d.replyIndex < 0 || d.replyIndex >= len(d.replies) {
		return nil
	}

	return &d.replies[d.replyIndex]
}

func (d *Screen) moveReply(delta, width, room int) {
	if len(d.replies) == 0 {
		return
	}
	if !d.replyFocused {
		d.replyFocused = true
		d.replyIndex = 0
	} else {
		d.replyIndex = min(max(0, d.replyIndex+delta), len(d.replies)-1)
	}

	lines, _, hits := d.content(width)
	for _, hit := range hits {
		if !hit.card || hit.index != d.replyIndex {
			continue
		}
		if hit.top < d.offset {
			d.offset = hit.top
		} else if hit.top > d.offset+room/2 {
			d.offset = hit.top - room/2
		}
		if hit.bottom >= d.offset+room {
			d.offset = hit.bottom - room + 1
		}
		break
	}
	d.offset = clampOffset(d.offset, len(lines), room)
}

func (d *Screen) nearReplyEnd(width, room int) bool {
	if d.nextReplyCursor == nil || strings.TrimSpace(*d.nextReplyCursor) == "" || d.loadingConversation {
		return false
	}
	if d.replyFocused && d.replyIndex >= max(0, len(d.replies)-4) {
		return true
	}

	return d.offset+room >= len(d.contentLines(width))-4
}
