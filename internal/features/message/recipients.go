package message

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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

const recipientDebounce = 250 * time.Millisecond

type recipientPurpose uint8

const (
	recipientNone recipientPurpose = iota
	recipientConversation
	recipientMembers
)

type recipientResult struct {
	Handles  []string
	Canceled bool
}

type recipientDebounceMsg struct {
	target *recipientPicker
	seq    uint64
	query  string
}

type recipientLoadedMsg struct {
	target *recipientPicker
	query  string
	page   north.UserPage
	err    error
}

type recipientPicker struct {
	reactea.BasicComponent

	api      domain.MessageRecipientAPI
	theme    ui.Theme
	input    *reactea.ReactifiedWidget[textinput.Model]
	users    []north.User
	selected int
	chosen   map[string]north.User
	seq      uint64
	loading  bool
	err      error
	title    string
	button   string
}

func (s *Screen) openRecipients(ctx *reactea.Ctx, purpose recipientPurpose) tea.Cmd {
	if (purpose == recipientConversation && s.write == nil) || (purpose == recipientMembers && s.group == nil) {
		return nil
	}
	if s.recipient == nil {
		if purpose == recipientMembers {
			return s.promptFor(ctx, promptMembers)
		}

		return s.promptFor(ctx, promptConversation)
	}
	title, button := "New message", "Start"
	if purpose == recipientMembers {
		title, button = "Add people", "Add"
	}
	s.recipients = purpose
	picker := newRecipientPicker(s.recipient, s.theme, title, button)

	return modal.PushAt(ctx, picker, dialog.Placement(ctx, 64, 20))
}

func newRecipientPicker(api domain.MessageRecipientAPI, theme ui.Theme, title, button string) *recipientPicker {
	model := textinput.New()
	model.Prompt = ""
	model.Placeholder = "Search accounts"
	model.SetVirtualCursor(false)
	model.Focus()
	input := reactea.ReactifyWidget(model).OnResize(func(model textinput.Model, width, _ int) textinput.Model {
		model.SetWidth(max(1, width))

		return model
	})

	return &recipientPicker{
		api: api, theme: theme, input: input, chosen: make(map[string]north.User), title: title, button: button,
	}
}

func (p *recipientPicker) Init(ctx *reactea.Ctx) tea.Cmd {
	return tea.Batch(p.input.Init(p.inputCtx(ctx)), p.input.Widget.Focus(), p.search(ctx.Context(), ""))
}

func (p *recipientPicker) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case recipientDebounceMsg:
		if msg.target != p || msg.seq != p.seq {
			return nil
		}

		return p.search(ctx.Context(), msg.query)
	case recipientLoadedMsg:
		if msg.target != p || msg.query != strings.TrimSpace(p.input.Widget.Value()) {
			return nil
		}
		p.loading = false
		p.err = msg.err
		if msg.err == nil {
			p.users = append([]north.User(nil), msg.page.Items...)
			p.selected = min(p.selected, max(0, len(p.users)-1))
		}

		return nil
	case tea.MouseClickMsg:
		return p.click(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Return(ctx, recipientResult{Canceled: true})
	case reactea.Key(msg, "down", "ctrl+n"):
		p.selected = min(p.selected+1, max(0, len(p.users)-1))

		return nil
	case reactea.Key(msg, "up", "ctrl+p"):
		p.selected = max(0, p.selected-1)

		return nil
	case reactea.Key(msg, "tab"):
		p.toggleSelected()

		return nil
	case reactea.Key(msg, "enter"):
		return p.submit(ctx)
	}

	if !reactea.IsInput(msg) {
		return nil
	}
	before := p.input.Widget.Value()
	command := p.input.Update(p.inputCtx(ctx), msg)
	after := p.input.Widget.Value()
	if before == after {
		return command
	}
	p.seq++
	p.loading = true
	p.err = nil
	seq, query := p.seq, strings.TrimSpace(after)

	return tea.Batch(command, tea.Tick(recipientDebounce, func(time.Time) tea.Msg {
		return recipientDebounceMsg{target: p, seq: seq, query: query}
	}))
}

