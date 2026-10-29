package message

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
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
			if s.current != nil && s.scrollMessage(3, ctx.Width(), s.room(ctx.Height())) {
				return nil
			}
			s.move(1, ctx.Width(), s.room(ctx.Height()))
			if s.current == nil {
				return s.loadNearEdge(ctx.Context())
			}

			return nil
		}
		if msg.Button == tea.MouseWheelUp {
			if s.current != nil && s.scrollMessage(-3, ctx.Width(), s.room(ctx.Height())) {
				return nil
			}
			s.move(-1, ctx.Width(), s.room(ctx.Height()))
			if s.current != nil {
				return s.loadNearEdge(ctx.Context())
			}
		}

		return nil
	case tea.WindowSizeMsg:
		s.ensureVisible(ctx.Width(), s.room(ctx.Height()))

		return nil
	case tea.MouseClickMsg:
		return s.handleClick(ctx, msg)
	}
	if s.composing && reactea.IsKeyboard(msg) {
		return s.updateComposer(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return s.back()
	case reactea.Key(msg, "j", "down"):
		s.move(1, ctx.Width(), s.room(ctx.Height()))
		if s.current == nil {
			return s.loadNearEdge(ctx.Context())
		}
	case reactea.Key(msg, "k", "up"):
		s.move(-1, ctx.Width(), s.room(ctx.Height()))
		if s.current != nil {
			return s.loadNearEdge(ctx.Context())
		}
	case reactea.Key(msg, "g", "home"):
		s.setIndex(0)
		s.setTop(0)
		if s.current != nil {
			return s.loadNearEdge(ctx.Context())
		}
	case reactea.Key(msg, "G", "end"):
		s.setIndex(max(0, s.itemCount()-1))
		s.ensureVisible(ctx.Width(), s.room(ctx.Height()))
		if s.current == nil {
			return s.loadNearEdge(ctx.Context())
		}
	case reactea.Key(msg, "ctrl+d", "pgdown", "space") && s.current != nil:
		s.scrollMessage(max(1, s.room(ctx.Height())*2/3), ctx.Width(), s.room(ctx.Height()))
	case reactea.Key(msg, "ctrl+u", "pgup") && s.current != nil:
		s.scrollMessage(-max(1, s.room(ctx.Height())*2/3), ctx.Width(), s.room(ctx.Height()))
	case reactea.Key(msg, "enter") && s.current == nil:
		return s.openConversation(ctx.Context())
	case reactea.Key(msg, "enter", "tab") && s.composerVisible():
		return s.openComposer(ctx)
	case reactea.Key(msg, "tab") && s.current == nil:
		return s.toggleRequests(ctx.Context())
	case reactea.Key(msg, "m"):
		return s.openRecipients(ctx, recipientConversation)
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
		if s.composing {
			return s.closeComposer()
		}

		return s.back()
	}
	if s.composerVisible() && y >= ctx.Height()-composerHeight {
		if !s.composing {
			return s.openComposer(ctx)
		}
		if y == ctx.Height()-1 {
			button := s.composerButton()
			if x >= max(0, ctx.Width()-lipgloss.Width(button)) {
				return s.submitComposer(ctx)
			}
			if x < lipgloss.Width("Esc  Done") {
				return s.closeComposer()
			}
		}

		return s.composer.Widget.Focus()
	}
	if index, row, ok := s.itemPositionAt(y-pageheader.Height, ctx.Width(), s.room(ctx.Height())); ok {
		s.setIndex(index)
		if s.current == nil {
			return s.openConversation(ctx.Context())
		}
		user := s.messages[index].Sender
		label := messageSender(s.messages[index])
		lines := strings.Split(s.renderMessage(s.messages[index], ctx.Width(), true), "\n")
		if user.Handle != "" && row < len(lines) && ui.TextAt(lines[row], x, label) {
			return navigation.OpenUser(user)
		}
	}

	return nil
}
