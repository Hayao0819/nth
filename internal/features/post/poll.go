package post

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/charmbracelet/x/ansi"
)

type pollChoice struct{ OptionID string }

type pollPicker struct {
	reactea.BasicComponent

	theme    ui.Theme
	poll     north.Poll
	selected int
}

type pollVotedMsg struct {
	target   *Screen
	post     north.Post
	response *north.Response
	err      error
}

func (m pollVotedMsg) Response() *north.Response { return m.response }

func (d *Screen) openPoll(ctx *reactea.Ctx) tea.Cmd {
	target := d.post.DisplayPost()
	if d.polls == nil || target == nil || !canVote(target.Poll) || d.pollBusy {
		return nil
	}

	return modal.PushAt(ctx, &pollPicker{theme: d.theme, poll: *target.Poll}, dialog.Placement(ctx, 60, max(8, len(target.Poll.Options)+5)))
}

func (d *Screen) vote(ctx context.Context, optionID string) tea.Cmd {
	target := d.post.DisplayPost()
	if d.polls == nil || target == nil || target.ID == "" || optionID == "" || d.pollBusy {
		return nil
	}
	d.pollBusy = true
	d.notice = "Voting…"
	postID := target.ID

	return func() tea.Msg {
		post, response, err := d.polls.VotePoll(ctx, postID, optionID)

		return pollVotedMsg{target: d, post: post, response: response, err: err}
	}
}

func (d *Screen) applyPollVote(msg pollVotedMsg) {
	if msg.target != d {
		return
	}
	d.pollBusy = false
	if msg.err != nil {
		d.notice = ui.FriendlyError(msg.err)

		return
	}
	if target := d.post.DisplayPost(); target != nil {
		*target = msg.post
	}
	d.notice = "Vote recorded"
}

func (d *Screen) pollOptionAt(x, row, width int) (string, bool) {
	target := d.post.DisplayPost()
	if d.polls == nil || target == nil || !canVote(target.Poll) {
		return "", false
	}
	lines := d.contentLines(width)
	if row < 0 || row >= len(lines) {
		return "", false
	}
	line := ansi.Strip(lines[row])
	for _, option := range target.Poll.Options {
		label := ui.SafeInline(option.Label)
		if ui.TextAt(line, x, label) {
			return option.ID, true
		}
	}

	return "", false
}

func canVote(poll *north.Poll) bool {
	return poll != nil && !poll.Ended && poll.ViewerOptionID == nil && len(poll.Options) > 0
}

func (p *pollPicker) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		_, y, inside := reactea.Mouse(ctx, msg)
		if inside {
			row := y - p.theme.Dialog.GetBorderTopSize() - 2
			if row >= 0 && row < len(p.poll.Options) {
				p.selected = row

				return p.choose(ctx)
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Dismiss(ctx)
	case reactea.Key(msg, "j", "down"):
		p.selected = min(p.selected+1, len(p.poll.Options)-1)
	case reactea.Key(msg, "k", "up"):
		p.selected = max(0, p.selected-1)
	case reactea.Key(msg, "enter", "space"):
		return p.choose(ctx)
	}

	return nil
}

func (p *pollPicker) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := p.theme.Dialog
	innerWidth := max(1, width-style.GetHorizontalFrameSize())
	innerHeight := max(1, height-style.GetVerticalFrameSize())
	lines := []string{p.theme.ModalTitle.Render("Vote")}
	for index, option := range p.poll.Options {
		marker := " "
		labelStyle := p.theme.Name
		if index == p.selected {
			marker = p.theme.Active.Render(">")
			labelStyle = p.theme.Active
		}
		label := marker + "  " + labelStyle.Render(ui.SafeInline(option.Label))
		lines = append(lines, ui.Sides(label, p.theme.Dim.Render(fmt.Sprintf("%.0f%%", option.Percent)), innerWidth))
	}
	body := ui.Fit(strings.Join(lines, "\n"), innerWidth, max(0, innerHeight-1))
	footer := ui.Sides(p.theme.Dim.Render("Esc  Close"), p.theme.Button.Padding(0, 1).Render("Enter  Vote"), innerWidth)

	return dialog.RenderDialog(style, body+"\n"+footer, width, height)
}

func (p *pollPicker) choose(ctx *reactea.Ctx) tea.Cmd {
	if p.selected < 0 || p.selected >= len(p.poll.Options) {
		return nil
	}

	return modal.Return(ctx, pollChoice{OptionID: p.poll.Options[p.selected].ID})
}

var _ reactea.Component = (*pollPicker)(nil)
