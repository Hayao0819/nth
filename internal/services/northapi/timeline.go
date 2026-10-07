package northapi

import (
	"context"

	"github.com/Hayao0819/go-north"
)

func (c *Hybrid) Me(ctx context.Context) (north.User, *north.Response, error) {
	if c.official != nil {
		return c.official.Me(ctx)
	}

	return withWeb(ctx, c, func(client *Client) (north.User, *north.Response, error) {
		return client.Me(ctx)
	})
}

func (c *Hybrid) User(ctx context.Context, handle string) (north.User, *north.Response, error) {
	if c.official != nil {
		return c.official.User(ctx, handle)
	}

	return withWeb(ctx, c, func(client *Client) (north.User, *north.Response, error) {
		return client.User(ctx, handle)
	})
}

func (c *Hybrid) HomeTimeline(ctx context.Context, options north.TimelineOptions) (north.PostPage, *north.Response, error) {
	if c.official != nil {
		return c.official.HomeTimeline(ctx, options)
	}

	return withWeb(ctx, c, func(client *Client) (north.PostPage, *north.Response, error) {
		return client.HomeTimeline(ctx, options)
	})
}

func (c *Hybrid) SearchPosts(ctx context.Context, query string, options north.SearchOptions) (north.PostPage, *north.Response, error) {
	if c.official != nil {
		return c.official.SearchPosts(ctx, query, options)
	}

	return withWeb(ctx, c, func(client *Client) (north.PostPage, *north.Response, error) {
		return client.SearchPosts(ctx, query, options)
	})
}

func (c *Hybrid) Mentions(ctx context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	if c.official != nil {
		return c.official.Mentions(ctx, handle, cursor)
	}

	return withWeb(ctx, c, func(client *Client) (north.PostPage, *north.Response, error) {
		return client.Mentions(ctx, handle, cursor)
	})
}

func (c *Hybrid) UserPosts(ctx context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	if c.official != nil {
		return c.official.UserPosts(ctx, handle, cursor)
	}

	return withWeb(ctx, c, func(client *Client) (north.PostPage, *north.Response, error) {
		return client.UserPosts(ctx, handle, cursor)
	})
}
