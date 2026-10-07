package bookmark

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type folderPicker struct {
	reactea.BasicComponent

	api      domain.BookmarkFolderAPI
	theme    ui.Theme
	postID   string
	folders  []domain.BookmarkFolder
	selected int
	top      int
	member   map[string]bool
	loading  bool
	busy     bool
	err      error
	notice   string
}

type foldersLoadedMsg struct {
	target   *folderPicker
	folders  []domain.BookmarkFolder
	members  []domain.BookmarkFolder
	response *north.Response
	err      error
}

func (m foldersLoadedMsg) Response() *north.Response { return m.response }

type folderToggledMsg struct {
	target   *folderPicker
	id       string
	member   bool
	response *north.Response
	err      error
}

func (m folderToggledMsg) Response() *north.Response { return m.response }

func newFolderPicker(api domain.BookmarkFolderAPI, theme ui.Theme, postID string) *folderPicker {
	return &folderPicker{api: api, theme: theme, postID: postID, member: make(map[string]bool)}
}

func (p *folderPicker) Init(ctx *reactea.Ctx) tea.Cmd {
	if p.api == nil || p.loading {
		return nil
	}
	p.loading = true

	return func() tea.Msg {
		folders, response, err := p.api.BookmarkFolders(ctx.Context(), "")
		if err != nil {
			return foldersLoadedMsg{target: p, response: response, err: err}
		}
		members, memberResponse, err := p.api.PostBookmarkFolders(ctx.Context(), p.postID, "")
		if memberResponse != nil {
			response = memberResponse
		}

		return foldersLoadedMsg{target: p, folders: folders, members: members, response: response, err: err}
	}
}

func (p *folderPicker) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case foldersLoadedMsg:
		if msg.target != p {
			return nil
		}
		p.loading = false
		p.err = msg.err
		if msg.err == nil {
			p.folders = append([]domain.BookmarkFolder(nil), msg.folders...)
			clear(p.member)
			for _, folder := range msg.members {
				p.member[folder.ID] = true
			}
		}

		return nil
	case folderToggledMsg:
		if msg.target != p {
			return nil
		}
		p.busy = false
		if msg.err != nil {
			p.notice = ui.FriendlyError(msg.err)

			return nil
		}
		p.member[msg.id] = msg.member
		for index := range p.folders {
			if p.folders[index].ID != msg.id {
				continue
			}
			if msg.member {
				p.folders[index].Count++
			} else {
				p.folders[index].Count = max(0, p.folders[index].Count-1)
			}
			break
		}
		p.notice = "Folder updated"

		return nil
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			p.move(1, p.room(ctx.Height()))
		} else if msg.Button == tea.MouseWheelUp {
			p.move(-1, p.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return nil
		}
		_, y, inside := reactea.Mouse(ctx, msg)
		if !inside {
			return nil
		}
		row := y - p.theme.Dialog.GetBorderTopSize() - 2
		index := p.top + row
		if row >= 0 && row < p.room(ctx.Height()) && index >= 0 && index < len(p.folders) {
			p.selected = index

			return p.toggle(ctx.Context())
		}

		return nil
	}

	switch {
	case reactea.Key(msg, "esc"):
		return modal.Dismiss(ctx)
	case reactea.Key(msg, "j", "down"):
		p.move(1, p.room(ctx.Height()))
	case reactea.Key(msg, "k", "up"):
		p.move(-1, p.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		p.selected, p.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if len(p.folders) > 0 {
			p.selected = len(p.folders) - 1
			p.ensureVisible(p.room(ctx.Height()))
		}
	case reactea.Key(msg, "enter", "space"):
		return p.toggle(ctx.Context())
	case reactea.Key(msg, ".") && p.err != nil:
		p.err = nil

		return p.Init(ctx)
	}

	return nil
}

func (p *folderPicker) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := p.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	header := p.theme.ModalTitle.Render(ui.Clip("Add to folder", max(1, innerWidth-2)))
	room := max(0, innerHeight-3)
	body := p.renderFolders(innerWidth, room)
	left := p.theme.Dim.Render("Esc  Close")
	right := p.theme.Dim.Render("j/k  Move")
	if p.busy {
		right = p.theme.Dim.Render("Saving…")
	} else if p.notice != "" {
		right = p.theme.Warn.Render(ui.Clip(p.notice, max(1, innerWidth-12)))
	}
	footer := ui.Sides(left, right, innerWidth)
	content := header + "\n\n" + body + "\n" + footer

	return dialog.RenderDialog(style, content, width, height)
}

func (p *folderPicker) renderFolders(width, room int) string {
	if len(p.folders) == 0 {
		message := "No bookmark folders"
		switch {
		case p.loading:
			message = "Loading folders…"
		case p.err != nil:
			message = ui.FriendlyError(p.err) + "\n\nPress . to try again"
		}

		return ui.Fit(message, width, room)
	}
	lines := make([]string, 0, room)
	for index := p.top; index < len(p.folders) && len(lines) < room; index++ {
		folder := p.folders[index]
		cursor := " "
		if index == p.selected {
			cursor = p.theme.Active.Render(">")
		}
		check := "[ ]"
		if p.member[folder.ID] {
			check = p.theme.Active.Render("[x]")
		}
		name := ui.SafeInline(folder.Name)
		if name == "" {
			name = "Untitled"
		}
		left := cursor + " " + check + " " + p.theme.Name.Render(name)
		lines = append(lines, ui.Sides(left, p.theme.Dim.Render(stringCount(folder.Count)), width))
	}

	return ui.Fit(strings.Join(lines, "\n"), width, room)
}

func (p *folderPicker) toggle(ctx context.Context) tea.Cmd {
	if p.api == nil || p.busy || p.selected < 0 || p.selected >= len(p.folders) {
		return nil
	}
	folder := p.folders[p.selected]
	want := !p.member[folder.ID]
	p.busy = true
	p.notice = ""

	return func() tea.Msg {
		var response *north.Response
		var err error
		if want {
			_, response, err = p.api.AddBookmarkToFolder(ctx, folder.ID, p.postID)
		} else {
			_, response, err = p.api.RemoveBookmarkFromFolder(ctx, folder.ID, p.postID)
		}

		return folderToggledMsg{target: p, id: folder.ID, member: want, response: response, err: err}
	}
}

func (p *folderPicker) move(delta, room int) {
	if len(p.folders) == 0 {
		return
	}
	p.selected = min(max(0, p.selected+delta), len(p.folders)-1)
	p.ensureVisible(room)
}

func (p *folderPicker) ensureVisible(room int) {
	if p.selected < p.top {
		p.top = p.selected
	}
	if p.selected >= p.top+max(1, room) {
		p.top = p.selected - max(1, room) + 1
	}
}

func (p *folderPicker) room(height int) int {
	return max(0, height-p.theme.Dialog.GetVerticalFrameSize()-3)
}

func stringCount(value int) string {
	if value == 1 {
		return "1 post"
	}

	return fmt.Sprintf("%d posts", value)
}

var _ reactea.Component = (*folderPicker)(nil)
