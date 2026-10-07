package user

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type membershipsLoadedMsg struct {
	target *listMembershipPicker
	items  []north.ListMembership
	err    error
}

type membershipChangedMsg struct {
	target   *listMembershipPicker
	id       string
	contains bool
	err      error
}

type listMembershipPicker struct {
	reactea.BasicComponent

	api      domain.ListMemberAPI
	theme    ui.Theme
	handle   string
	items    []north.ListMembership
	selected int
	loading  bool
	acting   bool
	err      error
	notice   string
}

func (d *Screen) openListMemberships(ctx *reactea.Ctx) tea.Cmd {
	if d.listMembers == nil || d.isOwnProfile() || strings.TrimSpace(d.user.Handle) == "" {
		return nil
	}
	picker := &listMembershipPicker{api: d.listMembers, theme: d.theme, handle: d.user.Handle}

	return modal.PushAt(ctx, picker, dialog.Placement(ctx, 60, 18))
}

func (p *listMembershipPicker) Init(ctx *reactea.Ctx) tea.Cmd {
	return p.load(ctx.Context())
}

func (p *listMembershipPicker) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case membershipsLoadedMsg:
		if msg.target != p {
			return nil
		}
		p.loading, p.err = false, msg.err
		if msg.err == nil {
			p.items = append([]north.ListMembership(nil), msg.items...)
			p.selected = min(p.selected, max(0, len(p.items)-1))
		}

		return nil
	case membershipChangedMsg:
		if msg.target != p {
			return nil
		}
		p.acting = false
		if msg.err != nil {
			p.notice = ui.FriendlyError(msg.err)

			return nil
		}
		for index := range p.items {
			if p.items[index].ID == msg.id {
				p.items[index].Contains = msg.contains
				break
			}
		}
		p.notice = "List membership updated"

		return nil
	case tea.MouseClickMsg:
		_, y, inside := reactea.Mouse(ctx, msg)
		if !inside || msg.Button != tea.MouseLeft {
			return nil
		}
		style := p.theme.Dialog
		if y >= ctx.Height()-style.GetBorderBottomSize()-1 {
			return modal.Dismiss(ctx)
		}
		row := y - style.GetBorderTopSize() - 2
		if row >= 0 && row/2 < len(p.items) {
			p.selected = row / 2

			return p.toggle(ctx.Context())
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Dismiss(ctx)
	case reactea.Key(msg, "j", "down"):
		p.selected = min(p.selected+1, max(0, len(p.items)-1))
	case reactea.Key(msg, "k", "up"):
		p.selected = max(0, p.selected-1)
	case reactea.Key(msg, "enter", "space"):
		return p.toggle(ctx.Context())
	case reactea.Key(msg, ".") && p.err != nil:
		return p.load(ctx.Context())
	}

	return nil
}

func (p *listMembershipPicker) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := p.theme.Dialog
	innerWidth := max(1, width-style.GetHorizontalFrameSize())
	innerHeight := max(1, height-style.GetVerticalFrameSize())
	lines := []string{p.theme.ModalTitle.Render("Lists for @" + ui.SafeInline(p.handle))}
	for index, item := range p.items[:min(len(p.items), max(0, (innerHeight-3)/2))] {
		mark := " "
		if item.Contains {
			mark = "✓"
		} else if index == p.selected {
			mark = ">"
		}
		lines = append(lines,
			ui.Clip(mark+"  "+p.theme.Name.Render(ui.SafeInline(item.Name)), innerWidth),
			"   "+p.theme.Dim.Render(ui.Clip(listMembershipMeta(item), max(1, innerWidth-3))),
		)
	}
	if len(p.items) == 0 {
		message := "No owned lists"
		if p.loading {
			message = "Loading lists…"
		} else if p.err != nil {
			message = ui.FriendlyError(p.err)
		}
		lines = append(lines, "", lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, message))
	}
	if p.notice != "" {
		lines = append(lines, p.theme.Dim.Render(ui.Clip(p.notice, innerWidth)))
	}
	body := ui.Fit(strings.Join(lines, "\n"), innerWidth, max(0, innerHeight-1))
	footer := ui.Sides(p.theme.Dim.Render("Esc  Close"), p.theme.Button.Padding(0, 1).Render("Enter  Toggle"), innerWidth)

	return dialog.RenderDialog(style, body+"\n"+footer, width, height)
}

func (p *listMembershipPicker) load(ctx context.Context) tea.Cmd {
	if p.loading {
		return nil
	}
	p.loading, p.err = true, nil

	return func() tea.Msg {
		items, _, err := p.api.ListMemberships(ctx, p.handle, "")

		return membershipsLoadedMsg{target: p, items: items, err: err}
	}
}

func (p *listMembershipPicker) toggle(ctx context.Context) tea.Cmd {
	if p.acting || p.selected < 0 || p.selected >= len(p.items) {
		return nil
	}
	item := p.items[p.selected]
	p.acting = true
	p.notice = "Updating list…"

	return func() tea.Msg {
		var err error
		if item.Contains {
			_, _, err = p.api.RemoveListMember(ctx, item.ID, p.handle)
		} else {
			_, _, err = p.api.AddListMember(ctx, item.ID, p.handle)
		}

		return membershipChangedMsg{target: p, id: item.ID, contains: !item.Contains, err: err}
	}
}

func listMembershipMeta(item north.ListMembership) string {
	meta := "public"
	if item.Private {
		meta = "private"
	}

	return meta + " · " + fmt.Sprintf("%d members", item.MemberCount)
}

var _ reactea.Component = (*listMembershipPicker)(nil)
