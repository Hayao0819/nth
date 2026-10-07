package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialBookmarkAPI interface {
	Bookmarks(context.Context, string) (north.BookmarkPage, *north.Response, error)
	Bookmark(context.Context, string) (north.BookmarkState, *north.Response, error)
	Unbookmark(context.Context, string) (bool, *north.Response, error)
}

type officialBookmarkFolderAPI interface {
	domain.BookmarkFolderAPI
}

func (c *Hybrid) Bookmarks(ctx context.Context, cursor string) (north.PostPage, *north.Response, error) {
	if official, ok := c.official.(officialBookmarkAPI); ok && c.officialSupportsScopedAPI() {
		page, response, err := official.Bookmarks(ctx, cursor)

		return north.PostPage{Items: page.Items, NextCursor: page.NextCursor}, response, err
	}

	return withWeb(ctx, c, func(client *Client) (north.PostPage, *north.Response, error) {
		return client.Bookmarks(ctx, cursor)
	})
}

func (c *Hybrid) Bookmark(ctx context.Context, id string) (north.BookmarkState, *north.Response, error) {
	if official, ok := c.official.(officialBookmarkAPI); ok && c.officialSupportsScopedAPI() {
		return official.Bookmark(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.BookmarkState, *north.Response, error) {
		return client.Bookmark(ctx, id)
	})
}

func (c *Hybrid) Unbookmark(ctx context.Context, id string) (bool, *north.Response, error) {
	if official, ok := c.official.(officialBookmarkAPI); ok && c.officialSupportsScopedAPI() {
		return official.Unbookmark(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (bool, *north.Response, error) {
		return client.Unbookmark(ctx, id)
	})
}

func (c *Hybrid) ClearBookmarks(ctx context.Context) (int, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return 0, nil, errors.New("bookmark management is not available with the configured credentials")
	}

	return official.ClearBookmarks(ctx)
}

func (c *Hybrid) BookmarkFolders(ctx context.Context, cursor string) ([]north.BookmarkFolder, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.BookmarkFolders(ctx, cursor)
}

func (c *Hybrid) CreateBookmarkFolder(ctx context.Context, name string) (north.BookmarkFolder, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.BookmarkFolder{}, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.CreateBookmarkFolder(ctx, name)
}

func (c *Hybrid) UpdateBookmarkFolder(ctx context.Context, id, name string) (north.BookmarkFolder, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.BookmarkFolder{}, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.UpdateBookmarkFolder(ctx, id, name)
}

func (c *Hybrid) DeleteBookmarkFolder(ctx context.Context, id string) (bool, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.DeleteBookmarkFolder(ctx, id)
}

func (c *Hybrid) BookmarkFolderPosts(ctx context.Context, id, cursor string) (north.BookmarkPage, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.BookmarkPage{}, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.BookmarkFolderPosts(ctx, id, cursor)
}

func (c *Hybrid) PostBookmarkFolders(ctx context.Context, postID, cursor string) ([]north.BookmarkFolder, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.PostBookmarkFolders(ctx, postID, cursor)
}

func (c *Hybrid) AddBookmarkToFolder(ctx context.Context, folderID, postID string) (north.BookmarkState, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.BookmarkState{}, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.AddBookmarkToFolder(ctx, folderID, postID)
}

func (c *Hybrid) RemoveBookmarkFromFolder(ctx context.Context, folderID, postID string) (bool, *north.Response, error) {
	official, ok := c.official.(officialBookmarkFolderAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("bookmark folders are not available with the configured credentials")
	}

	return official.RemoveBookmarkFromFolder(ctx, folderID, postID)
}
