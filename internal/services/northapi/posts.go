package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialConversationAPI interface {
	Conversation(context.Context, string, string) (north.Conversation, *north.Response, error)
}

type officialThreadAPI interface {
	CreateThread(context.Context, []north.ThreadItem) ([]north.Post, *north.Response, error)
}

type officialEditorAPI interface {
	EditPost(context.Context, string, north.EditPostRequest) (north.Post, *north.Response, error)
}

func (c *Hybrid) Post(ctx context.Context, id string) (north.Post, *north.Response, error) {
	if c.official != nil {
		return c.official.Post(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.Post, *north.Response, error) {
		return client.Post(ctx, id)
	})
}

func (c *Hybrid) PostConversation(ctx context.Context, id, cursor string) (domain.ConversationPage, *north.Response, error) {
	if official, ok := c.official.(officialConversationAPI); ok {
		page, response, err := official.Conversation(ctx, id, cursor)

		return domain.ConversationPage{
			Ancestors:  page.Ancestors,
			Post:       page.Post,
			Replies:    page.Replies,
			NextCursor: page.NextCursor,
		}, response, err
	}

	return withWeb(ctx, c, func(client *Client) (domain.ConversationPage, *north.Response, error) {
		return client.PostConversation(ctx, id, cursor)
	})
}

func (c *Hybrid) CreatePost(ctx context.Context, request north.CreatePostRequest) (north.CreatedPost, *north.Response, error) {
	if c.official != nil {
		return c.official.CreatePost(ctx, request)
	}

	return withWeb(ctx, c, func(client *Client) (north.CreatedPost, *north.Response, error) {
		return client.CreatePost(ctx, request)
	})
}

func (c *Hybrid) CreateThread(ctx context.Context, items []north.ThreadItem) ([]north.Post, *north.Response, error) {
	official, ok := c.official.(officialThreadAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, nil, errors.New("thread creation is not available with the configured credentials")
	}

	return official.CreateThread(ctx, items)
}

func (c *Hybrid) DeletePost(ctx context.Context, id string) (bool, *north.Response, error) {
	if c.official != nil {
		return c.official.DeletePost(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (bool, *north.Response, error) {
		return client.DeletePost(ctx, id)
	})
}

func (c *Hybrid) Like(ctx context.Context, id string) (north.LikeState, *north.Response, error) {
	if c.official != nil {
		return c.official.Like(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.LikeState, *north.Response, error) {
		return client.Like(ctx, id)
	})
}

func (c *Hybrid) Unlike(ctx context.Context, id string) (north.LikeState, *north.Response, error) {
	if c.official != nil {
		return c.official.Unlike(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.LikeState, *north.Response, error) {
		return client.Unlike(ctx, id)
	})
}

func (c *Hybrid) Repost(ctx context.Context, id string) (north.RepostState, *north.Response, error) {
	if c.official != nil {
		return c.official.Repost(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.RepostState, *north.Response, error) {
		return client.Repost(ctx, id)
	})
}

func (c *Hybrid) UndoRepost(ctx context.Context, id string) (north.RepostState, *north.Response, error) {
	if c.official != nil {
		return c.official.UndoRepost(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.RepostState, *north.Response, error) {
		return client.UndoRepost(ctx, id)
	})
}

func (c *Hybrid) EditablePost(ctx context.Context, id string) (north.Post, bool, *north.Response, error) {
	if c.official != nil {
		post, response, err := c.official.Post(ctx, id)

		return post, post.EditEligible, response, err
	}

	type result struct {
		post     north.Post
		eligible bool
	}
	value, response, err := withWeb(ctx, c, func(client *Client) (result, *north.Response, error) {
		post, eligible, response, err := client.EditablePost(ctx, id)

		return result{post: post, eligible: eligible}, response, err
	})

	return value.post, value.eligible, response, err
}

func (c *Hybrid) EditPost(ctx context.Context, id, text string, mediaIDs []string) (*north.Response, error) {
	if official, ok := c.official.(officialEditorAPI); ok && c.officialSupportsScopedAPI() {
		requestText := text
		requestMediaIDs := append([]string(nil), mediaIDs...)
		_, response, err := official.EditPost(ctx, id, north.EditPostRequest{
			Text:     &requestText,
			MediaIDs: &requestMediaIDs,
		})

		return response, err
	}

	_, response, err := withWeb(ctx, c, func(client *Client) (struct{}, *north.Response, error) {
		response, err := client.EditPost(ctx, id, text, mediaIDs)

		return struct{}{}, response, err
	})

	return response, err
}
