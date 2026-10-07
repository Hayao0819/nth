package user

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (d *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	innerWidth, room := d.layout(ctx.Width(), ctx.Height())
	if result, ok := msg.(modal.Result[domain.ProfileUpdate]); ok {
		if result.Ok() {
			return d.saveProfile(ctx.Context(), result.Value)
		}

		return nil
	}
	if confirmed, ok := msg.(modal.Result[dialog.Confirmation]); ok && d.confirmingBlock {
		d.confirmingBlock = false
		if confirmed.Ok() && confirmed.Value.Accepted {
			return d.changeRelationship(ctx.Context(), relationshipChange{kind: blockRelationship, enable: true})
		}

		return nil
	}
	if command, handled := d.updateResult(ctx.Context(), msg, innerWidth, room); handled {
		return command
	}
	if update, ok := msg.(postcomponent.ReactionUpdate); ok {
		d.applyReaction(update)

		return nil
	}

	var command tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return d.loadImages(ctx.Context(), ctx.Width(), d.user, d.posts)
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			d.offset += 3
			command = d.loadNearEnd(ctx.Context(), innerWidth, room)
		} else if msg.Button == tea.MouseWheelUp {
			d.offset -= 3
		}
	case tea.MouseClickMsg:
		return d.handleClick(ctx, msg, innerWidth, room)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "tab"):
		return d.cycleTab(ctx.Context())
	case reactea.Key(msg, "j", "down"):
		if len(d.posts) == 0 {
			d.offset++
		} else {
			d.moveSelection(1, innerWidth, room)
		}
		command = d.loadNearEnd(ctx.Context(), innerWidth, room)
	case reactea.Key(msg, "k", "up"):
		if len(d.posts) == 0 {
			d.offset--
		} else {
			d.moveSelection(-1, innerWidth, room)
		}
	case reactea.Key(msg, "ctrl+d", "pgdown", "space"):
		d.offset += max(1, room/2)
		command = d.loadNearEnd(ctx.Context(), innerWidth, room)
	case reactea.Key(msg, "ctrl+u", "pgup"):
		d.offset -= max(1, room/2)
	case reactea.Key(msg, "g", "home"):
		d.selected = 0
		d.offset = 0
	case reactea.Key(msg, "G", "end"):
		if len(d.posts) > 0 {
			d.selected = len(d.posts) - 1
			d.ensureSelectionVisible(innerWidth, room)
		} else {
			d.offset = len(d.content(innerWidth).lines)
		}
		command = d.loadNearEnd(ctx.Context(), innerWidth, room)
	case reactea.Key(msg, ".") && d.canRetry():
		return d.retry(ctx.Context())
	case reactea.Key(msg, "F"):
		return d.requestRelationship(ctx, followRelationship)
	case reactea.Key(msg, "B"):
		return d.requestRelationship(ctx, blockRelationship)
	case reactea.Key(msg, "M"):
		return d.requestRelationship(ctx, muteRelationship)
	case reactea.Key(msg, "N"):
		return d.requestRelationship(ctx, postNotificationRelationship)
	case reactea.Key(msg, "E"):
		return d.editProfile(ctx)
	case reactea.Key(msg, "S") && d.accountSafety && d.isOwnProfile():
		return navigation.OpenAccountSafety()
	case reactea.Key(msg, "A") && d.listMembers != nil && !d.isOwnProfile():
		return d.openListMemberships(ctx)
	case reactea.Key(msg, "o"):
		if d.connections != nil {
			return navigation.OpenFollowing(d.user)
		}
	case reactea.Key(msg, "O"):
		if d.connections != nil {
			return navigation.OpenFollowers(d.user)
		}
	case reactea.Key(msg, "enter"):
		if post := d.selectedPost(); post != nil {
			return navigation.OpenPost(*post)
		}
	case reactea.Key(msg, "u"):
		if post := d.selectedPost(); post != nil && post.DisplayPost() != nil {
			return navigation.OpenUser(post.DisplayPost().Author)
		}
	case reactea.Key(msg, "r", "R"):
		return d.requestSelected(postcomponent.Reply)
	case reactea.Key(msg, "t"):
		return d.requestSelected(postcomponent.Repost)
	case reactea.Key(msg, "l", "f"):
		return d.requestSelected(postcomponent.Like)
	case reactea.Key(msg, "Q"):
		return d.requestSelected(postcomponent.Quote)
	case reactea.Key(msg, "b"):
		return d.requestSelected(postcomponent.Bookmark)
	}

	d.clampOffset(innerWidth, room)

	return command
}

func (d *Screen) handleClick(ctx *reactea.Ctx, msg tea.MouseClickMsg, innerWidth, room int) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	if y < pageheader.Height || y >= pageheader.Height+room {
		return nil
	}
	content := d.content(innerWidth)
	contentRow := d.offset + y - pageheader.Height
	if following, ok := content.connectionAt(x, contentRow); ok {
		if following {
			return navigation.OpenFollowing(d.user)
		}

		return navigation.OpenFollowers(d.user)
	}
	if tab, ok := content.tabAt(x, contentRow); ok {
		return d.setTab(ctx.Context(), tab)
	}
	if relationship, ok := content.relationshipAt(x, contentRow); ok {
		return d.requestRelationship(ctx, relationship)
	}
	position, ok := content.postAt(contentRow)
	if !ok || x < position.left || x >= position.left+position.width {
		return nil
	}
	d.selected = position.index
	post := d.posts[position.index]
	cardX := x - position.left
	cardRow := d.offset + y - pageheader.Height - position.top
	if user, hit := postcomponent.CardUserAtWithImages(post, cardX, cardRow, position.width, d.images); hit {
		return navigation.OpenUser(user)
	}
	if quoted, hit := postcomponent.CardQuotedPostAtWithImages(post, cardX, cardRow, position.width, d.images); hit {
		return navigation.OpenPost(quoted)
	}
	if cardRow == position.height-2 {
		return postcomponent.Request(postcomponent.CardActionAt(cardX, position.width), post)
	}

	return navigation.OpenPost(post)
}
