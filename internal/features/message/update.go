package message

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (s *Screen) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case modal.Result[recipientResult]:
		purpose := s.recipients
		s.recipients = recipientNone
		if !msg.Ok() || msg.Value.Canceled {
			return nil
		}
		if purpose == recipientMembers {
			return s.addMembers(ctx.Context(), msg.Value.Handles)
		}

		return s.createConversation(ctx.Context(), msg.Value.Handles)
	case conversationsLoadedMsg:
		s.applyConversationsLoaded(ctx, msg)

		return nil
	case messagesLoadedMsg:
		s.applyMessagesLoaded(ctx, msg)

		return nil
	case readMsg:
		s.applyRead(msg)
		if msg.err == nil {
			return navigation.MessagesRead()
		}

		return nil
	case actionMsg:
		return s.applyAction(ctx, msg)
	case modal.Result[dialog.TextResult]:
		if s.prompt == promptNone {
			return nil
		}
		if !msg.Ok() || msg.Value.Canceled {
			s.prompt = promptNone
			s.promptID = ""

			return nil
		}

		return s.applyPrompt(ctx.Context(), msg.Value.Text)
	case modal.Result[dialog.Confirmation]:
		if s.confirm == confirmNone {
			return nil
		}

		return s.applyConfirmation(ctx.Context(), msg.Ok() && msg.Value.Accepted)
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
	case reactea.Key(msg, "tab") && s.current == nil:
		return s.toggleRequests(ctx.Context())
	case reactea.Key(msg, "N") && s.current == nil:
		return s.openRecipients(ctx, recipientConversation)
	case reactea.Key(msg, "s") && s.current != nil && !s.current.Request:
		return s.promptFor(ctx, promptMessage)
	case reactea.Key(msg, "r") && s.current != nil && !s.current.Request:
		return s.promptFor(ctx, promptReply)
	case reactea.Key(msg, "e") && s.current != nil && !s.current.Request:
		return s.promptFor(ctx, promptEdit)
	case reactea.Key(msg, "+") && s.current != nil && !s.current.Request:
		return s.promptFor(ctx, promptReaction)
	case reactea.Key(msg, "-") && s.current != nil && !s.current.Request:
		return s.removeReaction(ctx.Context())
	case reactea.Key(msg, "d") && s.current != nil && !s.current.Request:
		return s.confirmAction(ctx, confirmDeleteMessage)
	case reactea.Key(msg, "D") && s.current != nil && !s.current.Request:
		return s.confirmAction(ctx, confirmDeleteMessageForEveryone)
	case reactea.Key(msg, "y") && s.current != nil && s.current.Request:
		return s.acceptRequest(ctx.Context())
	case reactea.Key(msg, "x") && s.current != nil && s.current.Request:
		return s.confirmAction(ctx, confirmDeleteRequest)
	case reactea.Key(msg, "N") && s.current != nil && s.current.Group:
		return s.promptFor(ctx, promptRename)
	case reactea.Key(msg, "A") && s.current != nil && s.current.Group:
		return s.openRecipients(ctx, recipientMembers)
	case reactea.Key(msg, "L") && s.current != nil && s.current.Group:
		return s.confirmAction(ctx, confirmLeaveConversation)
	case reactea.Key(msg, "u") && s.current != nil:
		if user, ok := s.selectedSender(); ok {
			return navigation.OpenUser(user)
		}
	case reactea.Key(msg, "M"):
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
	if index, row, ok := s.itemPositionAt(y-pageheader.Height, ctx.Width(), s.room(ctx.Height())); ok {
		s.setIndex(index)
		if s.current == nil {
			return s.openConversation(ctx.Context())
		}
		if row == 0 {
			user := s.messages[index].Sender
			label := messageSender(s.messages[index])
			if user.Handle != "" && x >= 3 && x < 3+lipgloss.Width(label) {
				return navigation.OpenUser(user)
			}
		}
	}

	return nil
}
