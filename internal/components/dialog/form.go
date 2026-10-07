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

type FieldKind uint8

const (
	TextField FieldKind = iota
	ToggleField
)

type Field struct {
	Key         string
	Label       string
	Placeholder string
	Value       string
	Required    bool
	Kind        FieldKind
	Checked     bool
}

type FormResult struct {
	Values   map[string]string
	Toggles  map[string]bool
	Canceled bool
}

type formField struct {
	Field
	input textinput.Model
}

type Form struct {
	reactea.BasicComponent

	theme   ui.Theme
	title   string
	button  string
	fields  []formField
	focused int
	problem string
}

func NewForm(theme ui.Theme, title, button string, fields ...Field) *Form {
	form := &Form{theme: theme, title: title, button: button, fields: make([]formField, len(fields))}
	for index, field := range fields {
		input := textinput.New()
		input.Prompt = ""
		input.Placeholder = field.Placeholder
		input.SetValue(field.Value)
		input.SetVirtualCursor(false)
		input.Blur()
		form.fields[index] = formField{Field: field, input: input}
	}
	form.focus(0)

	return form
}

func (f *Form) Init(*reactea.Ctx) tea.Cmd {
	if len(f.fields) == 0 || f.fields[f.focused].Kind != TextField {
		return nil
	}

	return f.fields[f.focused].input.Focus()
}

func (f *Form) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if !inside {
			return nil
		}
		style := f.theme.Dialog
		innerWidth := max(1, ctx.Width()-style.GetHorizontalFrameSize())
		footerRow := ctx.Height() - style.GetBorderBottomSize() - 1
		if y == footerRow {
			buttonWidth := lipgloss.Width(f.saveButton())
			if x >= style.GetBorderLeftSize()+innerWidth-buttonWidth {
				return f.save(ctx)
			}
			if x < style.GetBorderLeftSize()+12 {
				return modal.Return(ctx, FormResult{Canceled: true})
			}
		}
		row := y - style.GetBorderTopSize() - 2
		if row >= 0 && row/2 < len(f.fields) {
			f.focus(row / 2)
			if f.fields[f.focused].Kind == ToggleField {
				f.fields[f.focused].Checked = !f.fields[f.focused].Checked
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Return(ctx, FormResult{Canceled: true})
	case reactea.Key(msg, "ctrl+s", "ctrl+enter", "ctrl+j", "ctrl+m", "alt+enter"):
		return f.save(ctx)
	case reactea.Key(msg, "tab", "down"):
		f.focus((f.focused + 1) % max(1, len(f.fields)))

		return f.focusCommand()
	case reactea.Key(msg, "shift+tab", "up"):
		f.focus((f.focused + max(1, len(f.fields)) - 1) % max(1, len(f.fields)))

		return f.focusCommand()
	case reactea.Key(msg, "space") && len(f.fields) > 0 && f.fields[f.focused].Kind == ToggleField:
		f.fields[f.focused].Checked = !f.fields[f.focused].Checked

		return nil
	}

	if len(f.fields) == 0 || f.fields[f.focused].Kind != TextField {
		return nil
	}
	f.problem = ""
	updated, command := f.fields[f.focused].input.Update(msg)
	f.fields[f.focused].input = updated

	return command
}

func (f *Form) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := f.theme.Dialog
	innerWidth := max(1, width-style.GetHorizontalFrameSize())
	innerHeight := max(1, height-style.GetVerticalFrameSize())
	lines := []string{f.theme.ModalTitle.Render(ui.Clip(f.title, innerWidth))}
	for index := range f.fields {
		field := &f.fields[index]
		label := f.theme.Dim.Render(field.Label)
		if index == f.focused {
			label = f.theme.Active.Render(field.Label)
		}
		value := ""
		if field.Kind == ToggleField {
			mark := " "
			if field.Checked {
				mark = "x"
			}
			value = "  [" + mark + "] " + field.Label
		} else {
			field.input.SetWidth(max(1, innerWidth-2))
			value = "  " + field.input.View()
		}
		lines = append(lines, label, value)
	}
	body := ui.Fit(strings.Join(lines, "\n"), innerWidth, max(0, innerHeight-2))
	hint := f.theme.Dim.Render("Esc  Close · Tab  Next")
	if f.problem != "" {
		hint = f.theme.Bad.Render(f.problem)
	}
	footer := ui.Sides(hint, f.saveButton(), innerWidth)

	if len(f.fields) > 0 && f.fields[f.focused].Kind == TextField {
		cursor := f.fields[f.focused].input.Cursor()
		if cursor != nil {
			cursor.X += style.GetBorderLeftSize() + 2
			cursor.Y = style.GetBorderTopSize() + 2 + f.focused*2
			ctx.SetCursor(cursor)
		}
	}

	return RenderDialog(style, body+"\n"+footer, width, height)
}

func (f *Form) focus(index int) {
	if len(f.fields) == 0 {
		f.focused = 0

		return
	}
	index = min(max(0, index), len(f.fields)-1)
	for fieldIndex := range f.fields {
		if fieldIndex == index && f.fields[fieldIndex].Kind == TextField {
			f.fields[fieldIndex].input.Focus()
		} else {
			f.fields[fieldIndex].input.Blur()
		}
	}
	f.focused = index
}

func (f *Form) focusCommand() tea.Cmd {
	if len(f.fields) == 0 || f.fields[f.focused].Kind != TextField {
		return nil
	}

	return f.fields[f.focused].input.Focus()
}

func (f *Form) save(ctx *reactea.Ctx) tea.Cmd {
	result := FormResult{Values: make(map[string]string), Toggles: make(map[string]bool)}
	for index := range f.fields {
		field := &f.fields[index]
		if field.Kind == ToggleField {
			result.Toggles[field.Key] = field.Checked

			continue
		}
		value := strings.TrimSpace(field.input.Value())
		if field.Required && value == "" {
			f.problem = field.Label + " cannot be empty"
			f.focus(index)

			return nil
		}
		result.Values[field.Key] = value
	}

	return modal.Return(ctx, result)
}

func (f *Form) saveButton() string {
	return f.theme.Button.Padding(0, 1).Render(f.button)
}

var _ reactea.Component = (*Form)(nil)
