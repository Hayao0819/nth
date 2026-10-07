package dialog

import (
	"fmt"
	"strings"

	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (d *Compose) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := d.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	inputHeight := max(1, innerHeight-2)

	title := "New post"
	target := d.replyTo
	if d.editID != "" {
		title = "Edit post"
	} else if target != nil {
		title = "Reply to @" + ui.SafeInline(target.Author.Handle)
	} else if d.quote != nil {
		target = d.quote
		title = "Quote @" + ui.SafeInline(target.Author.Handle)
	}
	if len(d.thread) > 0 {
		title = fmt.Sprintf("New thread · post %d", len(d.thread)+1)
	}
	if d.restored {
		title += " · Draft"
	}
	contextLine := ""
	inputTop := 1
	if target != nil {
		inputHeight = max(1, innerHeight-4)
		inputTop = 3
		preview := ui.SafeInline(target.Text)
		if target.Deleted {
			preview = "Post deleted"
		} else if target.HiddenReason != "" {
			preview = "Hidden: " + string(target.HiddenReason)
		} else if target.Unavailable {
			preview = "Post unavailable"
		}
		if preview == "" && len(target.Media) > 0 {
			preview = "[media]"
		}
		author := ui.SafeInline(target.Author.Name) + "  @" + ui.SafeInline(target.Author.Handle)
		contextLine = d.theme.Dim.Render(ui.Clip(author, innerWidth)) + "\n" +
			d.theme.Dim.Render(ui.Clip(preview, innerWidth)) + "\n"
	}
	optionLines := append(d.renderPostOptions(innerWidth), d.renderMedia(innerWidth)...)
	inputHeight = max(1, inputHeight-len(optionLines))

	inputCtx := ctx.Inset(
		style.GetBorderLeftSize(),
		style.GetBorderTopSize()+inputTop,
		innerWidth,
		inputHeight,
	)
	content := d.theme.ModalTitle.Render(ui.Clip(title, max(1, innerWidth-2))) + "\n" + contextLine +
		ui.Fit(d.Wrapper.Render(inputCtx), innerWidth, inputHeight) + "\n"
	if len(optionLines) > 0 {
		content += strings.Join(optionLines, "\n") + "\n"
	}

	length := northTextLength(d.input.Widget.Value())
	count := fmt.Sprintf("%d/280", length)
	if length > 280 {
		count = d.theme.Bad.Render(count)
	}
	buttonStyle := d.theme.Button
	if !d.canSend() || length > 280 {
		buttonStyle = d.theme.Dim
	}
	buttonLabel := "Post"
	if d.editID != "" {
		buttonLabel = "Save"
	}
	button := buttonStyle.Padding(0, 1).Render(buttonLabel)
	right := count + "  " + button
	hintText := "Esc close · Ctrl+O saved · Ctrl+S send · Alt+P poll"
	if d.mediaAPI != nil {
		hintText = "Esc close · Ctrl+S send · Alt+A media · Alt+P poll"
	}
	if d.threadable {
		hintText += " · Alt+N next"
	}
	hint := d.theme.Dim.Render(hintText)
	if d.problem != "" {
		hint = d.theme.Bad.Render(d.problem)
	}
	content += ui.Sides(hint, right, innerWidth)

	return RenderDialog(style, content, width, height)
}
