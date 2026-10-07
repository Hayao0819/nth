package dialog

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type TextResult struct {
	Text     string
	Canceled bool
}

type TextPrompt struct {
	reactea.Wrapper

	theme       ui.Theme
	input       *reactea.ReactifiedWidget[textinput.Model]
	title       string
	placeholder string
	button      string
	problem     string
}

func NewTextPrompt(theme ui.Theme, title, placeholder, initial, button string) *TextPrompt {
	model := textinput.New()
	model.Prompt = ""
	model.Placeholder = placeholder
	model.SetValue(initial)
	model.SetVirtualCursor(false)
	model.Focus()
	input := reactea.ReactifyWidget(model).OnResize(func(model textinput.Model, width, _ int) textinput.Model {
		model.SetWidth(max(1, width))

		return model
	})

	return &TextPrompt{
		Wrapper:     reactea.Wrap(input),
		theme:       theme,
		input:       input,
		title:       title,
		placeholder: placeholder,
		button:      button,
	}
}

func (d *TextPrompt) Init(ctx *reactea.Ctx) tea.Cmd {
	return tea.Batch(d.Wrapper.Init(d.inputCtx(ctx)), d.input.Widget.Focus())
}

func (d *TextPrompt) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && y == ctx.Height()-d.theme.Dialog.GetBorderBottomSize()-1 {
			innerWidth := max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())
			buttonWidth := lipgloss.Width(d.buttonView())
			buttonLeft := d.theme.Dialog.GetBorderLeftSize() + innerWidth - buttonWidth
			if x >= buttonLeft {
				return d.submit(ctx)
			}
			if x < d.theme.Dialog.GetBorderLeftSize()+12 {
				return modal.Return(ctx, TextResult{Canceled: true})
			}
		}
	}
	switch {
	case reactea.Key(msg, "esc"):
		return modal.Return(ctx, TextResult{Canceled: true})
	case reactea.Key(msg, "enter"):
		return d.submit(ctx)
	}
	d.problem = ""

	return d.Wrapper.Update(d.inputCtx(ctx), msg)
}

func (d *TextPrompt) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := d.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	title := d.theme.ModalTitle.Render(ui.Clip(d.title, max(1, innerWidth-2)))
	input := d.input.Render(d.inputCtx(ctx))
	message := d.theme.Dim.Render(d.placeholder)
	if d.problem != "" {
		message = d.theme.Warn.Render(d.problem)
	}
	footer := ui.Sides(d.theme.Dim.Render("Esc  Cancel"), d.buttonView(), innerWidth)
	content := title + "\n\n" + input + "\n" + message
	content = ui.Fit(content, innerWidth, max(0, innerHeight-1)) + "\n" + footer

	return RenderDialog(style, content, width, height)
}

func (d *TextPrompt) submit(ctx *reactea.Ctx) tea.Cmd {
	value := strings.TrimSpace(d.input.Widget.Value())
	if value == "" {
		d.problem = "Enter a value"

		return nil
	}

	return modal.Return(ctx, TextResult{Text: value})
}

func (d *TextPrompt) buttonView() string {
	return d.theme.Button.Padding(0, 1).Render("Enter  " + d.button)
}

func (d *TextPrompt) inputCtx(ctx *reactea.Ctx) *reactea.Ctx {
	style := d.theme.Dialog
	return ctx.Inset(
		style.GetBorderLeftSize(),
		style.GetBorderTopSize()+2,
		max(1, ctx.Width()-style.GetHorizontalFrameSize()),
		1,
	)
}

var _ reactea.Component = (*TextPrompt)(nil)
