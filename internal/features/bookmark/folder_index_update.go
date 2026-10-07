package bookmark

import (
	"errors"
	"slices"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

var errNotDeleted = errors.New("north did not delete the bookmark folder")

func (s *FolderIndex) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case loadedMsg:
		if msg.target != s {
			return nil
		}
		s.loading = false
		s.err = msg.err
		if msg.err == nil {
			s.folders = append(s.folders[:0], msg.folders...)
			s.selected, s.top = 0, 0
		}

		return nil
	case changedMsg:
		return s.applyChanged(msg)
	case modal.Result[dialog.TextResult]:
		id := s.editingID
		s.editingID = ""
		if !msg.Ok() || msg.Value.Canceled {
			return nil
		}

		return s.save(ctx.Context(), id, msg.Value.Text)
	case modal.Result[dialog.Confirmation]:
		id := s.deletingID
		s.deletingID = ""
		if msg.Ok() && msg.Value.Accepted {
			return s.delete(ctx.Context(), id)
		}

		return nil
	case tea.MouseWheelMsg:
		if _, _, inside := reactea.Mouse(ctx, msg); !inside {
			return nil
		}
		if msg.Button == tea.MouseWheelDown {
			s.move(1, s.room(ctx.Height()))
		} else if msg.Button == tea.MouseWheelUp {
			s.move(-1, s.room(ctx.Height()))
		}

		return nil
	case tea.MouseClickMsg:
		return s.click(ctx, msg)
	}

	switch {
	case reactea.Key(msg, "esc", "left"):
		return pageheader.Back()
	case reactea.Key(msg, "j", "down"):
		s.move(1, s.room(ctx.Height()))
	case reactea.Key(msg, "k", "up"):
		s.move(-1, s.room(ctx.Height()))
	case reactea.Key(msg, "g", "home"):
		s.selected, s.top = 0, 0
	case reactea.Key(msg, "G", "end"):
		if len(s.folders) > 0 {
			s.selected = len(s.folders) - 1
			s.ensureVisible(s.room(ctx.Height()))
		}
	case reactea.Key(msg, "enter"):
		return s.open()
	case reactea.Key(msg, "n"):
		s.editingID = ""

		return modal.PushAt(ctx, dialog.NewTextPrompt(s.theme, "New folder", "Folder name", "", "Create"), dialog.Placement(ctx, 52, 8))
	case reactea.Key(msg, "e"):
		if folder := s.selectedFolder(); folder != nil {
			s.editingID = folder.ID

			return modal.PushAt(ctx, dialog.NewTextPrompt(s.theme, "Rename folder", "Folder name", folder.Name, "Save"), dialog.Placement(ctx, 52, 8))
		}
	case reactea.Key(msg, "d"):
		if folder := s.selectedFolder(); folder != nil {
			s.deletingID = folder.ID

			return modal.PushAt(ctx, dialog.NewConfirm(s.theme, "Delete folder?", "Bookmarks remain in All bookmarks.", "Delete"), dialog.Placement(ctx, 52, 8))
		}
	case reactea.Key(msg, "."):
		return s.load(ctx.Context())
	}

	return nil
}

func (s *FolderIndex) applyChanged(msg changedMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}
	s.notice = "Folder saved"
	if msg.deleted != "" {
		s.folders = slices.DeleteFunc(s.folders, func(folder bookmarkdomain.BookmarkFolder) bool {
			return folder.ID == msg.deleted
		})
		s.notice = "Folder deleted"
		s.clampSelection()

		return nil
	}
	for index := range s.folders {
		if s.folders[index].ID == msg.folder.ID {
			s.folders[index] = msg.folder

			return nil
		}
	}
	s.folders = append(s.folders, msg.folder)
	s.selected = len(s.folders) - 1

	return nil
}

func (s *FolderIndex) selectedFolder() *bookmarkdomain.BookmarkFolder {
	if s.selected < 0 || s.selected >= len(s.folders) {
		return nil
	}

	return &s.folders[s.selected]
}

func (s *FolderIndex) open() tea.Cmd {
	if s.selected < 0 || s.selected >= len(s.folders) {
		return nil
	}

	return navigation.OpenBookmarkFolder(s.folders[s.selected])
}

func (s *FolderIndex) click(ctx *reactea.Ctx, msg tea.MouseClickMsg) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return nil
	}
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	if pageheader.BackAt(x, y) {
		return pageheader.Back()
	}
	index := s.top + (y-pageheader.Height)/folderHeight
	if y < pageheader.Height || index < s.top || index >= len(s.folders) {
		return nil
	}
	s.selected = index
	s.ensureVisible(s.room(ctx.Height()))

	return s.open()
}
