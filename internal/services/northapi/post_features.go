package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
)

type officialPostActivityAPI interface {
	Quotes(context.Context, string, string) (north.PostPage, *north.Response, error)
	PostLikes(context.Context, string, string) (north.UserPage, *north.Response, error)
	PostReposts(context.Context, string, string) (north.UserPage, *north.Response, error)
	PostEditHistory(context.Context, string) (north.EditHistory, *north.Response, error)
}

type officialPollAPI interface {
	VotePoll(context.Context, string, string) (north.Post, *north.Response, error)
}

func (c *Hybrid) SupportsPostActivity() bool {
	_, supported := c.official.(officialPostActivityAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsPostEditHistory() bool {
	_, supported := c.official.(officialPostActivityAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsPolls() bool {
	_, supported := c.official.(officialPollAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsThreads() bool {
	_, supported := c.official.(officialThreadAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) Quotes(ctx context.Context, id, cursor string) (north.PostPage, *north.Response, error) {
	api, ok := c.official.(officialPostActivityAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.PostPage{}, nil, errors.New("quotes are not available with the configured credentials")
	}

	return api.Quotes(ctx, id, cursor)
}

func (c *Hybrid) PostLikes(ctx context.Context, id, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.official.(officialPostActivityAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.UserPage{}, nil, errors.New("post likes are not available with the configured credentials")
	}

	return api.PostLikes(ctx, id, cursor)
}

func (c *Hybrid) PostReposts(ctx context.Context, id, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.official.(officialPostActivityAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.UserPage{}, nil, errors.New("post reposts are not available with the configured credentials")
	}

	return api.PostReposts(ctx, id, cursor)
}

func (c *Hybrid) PostEditHistory(ctx context.Context, id string) (north.EditHistory, *north.Response, error) {
	if api, ok := c.official.(officialPostActivityAPI); ok && c.officialSupportsScopedAPI() {
		return api.PostEditHistory(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.EditHistory, *north.Response, error) {
		return client.PostEditHistory(ctx, id)
	})
}

func (c *Hybrid) VotePoll(ctx context.Context, id, optionID string) (north.Post, *north.Response, error) {
	if api, ok := c.official.(officialPollAPI); ok && c.officialSupportsScopedAPI() {
		return api.VotePoll(ctx, id, optionID)
	}

	return withWeb(ctx, c, func(client *Client) (north.Post, *north.Response, error) {
		return client.VotePoll(ctx, id, optionID)
	})
}

func (c *Client) PostEditHistory(ctx context.Context, id string) (north.EditHistory, *north.Response, error) {
	history, response, err := c.web.PostEditHistory(ctx, id)
	result := north.EditHistory{
		Post:     history.Post.PublicPost(),
		Versions: make([]north.PostVersion, len(history.Versions)),
	}
	for index, version := range history.Versions {
		result.Versions[index] = north.PostVersion{
			ID:        version.ID,
			Text:      version.Text,
			Media:     publicMedia(version.Media),
			CreatedAt: version.CreatedAt,
			Current:   version.Current,
		}
	}

	return result, publicResponse(response), err
}

func (c *Client) VotePoll(ctx context.Context, id, optionID string) (north.Post, *north.Response, error) {
	_, response, err := c.web.VotePoll(ctx, id, optionID)
	if err != nil {
		return north.Post{}, publicResponse(response), err
	}
	post, fetched, err := c.Post(ctx, id)
	if fetched != nil {
		return post, fetched, err
	}

	return post, publicResponse(response), err
}
