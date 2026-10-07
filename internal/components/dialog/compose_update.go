package dialog

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/reactea/v2"
)

func (d *Compose) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if command, handled := d.updateMedia(ctx, msg); handled {
		return command
	}
	if command, handled := d.updatePostOptions(ctx, msg); handled {
		return command
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && y == ctx.Height()-d.theme.Dialog.GetBorderBottomSize()-1 {
			innerWidth := max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())
			buttonWidth := lipgloss.Width(d.theme.Button.Padding(0, 1).Render("Post"))
			buttonLeft := d.theme.Dialog.GetBorderLeftSize() + innerWidth - buttonWidth
			if x >= buttonLeft && x < d.theme.Dialog.GetBorderLeftSize()+innerWidth {
				return d.send(ctx)
			}
			if x < d.theme.Dialog.GetBorderLeftSize()+12 {
				return d.cancel(ctx)
			}
			if x < d.theme.Dialog.GetBorderLeftSize()+30 {
				return d.openSaved(ctx)
			}
			if x < d.theme.Dialog.GetBorderLeftSize()+48 && d.mediaAPI != nil {
				return d.promptMedia(ctx, mediaPromptPath)
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return d.cancel(ctx)
	case reactea.Key(msg, "ctrl+o"):
		return d.openSaved(ctx)
	case reactea.Key(msg, "alt+a"):
		return d.promptMedia(ctx, mediaPromptPath)
	case reactea.Key(msg, "alt+x"):
		return d.removeLastMedia(ctx.Context())
	case reactea.Key(msg, "alt+t"):
		return d.promptMedia(ctx, mediaPromptAlt)
	case reactea.Key(msg, "alt+w"):
		return d.cycleMediaWarning(ctx.Context())
	case reactea.Key(msg, "alt+p"):
		return d.openPoll(ctx)
	case reactea.Key(msg, "alt+n"):
		return d.addThreadItem()
	case reactea.Key(msg, "alt+u"):
		return d.undoThreadItem()
	case reactea.Key(msg, "ctrl+s", "ctrl+enter", "ctrl+j", "ctrl+m", "alt+enter"):
		return d.send(ctx)
	}

	d.problem = ""

	return d.Wrapper.Update(ctx, msg)
}
