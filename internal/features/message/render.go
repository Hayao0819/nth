package message

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	messagedomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (s *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	room := s.room(height)
	title := "Messages"
	if s.current != nil {
		title = conversationTitle(*s.current)
	} else if s.requests {
		title = "Message requests"
	}
	right := ""
	if s.loading {
		right = s.theme.Dim.Render("Loading…")
	} else if s.acting {
		right = s.theme.Dim.Render(ui.Clip(s.notice, max(1, width-len(title)-10)))
	} else if s.notice != "" {
		right = s.theme.Warn.Render(ui.Clip(s.notice, max(1, width-len(title)-10)))
	} else if s.current == nil && !s.requests && s.requestCount > 0 {
		right = s.theme.Dim.Render(fmt.Sprintf("%d requests", s.requestCount))
	}
	header := pageheader.Render(s.theme, title, right, width)
	body := s.renderItems(width, room)
	footer := "j/k move · Esc/← back"
	if s.current == nil {
		footer = "j/k move · Enter open"
		if s.request != nil {
			footer += " · Tab inbox/requests"
		}
		if s.write != nil {
			footer += " · N new"
		}
		footer += " · Esc/← back"
	} else if s.current.Request {
		footer = "j/k move · u profile"
		if s.request != nil {
			footer += " · y accept · x delete"
		}
		footer += " · Esc/← conversations"
	} else {
		footer = "j/k move · u profile"
		if s.write != nil {
			footer += " · s send · r reply · e edit · +/- react · d/D delete"
		}
		if s.current.Group && s.group != nil {
			footer += " · N rename · A add · L leave"
		}
		footer += " · Esc/← conversations"
	}
	if s.hasMore() {
		footer += " · M more"
	}

	return ui.Fit(header+"\n"+body+"\n"+s.theme.Dim.Render(ui.Clip(footer, width)), width, height)
}

func (s *Screen) renderItems(width, room int) string {
	if s.itemCount() == 0 {
		message := "No conversations"
		if s.current != nil {
			message = "No messages"
		}
		switch {
		case s.loading:
			message = "Loading…"
		case s.err != nil:
			message = ui.FriendlyError(s.err) + "\n\nPress . to try again"
		}

		return ui.Fit(lipgloss.Place(width, max(1, room), lipgloss.Center, lipgloss.Center, message), width, room)
	}

	lines := make([]string, 0, room)
	used := 0
	for index := s.top(); index < s.itemCount(); index++ {
		card := s.renderItem(index, width, index == s.index())
		height := lipgloss.Height(card)
		if used+height > room {
			break
		}
		lines = append(lines, card)
		used += height
	}
	if s.loadingMore && used < room {
		lines = append(lines, s.theme.Dim.Render("  Loading more…"))
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (s *Screen) renderItem(index, width int, selected bool) string {
	if s.current == nil {
		return s.renderConversation(s.conversations[index], width, selected)
	}

	return s.renderMessage(s.messages[index], width, selected)
}

func (s *Screen) renderConversation(conversation messagedomain.MessageConversation, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render("●")
	} else if conversation.UnreadCount > 0 {
		marker = s.theme.Active.Render("○")
	}
	title := s.theme.Name.Render(ui.Clip(conversationTitle(conversation), max(1, width-10)))
	right := timeLabel(conversation.UpdatedAt)
	if conversation.UnreadCount > 0 {
		right = fmt.Sprintf("%d unread", conversation.UnreadCount)
	}
	lines := []string{ui.Sides(marker+"  "+title, s.theme.Dim.Render(right), width)}
	if conversation.LastMessage != nil {
		preview := strings.TrimSpace(conversation.LastMessage.Text)
		if preview == "" && len(conversation.LastMessage.Media) > 0 {
			preview = "[media]"
		}
		lines = append(lines, "   "+s.theme.Dim.Render(ui.Clip(ui.SafeInline(preview), max(1, width-3))))
	}
	lines = append(lines, s.theme.Dim.Render(strings.Repeat("─", max(0, width-2))))

	return strings.Join(lines, "\n")
}

func (s *Screen) renderMessage(message messagedomain.Message, width int, selected bool) string {
	marker := " "
	if selected {
		marker = s.theme.Active.Render("●")
	}
	sender := messageSender(message)
	time := timeLabel(message.CreatedAt)
	if message.EditedAt != nil {
		time += " · edited"
	}
	first := ui.Sides(marker+"  "+s.theme.Name.Render(sender), s.theme.Dim.Render(time), width)
	text := strings.TrimSpace(message.Text)
	if text == "" && len(message.Media) > 0 {
		text = "[media]"
	}
	lines := []string{first}
	if message.ReplyTo != nil {
		reply := "Reply"
		if message.ReplyTo.Sender != nil {
			reply += " to @" + ui.SafeInline(message.ReplyTo.Sender.Handle)
		}
		preview := ui.SafeInline(message.ReplyTo.Text)
		if message.ReplyTo.Deleted {
			preview = "Message deleted"
		} else if preview == "" && message.ReplyTo.HasMedia {
			preview = "[media]"
		}
		lines = append(lines, "   "+s.theme.Dim.Render(ui.Clip(reply+": "+preview, max(1, width-3))))
	}
	wrapped := ui.WrappedLines(text, max(1, width-4))
	for _, line := range wrapped[:min(3, len(wrapped))] {
		lines = append(lines, "   "+line)
	}
	if message.Post != nil {
		preview := ui.SafeInline(message.Post.Text)
		if preview == "" {
			preview = "Shared post"
		}
		lines = append(lines, "   "+s.theme.Dim.Render(ui.Clip("Post · "+preview, max(1, width-3))))
	}
	if len(message.Media) > 0 && strings.TrimSpace(message.Text) != "" {
		lines = append(lines, "   "+s.theme.Dim.Render(fmt.Sprintf("%d media", len(message.Media))))
	}
	if len(message.Reactions) > 0 {
		labels := make([]string, 0, len(message.Reactions))
		for _, reaction := range message.Reactions {
			label := reaction.Emoji
			if reaction.Count > 1 {
				label += fmt.Sprintf(" %d", reaction.Count)
			}
			if reaction.ReactedByViewer {
				label = "[" + label + "]"
			}
			labels = append(labels, label)
		}
		lines = append(lines, "   "+s.theme.Dim.Render(ui.Clip(strings.Join(labels, "  "), max(1, width-3))))
	}
	lines = append(lines, s.theme.Dim.Render(strings.Repeat("─", max(0, width-2))))

	return strings.Join(lines, "\n")
}

func messageSender(message messagedomain.Message) string {
	sender := ui.SafeInline(message.Sender.Name)
	if sender == "" {
		sender = "@" + ui.SafeInline(message.Sender.Handle)
	}

	return sender
}

func conversationTitle(conversation messagedomain.MessageConversation) string {
	if conversation.Name != nil && strings.TrimSpace(*conversation.Name) != "" {
		return ui.SafeInline(*conversation.Name)
	}
	names := make([]string, 0, len(conversation.Participants))
	for _, participant := range conversation.Participants {
		name := ui.SafeInline(participant.Name)
		if name == "" {
			name = "@" + ui.SafeInline(participant.Handle)
		}
		if name != "@" && name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "Conversation"
	}

	return strings.Join(names, ", ")
}

func timeLabel(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	age := time.Since(value)
	switch {
	case age < time.Minute:
		return "now"
	case age < time.Hour:
		return fmt.Sprintf("%dm", int(age/time.Minute))
	case age < 24*time.Hour:
		return fmt.Sprintf("%dh", int(age/time.Hour))
	default:
		return value.Format("Jan 2")
	}
}
