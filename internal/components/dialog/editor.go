package dialog

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type TextEditor struct {
	reactea.Wrapper

	theme   ui.Theme
	input   *reactea.ReactifiedWidget[textarea.Model]
	title   string
	button  string
	problem string
}

func NewTextEditor(theme ui.Theme, title, placeholder, initial, button string) *TextEditor {
	model := textarea.New()
	model.Prompt = ""
	model.Placeholder = placeholder
	model.ShowLineNumbers = false
	model.SetVirtualCursor(false)
	styles := model.Styles()
	styles.Focused.CursorLine = styles.Focused.CursorLine.UnsetBackground()
	model.SetStyles(styles)
	model.SetValue(initial)
	model.Focus()
	input := reactea.ReactifyWidget(model).OnResize(func(model textarea.Model, width, height int) textarea.Model {
		model.SetWidth(max(1, width))
		model.SetHeight(max(1, height))

		return model
	})

	return &TextEditor{
		Wrapper: reactea.Wrap(input),
		theme:   theme,
		input:   input,
		title:   title,
		button:  button,
	}
}

func (d *TextEditor) Init(ctx *reactea.Ctx) tea.Cmd {
	return tea.Batch(d.Wrapper.Init(d.inputCtx(ctx)), d.input.Widget.Focus())
}

func (d *TextEditor) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && y == ctx.Height()-d.theme.Dialog.GetBorderBottomSize()-1 {
			innerWidth := max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())
			buttonWidth := lipgloss.Width(d.saveButton())
			buttonLeft := d.theme.Dialog.GetBorderLeftSize() + innerWidth - buttonWidth
			if x >= buttonLeft {
				return d.save(ctx)
			}
			if x < d.theme.Dialog.GetBorderLeftSize()+12 {
				return modal.Return(ctx, TextResult{Canceled: true})
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Return(ctx, TextResult{Canceled: true})
	case reactea.Key(msg, "ctrl+s", "ctrl+enter", "ctrl+j", "ctrl+m", "alt+enter"):
		return d.save(ctx)
	}
	d.problem = ""

	return d.Wrapper.Update(d.inputCtx(ctx), msg)
}

func (d *TextEditor) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := d.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	inputHeight := max(1, innerHeight-2)
	title := d.theme.ModalTitle.Render(ui.Clip(d.title, max(1, innerWidth-2)))
	body := ui.Fit(d.Wrapper.Render(d.inputCtx(ctx)), innerWidth, inputHeight)
	hint := d.theme.Dim.Render("Esc  Close · Ctrl+Enter saves")
	if d.problem != "" {
		hint = d.theme.Bad.Render(d.problem)
	}
	footer := ui.Sides(hint, d.saveButton(), innerWidth)

	return RenderDialog(style, title+"\n"+body+"\n"+footer, width, height)
}

func (d *TextEditor) save(ctx *reactea.Ctx) tea.Cmd {
	text := strings.TrimSpace(d.input.Widget.Value())
	if text == "" {
		d.problem = "Enter a message"

		return nil
	}

	return modal.Return(ctx, TextResult{Text: text})
}

func (d *TextEditor) saveButton() string {
	return d.theme.Button.Padding(0, 1).Render(d.button)
}

func (d *TextEditor) inputCtx(ctx *reactea.Ctx) *reactea.Ctx {
	style := d.theme.Dialog

	return ctx.Inset(
		style.GetBorderLeftSize(),
		style.GetBorderTopSize()+1,
		max(1, ctx.Width()-style.GetHorizontalFrameSize()),
		max(1, ctx.Height()-style.GetVerticalFrameSize()-2),
	)
}

var _ reactea.Component = (*TextEditor)(nil)
