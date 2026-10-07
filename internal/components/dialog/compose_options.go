package dialog

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (d *Compose) updatePostOptions(ctx *reactea.Ctx, msg tea.Msg) (tea.Cmd, bool) {
	result, ok := msg.(modal.Result[FormResult])
	if !ok || !d.pollForm {
		return nil, false
	}
	d.pollForm = false
	if !result.Ok() || result.Value.Canceled {
		return nil, true
	}
	if result.Value.Toggles["remove"] {
		d.poll = nil

		return nil, true
	}

	duration, err := strconv.Atoi(result.Value.Values["duration"])
	if err != nil || duration <= 0 {
		d.problem = "Poll duration must be a positive number of minutes"

		return nil, true
	}
	options := make([]string, 0, 4)
	for index := 1; index <= 4; index++ {
		option := strings.TrimSpace(result.Value.Values[fmt.Sprintf("option%d", index)])
		if option != "" {
			options = append(options, option)
		}
	}
	if len(options) < 2 {
		d.problem = "A poll needs at least two options"

		return nil, true
	}
	d.poll = &north.CreatePoll{Options: options, DurationMinutes: duration}

	return nil, true
}

func (d *Compose) openPoll(ctx *reactea.Ctx) tea.Cmd {
	if d.editID != "" {
		d.problem = "Polls cannot be changed while editing a post"

		return nil
	}
	values := make([]string, 4)
	duration := "1440"
	if d.poll != nil {
		copy(values, d.poll.Options)
		duration = strconv.Itoa(d.poll.DurationMinutes)
	}
	d.pollForm = true

	return modal.PushAt(ctx, NewForm(d.theme, "Poll", "Save",
		Field{Key: "option1", Label: "Option 1", Value: values[0], Required: true},
		Field{Key: "option2", Label: "Option 2", Value: values[1], Required: true},
		Field{Key: "option3", Label: "Option 3", Value: values[2]},
		Field{Key: "option4", Label: "Option 4", Value: values[3]},
		Field{Key: "duration", Label: "Duration in minutes", Value: duration, Required: true},
		Field{Key: "remove", Label: "Remove poll", Kind: ToggleField},
	), Placement(ctx, 64, 17))
}

func (d *Compose) addThreadItem() tea.Cmd {
	if !d.threadable {
		return nil
	}
	if len(d.thread) >= 24 {
		d.problem = "A thread can have at most 25 posts"

		return nil
	}
	if currentThreadItemEmpty(d.input.Widget.Value(), d.mediaIDs, d.poll) {
		d.problem = "Write something or attach media before adding another post"

		return nil
	}
	if length := northTextLength(d.input.Widget.Value()); length > 280 {
		d.problem = fmt.Sprintf("Post is %d units; the limit is 280", length)

		return nil
	}
	d.thread = append(d.thread, d.currentThreadItem())
	d.input.Widget.SetValue("")
	d.mediaIDs = nil
	d.media = nil
	d.poll = nil
	d.mediaNotice = ""
	d.problem = ""

	return d.input.Widget.Focus()
}

func (d *Compose) undoThreadItem() tea.Cmd {
	if !d.threadable || len(d.thread) == 0 {
		return nil
	}
	if !currentThreadItemEmpty(d.input.Widget.Value(), d.mediaIDs, d.poll) {
		d.problem = "Clear the current post before restoring the previous one"

		return nil
	}
	last := d.thread[len(d.thread)-1]
	d.thread = d.thread[:len(d.thread)-1]
	d.input.Widget.SetValue(last.Text)
	d.mediaIDs = append([]string(nil), last.MediaIDs...)
	d.poll = clonePoll(last.Poll)
	d.problem = ""

	return d.input.Widget.Focus()
}

func (d *Compose) currentThreadItem() north.ThreadItem {
	return north.ThreadItem{
		Text:     d.input.Widget.Value(),
		MediaIDs: append([]string(nil), d.mediaIDs...),
		Poll:     clonePoll(d.poll),
	}
}

func (d *Compose) renderPostOptions(width int) []string {
	lines := make([]string, 0, 2)
	if len(d.thread) > 0 {
		lines = append(lines, d.theme.Active.Render(ui.Clip(fmt.Sprintf("Thread · %d post(s) ready · Alt+U restores the previous post", len(d.thread)), width)))
	}
	if d.poll != nil {
		label := fmt.Sprintf("Poll · %d options · %s", len(d.poll.Options), pollDuration(d.poll.DurationMinutes))
		lines = append(lines, d.theme.Dim.Render(ui.Clip(label, width)))
	}

	return lines
}

func pollDuration(minutes int) string {
	if minutes%1440 == 0 {
		return fmt.Sprintf("%dd", minutes/1440)
	}
	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}

	return fmt.Sprintf("%dm", minutes)
}

func currentThreadItemEmpty(text string, mediaIDs []string, _ *north.CreatePoll) bool {
	return strings.TrimSpace(text) == "" && len(mediaIDs) == 0
}

func clonePoll(poll *north.CreatePoll) *north.CreatePoll {
	if poll == nil {
		return nil
	}

	return &north.CreatePoll{
		Options:         append([]string(nil), poll.Options...),
		DurationMinutes: poll.DurationMinutes,
	}
}

func cloneThreadItems(items []north.ThreadItem) []north.ThreadItem {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]north.ThreadItem, len(items))
	for index, item := range items {
		cloned[index] = north.ThreadItem{
			Text:        item.Text,
			MediaIDs:    append([]string(nil), item.MediaIDs...),
			ReplyPolicy: item.ReplyPolicy,
			Poll:        clonePoll(item.Poll),
		}
	}

	return cloned
}
