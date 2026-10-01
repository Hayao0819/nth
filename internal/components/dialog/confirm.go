package dialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type Confirmation struct {
	Accepted bool
}

type Confirm struct {
	reactea.BasicComponent

	theme   ui.Theme
	title   string
	message string
	label   string
}

func NewConfirm(theme ui.Theme, title, message, label string) *Confirm {
	return &Confirm{theme: theme, title: title, message: message, label: label}
}

func (d *Confirm) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && y == ctx.Height()-d.theme.Dialog.GetBorderBottomSize()-1 {
			innerWidth := max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())
			buttonWidth := lipgloss.Width(d.button())
			buttonLeft := d.theme.Dialog.GetBorderLeftSize() + innerWidth - buttonWidth
			switch {
			case x >= buttonLeft && x < d.theme.Dialog.GetBorderLeftSize()+innerWidth:
				return modal.Return(ctx, Confirmation{Accepted: true})
			case x < d.theme.Dialog.GetBorderLeftSize()+12:
				return modal.Return(ctx, Confirmation{})
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc", "n"):
		return modal.Return(ctx, Confirmation{})
	case reactea.Key(msg, "enter", "y"):
		return modal.Return(ctx, Confirmation{Accepted: true})
	default:
		return nil
	}
}

func (d *Confirm) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := d.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	bodyHeight := max(0, innerHeight-2)

	header := d.theme.ModalTitle.Render(ui.Clip(d.title, max(1, innerWidth-2)))
	body := strings.Join(ui.WrappedLines(d.message, innerWidth), "\n")
	footer := ui.Sides(d.theme.Dim.Render("Esc  Cancel"), d.button(), innerWidth)
	content := header + "\n" + ui.Fit(body, innerWidth, bodyHeight) + "\n" + footer

	return RenderDialog(style, content, width, height)
}

func (d *Confirm) button() string {
	return d.theme.Bad.Bold(true).Padding(0, 1).Render("Enter  " + d.label)
}

var _ reactea.Component = (*Confirm)(nil)
