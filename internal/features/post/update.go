package post

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (d *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	innerWidth, room := d.layout(ctx.Width(), ctx.Height())
	scrollStep := max(1, room/2)
	if opened, ok := msg.(navigation.BrowserOpenedMsg); ok && opened.URL == d.browserURL() {
		if opened.Err != nil {
			d.notice = "Could not open browser: " + ui.FriendlyError(opened.Err)
		} else {
			d.notice = "Opened in browser"
		}

		return nil
	}
	if choice, ok := msg.(modal.Result[pollChoice]); ok {
		if choice.Ok() {
			return d.vote(ctx.Context(), choice.Value.OptionID)
		}

		return nil
	}
	if confirmed, ok := msg.(modal.Result[dialog.Confirmation]); ok {
		if !d.confirmingDelete {
			return nil
		}
		d.confirmingDelete = false
		if confirmed.Ok() && confirmed.Value.Accepted {
			return postcomponent.Request(postcomponent.Delete, d.post)
		}

		return nil
	}
	if loaded, ok := msg.(ancestorsLoadedMsg); ok {
		if loaded.target != d {
			return nil
		}
		d.loadingAncestors = false
		d.ancestors = loaded.posts
		d.ancestorErr = loaded.err
		d.ancestorsTruncated = loaded.truncated
		d.offset = max(0, len(d.threadLines(innerWidth))-room/2)
		d.offset = clampOffset(d.offset, len(d.contentLines(innerWidth)), room)

		return d.loadImages(ctx.Context(), ctx.Width(), loaded.posts)
	}
	if loaded, ok := msg.(conversationLoadedMsg); ok {
		if loaded.target != d {
			return nil
		}
		d.loadingConversation = false
		d.loadingMoreReplies = false
		d.replyErr = loaded.err
		if loaded.err != nil {
			if len(d.ancestors) == 0 && d.api != nil && replyParentID(d.post) != "" {
				d.loadingAncestors = true

				return d.loadAncestors(ctx.Context())
			}

			return nil
		}
		if loaded.more {
			d.replies = appendUniqueReplies(d.replies, loaded.page.Replies)
		} else {
			d.ancestors = append([]north.Post(nil), loaded.page.Ancestors...)
			d.replies = append([]north.Post(nil), loaded.page.Replies...)
			d.offset = max(0, len(d.threadLines(innerWidth))-room/2)
		}
		d.nextReplyCursor = loaded.page.NextCursor
		d.offset = clampOffset(d.offset, len(d.contentLines(innerWidth)), room)
		posts := append(append([]north.Post(nil), loaded.page.Ancestors...), loaded.page.Replies...)
		images := d.loadImages(ctx.Context(), ctx.Width(), posts)
		if d.nextReplyCursor != nil && len(d.contentLines(innerWidth)) <= room {
			return tea.Batch(images, d.loadConversation(ctx.Context(), true))
		}

		return images
	}
	if loaded, ok := msg.(editableLoadedMsg); ok {
		if loaded.target != d {
			return nil
		}
		d.checkingEditable = false
		d.editableChecked = true
		d.editable = loaded.err == nil && loaded.eligible
		if loaded.err == nil {
			d.post = loaded.post
			d.editETag = loaded.etag
		} else {
			d.editETag = ""
			d.notice = "Editing unavailable: " + ui.FriendlyError(loaded.err)
		}

		return nil
	}
	if voted, ok := msg.(pollVotedMsg); ok {
		d.applyPollVote(voted)

		return nil
	}
	if update, ok := msg.(ManagementUpdate); ok {
		wasManage := d.manage
		d.manage = ownedBy(d.post, update.Account)
		if d.manage && !wasManage && d.editor != nil && !d.editableChecked && !d.checkingEditable {
			d.checkingEditable = true
			return d.loadEditable(ctx.Context())
		}
	}
	if update, ok := msg.(postcomponent.ReactionUpdate); ok {
		if !d.hasPost(update.PostID) {
			return nil
		}
		delete(d.acting, update.Action)
		if update.Err != nil {
			d.notice = ui.FriendlyError(update.Err)

			return nil
		}
		d.applyReaction(update)
		switch {
		case update.Like != nil:
			if update.Like.Liked {
				d.notice = "Liked"
			} else {
				d.notice = "Like removed"
			}
		case update.Repost != nil:
			if update.Repost.Reposted {
				d.notice = "Reposted"
			} else {
				d.notice = "Repost removed"
			}
		case update.Bookmark != nil:
			if *update.Bookmark {
				d.notice = "Bookmarked"
			} else {
				d.notice = "Bookmark removed"
			}
		}

		return nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		posts := append(append([]north.Post(nil), d.ancestors...), d.post)
		posts = append(posts, d.replies...)

		return d.loadImages(ctx.Context(), ctx.Width(), posts)
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			d.offset += 3
		} else if msg.Button == tea.MouseWheelUp {
			d.offset -= 3
		}
	case tea.MouseClickMsg:
		x, y, inside := reactea.Mouse(ctx, msg)
		if !inside || msg.Button != tea.MouseLeft {
			return nil
		}
		if pageheader.ActionAt(d.theme, x, y, innerWidth, d.browserAction()) {
			return d.openBrowser()
		}
		if pageheader.BackAt(x, y) {
			return pageheader.Back()
		}
		contentRow := clampOffset(d.offset, len(d.contentLines(innerWidth)), room) + y - pageheader.Height
		if optionID, ok := d.pollOptionAt(x, contentRow, innerWidth); ok {
			return d.vote(ctx.Context(), optionID)
		}
		if activity, ok := d.activityAtPosition(x, contentRow, innerWidth); ok {
			return d.requestActivity(activity)
		}
		if user, ok := d.userAtPosition(x, y, innerWidth, room); ok {
			return d.requestUser(ctx, user)
		}
		if action, post, ok := d.replyActionAtPosition(x, y, innerWidth, room); ok {
			return postcomponent.Request(action, post)
		}
		if post, ok := d.postAtPosition(x, y, innerWidth, room); ok {
			return postcomponent.Request(postcomponent.ViewPost, post)
		}
		if contentRow == d.actionRow(innerWidth) && postcomponent.CanInteract(&d.post) {
			actions := d.actions()
			column := ui.ColumnAt(x, innerWidth, len(actions))

			return d.request(ctx, actions[column])
		}
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "w"):
		return d.openBrowser()
	case reactea.Key(msg, "enter") && d.replyFocused:
		if reply := d.selectedReply(); reply != nil {
			selected := *reply

			return postcomponent.Request(postcomponent.ViewPost, selected)
		}
	case reactea.Key(msg, "J"):
		d.moveReply(1, innerWidth, room)
	case reactea.Key(msg, "K"):
		d.moveReply(-1, innerWidth, room)
	case reactea.Key(msg, "u"):
		if reply := d.selectedReply(); d.replyFocused && reply != nil && reply.DisplayPost() != nil {
			return d.requestUser(ctx, reply.DisplayPost().Author)
		}
		return d.request(ctx, postcomponent.ViewAuthor)
	case reactea.Key(msg, "U"):
		return d.requestNextUser(ctx)
	case reactea.Key(msg, "p") && len(d.ancestors) > 0:
		parent := d.ancestors[len(d.ancestors)-1]
		return postcomponent.Request(postcomponent.ViewPost, parent)
	case reactea.Key(msg, "r", "R"):
		return d.requestSelected(ctx, postcomponent.Reply)
	case reactea.Key(msg, "t"):
		return d.requestSelected(ctx, postcomponent.Repost)
	case reactea.Key(msg, "l", "f"):
		return d.requestSelected(ctx, postcomponent.Like)
	case reactea.Key(msg, "Q"):
		return d.requestSelected(ctx, postcomponent.Quote)
	case reactea.Key(msg, "b"):
		return d.requestSelected(ctx, postcomponent.Bookmark)
	case reactea.Key(msg, "v"):
		return d.openPoll(ctx)
	case reactea.Key(msg, "V"):
		return d.requestActivity(navigation.PostQuotes)
	case reactea.Key(msg, "I"):
		return d.requestActivity(navigation.PostLikes)
	case reactea.Key(msg, "T"):
		return d.requestActivity(navigation.PostReposts)
	case reactea.Key(msg, "H"):
		return d.requestActivity(navigation.PostHistory)
	case reactea.Key(msg, "e") && d.manage && d.editable:
		return d.request(ctx, postcomponent.Edit)
	case reactea.Key(msg, "d") && d.manage:
		return d.request(ctx, postcomponent.Delete)
	case reactea.Key(msg, "j", "down"):
		d.offset++
	case reactea.Key(msg, "k", "up"):
		d.offset--
	case reactea.Key(msg, "ctrl+d", "pgdown", "space"):
		d.offset += scrollStep
	case reactea.Key(msg, "ctrl+u", "pgup"):
		d.offset -= scrollStep
	case reactea.Key(msg, "g", "home"):
		d.offset = 0
	case reactea.Key(msg, "G", "end"):
		d.offset = len(d.contentLines(innerWidth))
	case reactea.Key(msg, "L"):
		return d.loadConversation(ctx.Context(), true)
	case reactea.Key(msg, ".") && d.replyErr != nil:
		return d.loadConversation(ctx.Context(), len(d.replies) > 0)
	}

	d.offset = clampOffset(d.offset, len(d.contentLines(innerWidth)), room)
	if d.nearReplyEnd(innerWidth, room) {
		return d.loadConversation(ctx.Context(), true)
	}

	return nil
}
