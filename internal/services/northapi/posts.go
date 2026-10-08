package northapi

import (
	"context"
	"errors"
	"net/http"
	"slices"

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

type officialConditionalEditorAPI interface {
	EditPostIfMatch(context.Context, string, north.EditPostRequest, string) (north.Post, *north.Response, error)
}

var errPostEditConflict = errors.New("post changed in another client; latest version loaded and draft kept")

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

func (c *Hybrid) EditablePost(ctx context.Context, id string) (north.Post, bool, string, *north.Response, error) {
	if c.official != nil {
		post, response, err := c.official.Post(ctx, id)

		return post, post.EditEligible, responseETag(response), response, err
	}

	type result struct {
		post     north.Post
		eligible bool
	}
	value, response, err := withWeb(ctx, c, func(client *Client) (result, *north.Response, error) {
		post, eligible, response, err := client.EditablePost(ctx, id)

		return result{post: post, eligible: eligible}, response, err
	})

	return value.post, value.eligible, "", response, err
}

func (c *Hybrid) EditPost(ctx context.Context, edit domain.PostEdit) (north.Post, *north.Response, error) {
	if official, ok := c.official.(officialEditorAPI); ok && c.officialSupportsScopedAPI() {
		requestText := edit.Text
		requestMediaIDs := append([]string{}, edit.MediaIDs...)
		request := north.EditPostRequest{
			Text:     &requestText,
			MediaIDs: &requestMediaIDs,
		}
		if conditional, supported := c.official.(officialConditionalEditorAPI); supported && edit.ETag != "" {
			post, response, err := conditional.EditPostIfMatch(ctx, edit.ID, request, edit.ETag)
			if !editPreconditionFailed(err) {
				return post, response, err
			}

			return c.retryPostEdit(ctx, conditional, edit, request, response, err)
		}

		return official.EditPost(ctx, edit.ID, request)
	}

	_, response, err := withWeb(ctx, c, func(client *Client) (struct{}, *north.Response, error) {
		response, err := client.EditPost(ctx, edit.ID, edit.Text, edit.MediaIDs)

		return struct{}{}, response, err
	})

	return north.Post{}, response, err
}

func (c *Hybrid) retryPostEdit(
	ctx context.Context,
	client officialConditionalEditorAPI,
	edit domain.PostEdit,
	request north.EditPostRequest,
	response *north.Response,
	preconditionErr error,
) (north.Post, *north.Response, error) {
	if edit.Base.ID == "" {
		return north.Post{}, response, preconditionErr
	}

	latest, latestResponse, err := c.official.Post(ctx, edit.ID)
	if err != nil {
		return north.Post{}, response, preconditionErr
	}
	etag := responseETag(latestResponse)
	if !sameEditablePost(edit.Base, latest) {
		return latest, responseWithETag(response, etag), errPostEditConflict
	}
	if etag == "" {
		return latest, response, preconditionErr
	}

	return client.EditPostIfMatch(ctx, edit.ID, request, etag)
}

func editPreconditionFailed(err error) bool {
	var apiError *north.APIError

	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusPreconditionFailed
}

func sameEditablePost(base, latest north.Post) bool {
	basePost := base.DisplayPost()
	latestPost := latest.DisplayPost()
	if basePost == nil || latestPost == nil || basePost.ID != latestPost.ID || basePost.Text != latestPost.Text {
		return false
	}

	return slices.Equal(editMediaIDs(basePost.Media), editMediaIDs(latestPost.Media))
}

func editMediaIDs(media []north.Media) []string {
	ids := make([]string, 0, len(media))
	for _, item := range media {
		if item.ID != "" {
			ids = append(ids, item.ID)
		}
	}

	return ids
}

func responseETag(response *north.Response) string {
	if response == nil {
		return ""
	}

	return response.Header.Get("ETag")
}

func responseWithETag(response *north.Response, etag string) *north.Response {
	if response == nil {
		response = &north.Response{StatusCode: http.StatusPreconditionFailed}
	} else {
		clone := *response
		response = &clone
	}
	response.Header = response.Header.Clone()
	if response.Header == nil {
		response.Header = make(http.Header)
	}
	response.Header.Set("ETag", etag)

	return response
}
