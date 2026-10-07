package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	listdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type detailLoadedMsg struct {
	target   *Detail
	item     listdomain.ListItem
	response *north.Response
	err      error
}

func (m detailLoadedMsg) Response() *north.Response { return m.response }

type changedMsg struct {
	target   *Detail
	follow   bool
	enabled  bool
	response *north.Response
	err      error
}

func (m changedMsg) Response() *north.Response { return m.response }

func (s *Detail) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case modal.Result[dialog.FormResult]:
		if !msg.Ok() || msg.Value.Canceled {
			return nil
		}

		return s.updateList(ctx.Context(), msg.Value)
	case modal.Result[dialog.Confirmation]:
		if !s.deleting {
			return nil
		}
		s.deleting = false
		if !msg.Ok() || !msg.Value.Accepted {
			return nil
		}

		return s.deleteList(ctx.Context())
	case detailLoadedMsg:
		if msg.target != s {
			return nil
		}
		s.loading = false
		s.err = msg.err
		if msg.err == nil {
			s.item = msg.item
		}

		return nil
	case changedMsg:
		if msg.target != s {
			return nil
		}
		s.acting = false
		if msg.err != nil {
			s.notice = ui.FriendlyError(msg.err)
		} else if msg.follow {
			s.item.FollowedByViewer = msg.enabled
		} else {
			s.item.PinnedByViewer = msg.enabled
		}

		return nil
	case editedMsg:
		s.applyEdited(msg)

		return nil
	case deletedMsg:
		return s.applyDeleted(msg)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			x, y, inside := reactea.Mouse(ctx, msg)
			if !inside {
				return nil
			}
			if pageheader.BackAt(x, y) {
				return pageheader.Back()
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "enter"):
		if post := s.feed.SelectedPost(); post != nil {
			return navigation.OpenPost(*post)
		}
	case reactea.Key(msg, "u"):
		if post := s.feed.SelectedPost(); post != nil && post.DisplayPost() != nil {
			return navigation.OpenUser(post.DisplayPost().Author)
		}

		return navigation.OpenUser(s.item.Owner)
	case reactea.Key(msg, "o"):
		return navigation.OpenUser(s.item.Owner)
	case reactea.Key(msg, "f"):
		return s.toggle(ctx.Context(), true)
	case reactea.Key(msg, "p"):
		return s.toggle(ctx.Context(), false)
	case reactea.Key(msg, "m"):
		return navigation.OpenListMembers(s.item)
	case reactea.Key(msg, "e"):
		return s.openEdit(ctx)
	case reactea.Key(msg, "d"):
		return s.confirmDelete(ctx)
	case reactea.Key(msg, "r", "R"):
		return s.action(postcomponent.Reply)
	case reactea.Key(msg, "Q"):
		return s.action(postcomponent.Quote)
	}

	return s.updateFeed(ctx, msg)
}

func (s *Detail) load(ctx context.Context) tea.Cmd {
	if s.api == nil || s.loading || s.item.ID == "" {
		return nil
	}
	s.loading = true
	s.err = nil

	return func() tea.Msg {
		item, response, err := s.api.List(ctx, s.item.ID)

		return detailLoadedMsg{target: s, item: item, response: response, err: err}
	}
}

func (s *Detail) toggle(ctx context.Context, follow bool) tea.Cmd {
	if s.api == nil || s.acting || s.item.ID == "" {
		return nil
	}
	if follow && s.item.OwnedByViewer {
		s.notice = "You own this list"

		return nil
	}
	s.acting = true
	s.notice = ""
	enabled := !s.item.PinnedByViewer
	if follow {
		enabled = !s.item.FollowedByViewer
	}

	return func() tea.Msg {
		var response *north.Response
		var err error
		if follow {
			if enabled {
				_, response, err = s.api.FollowList(ctx, s.item.ID)
			} else {
				_, response, err = s.api.UnfollowList(ctx, s.item.ID)
			}
		} else if enabled {
			_, response, err = s.api.PinList(ctx, s.item.ID)
		} else {
			_, response, err = s.api.UnpinList(ctx, s.item.ID)
		}

		return changedMsg{target: s, follow: follow, enabled: enabled, response: response, err: err}
	}
}

func (s *Detail) action(action postcomponent.Action) tea.Cmd {
	post := s.feed.SelectedPost()
	if post == nil {
		return nil
	}

	return postcomponent.Request(action, *post)
}

func (s *Detail) updateFeed(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	feedCtx := s.feedCtx(ctx)
	if reactea.IsMouse(msg) {
		parentX, parentY := ctx.Origin()
		feedX, feedY := feedCtx.Origin()
		msg = reactea.TranslateMouse(msg, feedX-parentX, feedY-parentY)
	}

	return s.feed.Update(feedCtx, msg)
}
