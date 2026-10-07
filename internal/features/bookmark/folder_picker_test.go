package bookmark

import (
	"context"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type folderAPI struct {
	folders []north.BookmarkFolder
	members []north.BookmarkFolder
	added   []string
	removed []string
}

func (*folderAPI) ClearBookmarks(context.Context) (int, *north.Response, error) {
	return 0, nil, nil
}

func (a *folderAPI) BookmarkFolders(context.Context, string) ([]north.BookmarkFolder, *north.Response, error) {
	return append([]north.BookmarkFolder(nil), a.folders...), nil, nil
}

func (*folderAPI) CreateBookmarkFolder(context.Context, string) (north.BookmarkFolder, *north.Response, error) {
	return north.BookmarkFolder{}, nil, nil
}

func (*folderAPI) UpdateBookmarkFolder(context.Context, string, string) (north.BookmarkFolder, *north.Response, error) {
	return north.BookmarkFolder{}, nil, nil
}

func (*folderAPI) DeleteBookmarkFolder(context.Context, string) (bool, *north.Response, error) {
	return true, nil, nil
}

func (*folderAPI) BookmarkFolderPosts(context.Context, string, string) (north.BookmarkPage, *north.Response, error) {
	return north.BookmarkPage{}, nil, nil
}

func (a *folderAPI) PostBookmarkFolders(context.Context, string, string) ([]north.BookmarkFolder, *north.Response, error) {
	return append([]north.BookmarkFolder(nil), a.members...), nil, nil
}

func (a *folderAPI) AddBookmarkToFolder(_ context.Context, folderID, postID string) (north.BookmarkState, *north.Response, error) {
	a.added = append(a.added, folderID+":"+postID)

	return north.BookmarkState{Bookmarked: true, OK: true}, nil, nil
}

func (a *folderAPI) RemoveBookmarkFromFolder(_ context.Context, folderID, postID string) (bool, *north.Response, error) {
	a.removed = append(a.removed, folderID+":"+postID)

	return true, nil, nil
}

func TestFolderPickerLoadsAndTogglesMembership(t *testing.T) {
	t.Parallel()

	api := &folderAPI{
		folders: []north.BookmarkFolder{
			{ID: "read", Name: "Read later", Count: 3},
			{ID: "work", Name: "Work", Count: 1},
		},
		members: []north.BookmarkFolder{{ID: "read", Name: "Read later"}},
	}
	picker := newFolderPicker(api, ui.NewTheme(), "post-1")
	program := reactea.New(picker, reactea.WithSize(52, 12))
	program.Start()

	plain := testkit.Plain(program)
	if !strings.Contains(plain, "[x] Read later") || !strings.Contains(plain, "[ ] Work") {
		t.Fatalf("folder picker did not show membership:\n%s", plain)
	}
	testkit.SendKeys(program, "space", "down", "space")
	if got := strings.Join(api.removed, ","); got != "read:post-1" {
		t.Fatalf("removed = %q", got)
	}
	if got := strings.Join(api.added, ","); got != "work:post-1" {
		t.Fatalf("added = %q", got)
	}
	plain = testkit.Plain(program)
	if !strings.Contains(plain, "[ ] Read later") || !strings.Contains(plain, "[x] Work") {
		t.Fatalf("folder picker did not update membership:\n%s", plain)
	}
}
