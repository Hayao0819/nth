package message

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	messagedomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
)

func (s *Screen) renderMessage(message messagedomain.Message, width int, selected bool) string {
	outerLimit := max(1, width)
	if width > 4 {
		outerLimit = width - 2
	}
	outerLimit = min(outerLimit, max(16, width*3/4))
	contentLimit := max(1, outerLimit-4)
	body := s.messageBody(message, contentLimit)
	sender := messageSender(message)
	metadata := make([]string, 0, 3)
	if timestamp := timeLabel(message.CreatedAt); timestamp != "" {
		metadata = append(metadata, timestamp)
	}
	if message.EditedAt != nil {
		metadata = append(metadata, "edited")
	}
	if s.ownsMessage(message) && message.Read {
		metadata = append(metadata, "✓ Seen")
	}
	timestamp := strings.Join(metadata, " · ")
	marker := "  "
	if selected {
		marker = "● "
	}
	naturalWidth := lipgloss.Width(marker + sender)
	if timestamp != "" {
		naturalWidth += lipgloss.Width(timestamp) + 1
	}
	for _, line := range body {
		naturalWidth = max(naturalWidth, lipgloss.Width(line))
	}
	contentWidth := min(contentLimit, max(1, naturalWidth))
	header := marker + sender
	if timestamp != "" {
		header = ui.Sides(header, timestamp, contentWidth)
	}
	content := strings.Join(append([]string{header}, body...), "\n")
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(contentWidth + 4)
	if s.ownsMessage(message) {
		style = style.
			Foreground(lipgloss.Color("0")).
			Background(s.theme.Active.GetForeground()).
			BorderForeground(s.theme.Active.GetForeground())
	} else {
		style = style.
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("235")).
			BorderForeground(lipgloss.Color("8"))
		if selected {
			style = style.BorderForeground(s.theme.Active.GetForeground())
		}
	}
	bubble := style.Render(content)
	if s.ownsMessage(message) {
		bubble = lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(bubble)
	}

	return bubble + "\n "
}

func (s *Screen) messageBody(message messagedomain.Message, width int) []string {
	lines := make([]string, 0, 4)
	if message.ReplyTo != nil {
		reply := "↪ Reply"
		if message.ReplyTo.Sender != nil {
			reply += " to @" + ui.SafeInline(message.ReplyTo.Sender.Handle)
		}
		preview := ui.SafeInline(message.ReplyTo.Text)
		if message.ReplyTo.Deleted {
			preview = "Message deleted"
		} else if preview == "" && message.ReplyTo.HasMedia {
			preview = "[media]"
		}
		if preview != "" {
			reply += ": " + preview
		}
		lines = append(lines, ui.Clip(reply, width))
	}
	text := strings.TrimSpace(message.Text)
	if text == "" && len(message.Media) > 0 {
		text = "[media]"
	}
	written := ui.WrappedLines(text, width)
	lines = append(lines, written...)
	if message.Post != nil {
		preview := ui.SafeInline(message.Post.Text)
		if preview == "" {
			preview = "Shared post"
		}
		lines = append(lines, ui.Clip("Post · "+preview, width))
	}
	if len(message.Media) > 0 && strings.TrimSpace(message.Text) != "" {
		lines = append(lines, ui.Clip(fmt.Sprintf("%d media", len(message.Media)), width))
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
		lines = append(lines, ui.Clip(strings.Join(labels, "  "), width))
	}
	if len(lines) == 0 {
		lines = append(lines, "Message")
	}

	return lines
}
