package bookmark

import (
	"context"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type folderIndexAPI struct {
	folders []north.BookmarkFolder
	created []string
	renamed []string
	deleted []string
}

func (*folderIndexAPI) ClearBookmarks(context.Context) (int, *north.Response, error) {
	return 0, nil, nil
}

func (a *folderIndexAPI) BookmarkFolders(context.Context, string) ([]north.BookmarkFolder, *north.Response, error) {
	return append([]north.BookmarkFolder(nil), a.folders...), nil, nil
}

func (a *folderIndexAPI) CreateBookmarkFolder(_ context.Context, name string) (north.BookmarkFolder, *north.Response, error) {
	a.created = append(a.created, name)

	return north.BookmarkFolder{ID: "new", Name: name}, nil, nil
}

func (a *folderIndexAPI) UpdateBookmarkFolder(_ context.Context, id, name string) (north.BookmarkFolder, *north.Response, error) {
	a.renamed = append(a.renamed, id+":"+name)

	return north.BookmarkFolder{ID: id, Name: name}, nil, nil
}

func (a *folderIndexAPI) DeleteBookmarkFolder(_ context.Context, id string) (bool, *north.Response, error) {
	a.deleted = append(a.deleted, id)

	return true, nil, nil
}

func (*folderIndexAPI) BookmarkFolderPosts(context.Context, string, string) (north.BookmarkPage, *north.Response, error) {
	return north.BookmarkPage{}, nil, nil
}

func (*folderIndexAPI) PostBookmarkFolders(context.Context, string, string) ([]north.BookmarkFolder, *north.Response, error) {
	return nil, nil, nil
}

func (*folderIndexAPI) AddBookmarkToFolder(context.Context, string, string) (north.BookmarkState, *north.Response, error) {
	return north.BookmarkState{}, nil, nil
}

func (*folderIndexAPI) RemoveBookmarkFromFolder(context.Context, string, string) (bool, *north.Response, error) {
	return true, nil, nil
}

func TestFolderListLoadsAndChangesFolders(t *testing.T) {
	t.Parallel()

	api := &folderIndexAPI{folders: []north.BookmarkFolder{
		{ID: "read", Name: "Read later", Count: 4},
		{ID: "work", Name: "Work", Count: 2},
	}}
	screen := NewFolderIndex(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(60, 16))
	program.Start()

	if plain := testkit.Plain(program); !strings.Contains(plain, "Read later") || !strings.Contains(plain, "4 bookmarks") {
		t.Fatalf("folder list is incomplete:\n%s", plain)
	}
	message, ok := screen.open()().(navigation.OpenBookmarkFolderMsg)
	if !ok || message.Folder.ID != "read" {
		t.Fatalf("open = %#v", message)
	}

	command := screen.save(context.Background(), "", "Personal")
	screen.Update(program.Ctx(), command())
	if got := strings.Join(api.created, ","); got != "Personal" || len(screen.folders) != 3 {
		t.Fatalf("created = %q, folders = %#v", got, screen.folders)
	}
	command = screen.save(context.Background(), "new", "Private")
	screen.Update(program.Ctx(), command())
	if got := strings.Join(api.renamed, ","); got != "new:Private" || screen.folders[2].Name != "Private" {
		t.Fatalf("renamed = %q, folders = %#v", got, screen.folders)
	}
	command = screen.delete(context.Background(), "new")
	screen.Update(program.Ctx(), command())
	if got := strings.Join(api.deleted, ","); got != "new" || len(screen.folders) != 2 {
		t.Fatalf("deleted = %q, folders = %#v", got, screen.folders)
	}
}
