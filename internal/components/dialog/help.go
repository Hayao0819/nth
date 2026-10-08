package dialog

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type Text struct {
	reactea.BasicComponent

	theme   ui.Theme
	title   string
	content func(int) []string
	offset  int
}

func (d *Text) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	innerHeight := max(0, ctx.Height()-d.theme.Dialog.GetVerticalFrameSize())
	room := max(1, innerHeight-4)
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		if _, _, inside := reactea.Mouse(ctx, msg); inside {
			if wheel.Button == tea.MouseWheelDown {
				d.offset += 3
			} else if wheel.Button == tea.MouseWheelUp {
				d.offset -= 3
			}
		}
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && y == 1 && x >= ctx.Width()-4 {
			return modal.Dismiss(ctx)
		}
	}
	switch {
	case reactea.Key(msg, "esc", "q", "enter"):
		return modal.Dismiss(ctx)
	case reactea.Key(msg, "j", "down"):
		d.offset++
	case reactea.Key(msg, "k", "up"):
		d.offset--
	case reactea.Key(msg, "ctrl+d", "pgdown"):
		d.offset += max(1, room/2)
	case reactea.Key(msg, "ctrl+u", "pgup"):
		d.offset -= max(1, room/2)
	case reactea.Key(msg, "g", "home"):
		d.offset = 0
	case reactea.Key(msg, "G", "end"):
		d.offset = len(d.content(max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())))
	}
	innerWidth := max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())
	maximum := max(0, len(d.content(innerWidth))-room)
	d.offset = min(max(0, d.offset), maximum)

	return nil
}

func (d *Text) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := d.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	room := max(0, innerHeight-4)
	lines := d.content(innerWidth)
	offset := min(d.offset, max(0, len(lines)-room))
	end := min(len(lines), offset+room)

	visible := ""
	if offset < end {
		visible = strings.Join(lines[offset:end], "\n")
	}
	footer := "↑/↓ scroll · Esc close"
	if len(lines) > room {
		position := fmt.Sprintf("%d–%d of %d", min(len(lines), offset+1), end, len(lines))
		footer = ui.Sides(footer, position, innerWidth)
	}
	header := ui.Sides(d.theme.ModalTitle.Render(ui.Clip(d.title, max(1, innerWidth-2))), d.theme.Key.Render("×"), innerWidth)
	content := header + "\n\n" +
		ui.Fit(visible, innerWidth, room) + "\n\n" + d.theme.Dim.Render(ui.Clip(footer, innerWidth))

	return RenderDialog(style, content, width, height)
}

func NewHelp(theme ui.Theme, notifications, messages, bookmarks bool) *Text {
	return &Text{theme: theme, title: "Keyboard help", content: func(width int) []string {
		return helpLines(theme, width, notifications, messages, bookmarks)
	}}
}

type helpItem struct {
	key         string
	description string
}

func helpLines(theme ui.Theme, width int, notifications, messages, bookmarks bool) []string {
	timelineItems := []helpItem{
		{"1", "Open Home"},
		{"Tab / Shift+Tab", "Switch For you / Following"},
	}
	if notifications {
		timelineItems = append(timelineItems, helpItem{"2 / v", "Open notifications"})
	}
	if messages {
		timelineItems = append(timelineItems, helpItem{"3", "Open messages"})
	}
	if bookmarks {
		timelineItems = append(timelineItems, helpItem{"4", "Open bookmarks"})
	}
	timelineItems = append(timelineItems, helpItem{"6–0", "Search a trend from the right column"})
	groups := []struct {
		title string
		items []helpItem
	}{
		{"TIMELINES", timelineItems},
		{"NAVIGATION", []helpItem{
			{"j / k · ↑ / ↓", "Move the selection"},
			{"Space · Ctrl+D/U", "Page down / up"},
			{"g / G", "First / last post"},
			{"Mouse", "Select, act, or scroll"},
		}},
		{"POSTS", []helpItem{
			{"Enter", "Open the selected post"},
			{"w", "Open post or profile details in a browser"},
			{"p", "Open the parent from a reply's details"},
			{"u", "Open the selected post's author"},
			{"U", "Open the next linked account (reply, repost, or quote)"},
			{"n / c", "New post"},
			{"r / R", "Reply to the selected post"},
			{"Q", "Quote the selected post"},
			{"l / f", "Like or unlike the selected post"},
			{"t", "Repost or undo the selected post"},
			{"e", "Edit an eligible post you own"},
			{"d", "Delete your own post from its details"},
			{"Ctrl+S", "Send; Ctrl+Enter where supported"},
			{"Enter / y", "Confirm deletion"},
			{"Esc / n", "Close or cancel; the composer keeps its draft"},
		}},
		{"OTHER", []helpItem{
			{"a", "Open your profile"},
			{"/", "Open search; results update after typing pauses"},
			{"↑/↓ · Ctrl+P/N", "Choose a search result"},
			{"↑ at top", "Fetch the latest posts"},
			{"L", "Load older posts"},
			{"? · q", "Help / quit"},
			{"last request", "Rate limit for the endpoint group used by the latest request"},
			{"nth setup", "Configure authentication credentials"},
			{"Credentials", "Saved in the current user's system keyring"},
		}},
	}

	var lines []string
	for index, group := range groups {
		if index > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, theme.Section.Render(group.title))
		for _, item := range group.items {
			lines = append(lines, renderHelpItem(theme, item, width)...)
		}
	}

	return lines
}

func renderHelpItem(theme ui.Theme, item helpItem, width int) []string {
	if width < 30 {
		lines := []string{theme.Key.Render(item.key)}
		for _, line := range ui.WrappedLines(item.description, max(1, width-2)) {
			lines = append(lines, "  "+line)
		}

		return lines
	}

	keyWidth := min(20, max(12, width/3))
	description := ui.WrappedLines(item.description, max(1, width-keyWidth-2))
	lines := make([]string, 0, len(description))
	for index, line := range description {
		key := ""
		if index == 0 {
			key = theme.Key.Render(item.key)
		}
		lines = append(lines, ui.Left(key, keyWidth)+"  "+line)
	}

	return lines
}

func Placement(ctx *reactea.Ctx, preferredWidth, preferredHeight int) modal.Placement {
	width, height := ctx.Size()

	return modal.Centered(min(max(1, width), preferredWidth), min(max(1, height), preferredHeight))
}
