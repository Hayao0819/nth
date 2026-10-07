package user

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type profileEditor struct {
	reactea.BasicComponent

	theme   ui.Theme
	inputs  []textinput.Model
	labels  []string
	focused int
	problem string
}

func newProfileEditor(theme ui.Theme, user north.User, media bool) *profileEditor {
	values := []string{user.Name, stringValue(user.Bio), stringValue(user.Location), stringValue(user.Website)}
	labels := []string{"Name", "Bio", "Location", "Website"}
	placeholders := []string{"Name", "Bio", "Location", "Website"}
	if media {
		values = append(values, "", "")
		labels = append(labels, "Avatar file · blank keeps current", "Header file · blank keeps current")
		placeholders = append(placeholders, "Path to image", "Path to image")
	}
	editor := &profileEditor{
		theme:  theme,
		inputs: make([]textinput.Model, len(values)),
		labels: labels,
	}
	for index := range values {
		input := textinput.New()
		input.Prompt = ""
		input.Placeholder = placeholders[index]
		input.SetValue(values[index])
		input.SetVirtualCursor(false)
		input.Blur()
		editor.inputs[index] = input
	}
	editor.inputs[0].Focus()

	return editor
}

func (e *profileEditor) Init(*reactea.Ctx) tea.Cmd {
	return e.inputs[e.focused].Focus()
}

func (e *profileEditor) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if !inside {
			return nil
		}
		style := e.theme.Dialog
		innerWidth := max(1, ctx.Width()-style.GetHorizontalFrameSize())
		footerRow := ctx.Height() - style.GetBorderBottomSize() - 1
		if y == footerRow {
			buttonWidth := lipgloss.Width(e.saveButton())
			if x >= style.GetBorderLeftSize()+innerWidth-buttonWidth {
				return e.save(ctx)
			}
			if x < style.GetBorderLeftSize()+12 {
				return modal.Dismiss(ctx)
			}
		}
		row := y - style.GetBorderTopSize() - 2
		if row >= 0 && row/2 < len(e.inputs) {
			e.focus(row / 2)
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Dismiss(ctx)
	case reactea.Key(msg, "ctrl+s", "ctrl+enter", "ctrl+j", "ctrl+m", "alt+enter"):
		return e.save(ctx)
	case reactea.Key(msg, "tab", "down"):
		e.focus((e.focused + 1) % len(e.inputs))

		return e.inputs[e.focused].Focus()
	case reactea.Key(msg, "shift+tab", "up"):
		e.focus((e.focused + len(e.inputs) - 1) % len(e.inputs))

		return e.inputs[e.focused].Focus()
	}
	e.problem = ""
	updated, command := e.inputs[e.focused].Update(msg)
	e.inputs[e.focused] = updated

	return command
}

func (e *profileEditor) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := e.theme.Dialog
	innerWidth := max(1, width-style.GetHorizontalFrameSize())
	innerHeight := max(1, height-style.GetVerticalFrameSize())
	inputWidth := max(1, innerWidth-2)
	lines := []string{e.theme.ModalTitle.Render("Edit profile")}
	for index := range e.inputs {
		e.inputs[index].SetWidth(inputWidth)
		label := e.theme.Dim.Render(e.labels[index])
		if index == e.focused {
			label = e.theme.Active.Render(e.labels[index])
		}
		lines = append(lines, label, "  "+e.inputs[index].View())
	}
	bodyHeight := max(0, innerHeight-2)
	body := ui.Fit(strings.Join(lines, "\n"), innerWidth, bodyHeight)
	hint := e.theme.Dim.Render("Esc  Close · Tab  Next")
	if e.problem != "" {
		hint = e.theme.Bad.Render(e.problem)
	}
	footer := ui.Sides(hint, e.saveButton(), innerWidth)

	cursor := e.inputs[e.focused].Cursor()
	if cursor != nil {
		cursor.X += style.GetBorderLeftSize() + 2
		cursor.Y = style.GetBorderTopSize() + 2 + e.focused*2
		ctx.SetCursor(cursor)
	}

	return dialog.RenderDialog(style, body+"\n"+footer, width, height)
}

func (e *profileEditor) focus(index int) {
	for inputIndex := range e.inputs {
		if inputIndex == index {
			e.inputs[inputIndex].Focus()
		} else {
			e.inputs[inputIndex].Blur()
		}
	}
	e.focused = index
}

func (e *profileEditor) save(ctx *reactea.Ctx) tea.Cmd {
	name := strings.TrimSpace(e.inputs[0].Value())
	if name == "" {
		e.problem = "Name cannot be empty"
		e.focus(0)

		return nil
	}

	update := domain.ProfileUpdate{
		Name:     name,
		Bio:      strings.TrimSpace(e.inputs[1].Value()),
		Location: strings.TrimSpace(e.inputs[2].Value()),
		Website:  strings.TrimSpace(e.inputs[3].Value()),
	}
	if len(e.inputs) > 4 {
		update.AvatarPath = strings.TrimSpace(e.inputs[4].Value())
		update.HeaderPath = strings.TrimSpace(e.inputs[5].Value())
	}

	return modal.Return(ctx, update)
}

func (e *profileEditor) saveButton() string {
	return e.theme.Button.Padding(0, 1).Render("Save")
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

var _ reactea.Component = (*profileEditor)(nil)
