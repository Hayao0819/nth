package bookmark

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	bookmarkdomain "github.com/Hayao0819/nth/internal/domain"
)

type loadedMsg struct {
	target   *FolderIndex
	folders  []bookmarkdomain.BookmarkFolder
	response *north.Response
	err      error
}

func (m loadedMsg) Response() *north.Response { return m.response }

type changedMsg struct {
	target   *FolderIndex
	folder   bookmarkdomain.BookmarkFolder
	deleted  string
	response *north.Response
	err      error
}

func (m changedMsg) Response() *north.Response { return m.response }

func (s *FolderIndex) load(ctx context.Context) tea.Cmd {
	if s.api == nil || s.loading {
		return nil
	}
	s.loading = true
	s.err = nil

	return func() tea.Msg {
		folders, response, err := s.api.BookmarkFolders(ctx, "")

		return loadedMsg{target: s, folders: folders, response: response, err: err}
	}
}

func (s *FolderIndex) save(ctx context.Context, id, name string) tea.Cmd {
	if s.api == nil || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Saving folder…"

	return func() tea.Msg {
		var folder bookmarkdomain.BookmarkFolder
		var response *north.Response
		var err error
		if id == "" {
			folder, response, err = s.api.CreateBookmarkFolder(ctx, name)
		} else {
			folder, response, err = s.api.UpdateBookmarkFolder(ctx, id, name)
		}

		return changedMsg{target: s, folder: folder, response: response, err: err}
	}
}

func (s *FolderIndex) delete(ctx context.Context, id string) tea.Cmd {
	if s.api == nil || s.acting || id == "" {
		return nil
	}
	s.acting = true
	s.notice = "Deleting folder…"

	return func() tea.Msg {
		deleted, response, err := s.api.DeleteBookmarkFolder(ctx, id)
		if err == nil && !deleted {
			err = errNotDeleted
		}

		return changedMsg{target: s, deleted: id, response: response, err: err}
	}
}