func (p *recipientPicker) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := p.theme.Dialog
	innerWidth := max(1, width-style.GetHorizontalFrameSize())
	innerHeight := max(1, height-style.GetVerticalFrameSize())
	lines := []string{p.theme.ModalTitle.Render(p.title), "", p.input.Render(p.inputCtx(ctx))}
	status := "Tab  Select · Enter  " + p.button
	if p.loading {
		status = "Searching…"
	} else if p.err != nil {
		status = ui.FriendlyError(p.err)
	} else if len(p.chosen) > 0 {
		status = fmt.Sprintf("%d selected · Tab remove · Enter %s", len(p.chosen), strings.ToLower(p.button))
	}
	lines = append(lines, p.theme.Dim.Render(ui.Clip(status, innerWidth)), "")
	room := max(0, innerHeight-6)
	for index, user := range p.users[:min(len(p.users), room/2)] {
		marker := " "
		key := strings.ToLower(user.Handle)
		if _, ok := p.chosen[key]; ok {
			marker = "✓"
		} else if index == p.selected {
			marker = ">"
		}
		name := ui.SafeInline(user.Name)
		if name == "" {
			name = "@" + ui.SafeInline(user.Handle)
		}
		lines = append(lines,
			p.theme.Name.Render(ui.Clip(marker+"  "+name, innerWidth)),
			p.theme.Handle.Render(ui.Clip("   @"+strings.TrimPrefix(ui.SafeInline(user.Handle), "@"), innerWidth)),
		)
	}
	if len(p.users) == 0 && !p.loading && p.err == nil {
		lines = append(lines, p.theme.Dim.Render("No accounts found"))
	}
	body := ui.Fit(strings.Join(lines, "\n"), innerWidth, max(0, innerHeight-1))
	footer := ui.Sides(p.theme.Dim.Render("Esc  Cancel"), p.theme.Button.Padding(0, 1).Render("Enter  "+p.button), innerWidth)

	return dialog.RenderDialog(style, body+"\n"+footer, width, height)
}

func (p *recipientPicker) search(ctx context.Context, query string) tea.Cmd {
	p.loading = true
	p.err = nil

	return func() tea.Msg {
		page, _, err := p.api.DMRecipients(ctx, query, "")

		return recipientLoadedMsg{target: p, query: query, page: page, err: err}
	}
}

func (p *recipientPicker) toggleSelected() {
	if p.selected < 0 || p.selected >= len(p.users) {
		return
	}
	user := p.users[p.selected]
	key := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(user.Handle), "@"))
	if key == "" {
		return
	}
	if _, ok := p.chosen[key]; ok {
		delete(p.chosen, key)
	} else if len(p.chosen) < 19 {
		p.chosen[key] = user
	}
}

func (p *recipientPicker) submit(ctx *reactea.Ctx) tea.Cmd {
	if len(p.chosen) == 0 {
		p.toggleSelected()
	}
	if len(p.chosen) == 0 {
		return nil
	}
	handles := make([]string, 0, len(p.chosen))
	for handle := range p.chosen {
		handles = append(handles, handle)
	}
	sort.Strings(handles)

	return modal.Return(ctx, recipientResult{Handles: handles})
}

func (p *recipientPicker) click(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside || msg.Button != tea.MouseLeft {
		return nil
	}
	style := p.theme.Dialog
	if y == ctx.Height()-style.GetBorderBottomSize()-1 {
		innerWidth := max(1, ctx.Width()-style.GetHorizontalFrameSize())
		button := p.theme.Button.Padding(0, 1).Render("Enter  " + p.button)
		buttonLeft := style.GetBorderLeftSize() + innerWidth - lipgloss.Width(button)
		if x >= buttonLeft {
			return p.submit(ctx)
		}
		if x < style.GetBorderLeftSize()+12 {
			return modal.Return(ctx, recipientResult{Canceled: true})
		}

		return nil
	}
	row := y - style.GetBorderTopSize() - 5
	if row >= 0 {
		index := row / 2
		if index >= 0 && index < len(p.users) {
			p.selected = index
			p.toggleSelected()
		}
	}

	return nil
}

func (p *recipientPicker) inputCtx(ctx *reactea.Ctx) *reactea.Ctx {
	style := p.theme.Dialog

	return ctx.Inset(style.GetBorderLeftSize(), style.GetBorderTopSize()+2, max(1, ctx.Width()-style.GetHorizontalFrameSize()), 1)
}

var _ reactea.Component = (*recipientPicker)(nil)
