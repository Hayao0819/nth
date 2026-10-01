package feed

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/navigation"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/reactea/v2"
)

func (f *Feed) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case feedLoadedMsg:
		return f.applyLoadedPage(ctx, msg)

	case reactionMsg:
		f.applyReaction(msg)

		return nil

	case tea.WindowSizeMsg:
		f.fillLoads = 0
		f.ensureVisible(ctx.Width(), ctx.Height())
		imageCommand := f.loadImages(ctx, f.posts)
		if f.nextCursor != nil && f.renderedHeight(ctx.Width()) < ctx.Height() {
			return tea.Batch(imageCommand, f.load(ctx, true))
		}

		return imageCommand

	case tea.MouseWheelMsg:
		f.notice = ""
		switch msg.Button {
		case tea.MouseWheelDown:
			f.refreshReady = true
			f.moveByHeight(1, 3, ctx.Width(), ctx.Height())

			return f.loadNearEnd(ctx)
		case tea.MouseWheelUp:
			if f.atTop() {
				return f.refreshAtTop(ctx)
			}
			f.moveByHeight(-1, 3, ctx.Width(), ctx.Height())
		}

		return nil

	case tea.MouseClickMsg:
		return f.handleClick(ctx, msg)
	}

	if reactea.IsKeyboard(msg) {
		f.notice = ""
	}

	switch {
	case reactea.Key(msg, "j", "down"):
		f.refreshReady = true
		f.move(1, ctx.Width(), ctx.Height())

		return f.loadNearEnd(ctx)
	case reactea.Key(msg, "k", "up"):
		if f.atTop() {
			return f.refreshAtTop(ctx)
		}
		f.move(-1, ctx.Width(), ctx.Height())
	case reactea.Key(msg, "g", "home"):
		f.selected = 0
		f.ensureVisible(ctx.Width(), ctx.Height())
	case reactea.Key(msg, "G", "end"):
		f.refreshReady = true
		f.selected = max(0, len(f.posts)-1)
		f.ensureVisible(ctx.Width(), ctx.Height())

		return f.loadNearEnd(ctx)
	case reactea.Key(msg, "ctrl+d", "pgdown", "space"):
		f.refreshReady = true
		f.moveByHeight(1, max(1, ctx.Height()*2/3), ctx.Width(), ctx.Height())

		return f.loadNearEnd(ctx)
	case reactea.Key(msg, "ctrl+u", "pgup"):
		if f.atTop() {
			return f.refreshAtTop(ctx)
		}
		f.moveByHeight(-1, max(1, ctx.Height()*2/3), ctx.Width(), ctx.Height())
	case reactea.Key(msg, "."):
		return f.load(ctx, false)
	case reactea.Key(msg, "L"):
		if f.nextCursor != nil {
			return f.load(ctx, true)
		}
	case reactea.Key(msg, "f", "l"):
		return f.ToggleLike(ctx.Context())
	case reactea.Key(msg, "t"):
		return f.ToggleRepost(ctx.Context())
	}

	return nil
}

func (f *Feed) handleClick(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	f.notice = ""
	if msg.Button != tea.MouseLeft {
		return nil
	}
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	index, row, ok := f.postAt(y, ctx.Width(), ctx.Height())
	if !ok {
		return nil
	}
	cardHeight := lipgloss.Height(postcomponent.RenderCardWithImages(f.posts[index], ctx.Width(), true, f.theme, time.Now(), f.images))
	if user, hit := postcomponent.CardUserAtWithImages(f.posts[index], x, row, ctx.Width(), f.images); hit {
		f.selectPost(index, ctx.Width(), ctx.Height())

		return navigation.OpenUser(user)
	}
	if quoted, hit := postcomponent.CardQuotedPostAtWithImages(f.posts[index], x, row, ctx.Width(), f.images); hit {
		f.selectPost(index, ctx.Width(), ctx.Height())

		return navigation.OpenPost(quoted)
	}
	if row == cardHeight-2 {
		action := postcomponent.CardActionAt(x, ctx.Width())
		if action == postcomponent.Reply || action == postcomponent.Quote {
			f.selectPost(index, ctx.Width(), ctx.Height())
		}
		post := f.posts[index]

		return postcomponent.Request(action, post)
	}
	f.selectPost(index, ctx.Width(), ctx.Height())
	post := f.posts[index]

	return navigation.OpenPost(post)
}
