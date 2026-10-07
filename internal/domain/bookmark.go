package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type BookmarkAPI interface {
	Bookmarks(context.Context, string) (north.PostPage, *north.Response, error)
	Bookmark(context.Context, string) (north.BookmarkState, *north.Response, error)
	Unbookmark(context.Context, string) (bool, *north.Response, error)
}

type BookmarkFolder = north.BookmarkFolder

type BookmarkFolderAPI interface {
	ClearBookmarks(context.Context) (int, *north.Response, error)
	BookmarkFolders(context.Context, string) ([]BookmarkFolder, *north.Response, error)
	CreateBookmarkFolder(context.Context, string) (BookmarkFolder, *north.Response, error)
	UpdateBookmarkFolder(context.Context, string, string) (BookmarkFolder, *north.Response, error)
	DeleteBookmarkFolder(context.Context, string) (bool, *north.Response, error)
	BookmarkFolderPosts(context.Context, string, string) (north.BookmarkPage, *north.Response, error)
	PostBookmarkFolders(context.Context, string, string) ([]BookmarkFolder, *north.Response, error)
	AddBookmarkToFolder(context.Context, string, string) (north.BookmarkState, *north.Response, error)
	RemoveBookmarkFromFolder(context.Context, string, string) (bool, *north.Response, error)
}

var _ BookmarkFolderAPI = (*north.Client)(nil)
