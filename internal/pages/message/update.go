package message

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case conversationsLoadedMsg:
		s.applyConversationsLoaded(ctx, msg)

		return nil
	case messagesLoadedMsg:
		s.applyMessagesLoaded(ctx, msg)

		return nil
	case readMsg:
		s.applyRead(msg)

		return nil
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			s.move(1, ctx.Width(), s.room(ctx.Height()))

			return s.loadNearEnd(ctx.Context())
		}
		if msg.Button == tea.MouseWheelUp {
			s.move(-1, ctx.Width(), s.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		return s.handleClick(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return s.back()
	case reactea.Key(msg, "j", "down"):
		s.move(1, ctx.Width(), s.room(ctx.Height()))

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "k", "up"):
		s.move(-1, ctx.Width(), s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.setIndex(0)
		s.setTop(0)
	case reactea.Key(msg, "G", "end"):
		s.setIndex(max(0, s.itemCount()-1))
		s.ensureVisible(ctx.Width(), s.room(ctx.Height()))

		return s.loadNearEnd(ctx.Context())
	case reactea.Key(msg, "enter") && s.current == nil:
		return s.openConversation(ctx.Context())
	case reactea.Key(msg, "u") && s.current != nil:
		if user, ok := s.selectedSender(); ok {
			return navigation.OpenUser(user)
		}
	case reactea.Key(msg, "L"):
		if s.current == nil {
			return s.loadConversations(ctx.Context(), true)
		}

		return s.loadMessages(ctx.Context(), true)
	case reactea.Key(msg, ".") && s.err != nil:
		if s.current == nil {
			return s.loadConversations(ctx.Context(), false)
		}

		return s.loadMessages(ctx.Context(), false)
	}

	return nil
}

func (s *Screen) handleClick(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return s.back()
	}
	if index, ok := s.itemAt(y-pageheader.Height, ctx.Width(), s.room(ctx.Height())); ok {
		s.setIndex(index)
		if s.current == nil {
			return s.openConversation(ctx.Context())
		}
	}

	return nil
}
