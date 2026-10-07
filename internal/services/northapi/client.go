// Package northapi adapts north's private web API to the interface used by nth.
package northapi

import (
	"context"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/domain"
)

// Client uses a north.rip web session for all nth operations.
type Client struct {
	web *unofficial.Client
}

// New creates a client from the value of a browser Cookie header.
func New(cookie string, options ...unofficial.Option) (*Client, error) {
	client, err := unofficial.NewClient(cookie, options...)
	if err != nil {
		return nil, err
	}

	return &Client{web: client}, nil
}

func (c *Client) Me(ctx context.Context) (north.User, *north.Response, error) {
	user, response, err := c.web.Me(ctx)

	return user.User, publicResponse(response), err
}

func (c *Client) User(ctx context.Context, handle string) (north.User, *north.Response, error) {
	user, response, err := c.web.User(ctx, handle)

	return user.User, publicResponse(response), err
}

func (c *Client) HomeTimeline(ctx context.Context, options north.TimelineOptions) (north.PostPage, *north.Response, error) {
	return publicPostPageResult(c.web.HomeTimeline(ctx, options))
}

func (c *Client) SearchPosts(ctx context.Context, query string, options north.SearchOptions) (north.PostPage, *north.Response, error) {
	return publicPostPageResult(c.web.SearchPosts(ctx, query, options))
}

func (c *Client) Mentions(ctx context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	return publicPostPageResult(c.web.Mentions(ctx, handle, cursor))
}

func (c *Client) UserPosts(ctx context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	return publicPostPageResult(c.web.UserPosts(ctx, handle, cursor))
}

func (c *Client) Trends(ctx context.Context, _ string) ([]north.Trend, *north.Response, error) {
	list, response, err := c.web.Trends(ctx)

	return publicTrends(list.Items), publicResponse(response), err
}

func (c *Client) Bookmarks(ctx context.Context, cursor string) (north.PostPage, *north.Response, error) {
	return publicPostPageResult(c.web.Bookmarks(ctx, cursor))
}

func (c *Client) Bookmark(ctx context.Context, id string) (north.BookmarkState, *north.Response, error) {
	response, err := c.web.Bookmark(ctx, id)

	return north.BookmarkState{Bookmarked: err == nil, OK: err == nil}, publicResponse(response), err
}

func (c *Client) Unbookmark(ctx context.Context, id string) (bool, *north.Response, error) {
	response, err := c.web.Unbookmark(ctx, id)

	return err == nil, publicResponse(response), err
}

func (c *Client) Post(ctx context.Context, id string) (north.Post, *north.Response, error) {
	post, response, err := c.web.Post(ctx, id)

	return post.PublicPost(), publicResponse(response), err
}

func (c *Client) PostConversation(ctx context.Context, id, cursor string) (domain.ConversationPage, *north.Response, error) {
	page, response, err := c.web.PostConversation(ctx, id, cursor)

	return domain.ConversationPage{
		Ancestors:    publicPosts(page.Ancestors),
		Post:         page.Tweet.PublicPost(),
		Replies:      publicPosts(page.Replies),
		NextCursor:   page.NextCursor,
		ReaderRootID: page.ReaderRootID,
	}, publicResponse(response), err
}

func (c *Client) CreatePost(ctx context.Context, request north.CreatePostRequest) (north.CreatedPost, *north.Response, error) {
	webRequest := unofficial.CreatePostRequest{
		Text:     request.Text,
		QuotedID: request.QuotePostID,
	}
	if request.Poll != nil {
		webRequest.Poll = &unofficial.CreatePollRequest{
			Options:         append([]string(nil), request.Poll.Options...),
			DurationMinutes: request.Poll.DurationMinutes,
		}
	}
	if request.Media != nil {
		webRequest.MediaIDs = request.Media.MediaIDs
	}
	if request.Reply != nil {
		webRequest.InReplyToID = request.Reply.InReplyToPostID
	}
	post, response, err := c.web.CreatePost(ctx, webRequest)

	return post, publicResponse(response), err
}

func (c *Client) DeletePost(ctx context.Context, id string) (bool, *north.Response, error) {
	deleted, response, err := c.web.DeletePost(ctx, id)

	return deleted, publicResponse(response), err
}

func (c *Client) EditablePost(ctx context.Context, id string) (north.Post, bool, *north.Response, error) {
	post, response, err := c.web.Post(ctx, id)

	return post.PublicPost(), post.EditEligible, publicResponse(response), err
}

func (c *Client) EditPost(ctx context.Context, id, text string, mediaIDs []string) (*north.Response, error) {
	response, err := c.web.EditPost(ctx, id, unofficial.EditPostRequest{Text: text, MediaIDs: mediaIDs})

	return publicResponse(response), err
}

func (c *Client) Like(ctx context.Context, id string) (north.LikeState, *north.Response, error) {
	state, response, err := c.web.Like(ctx, id)

	return state, publicResponse(response), err
}

func (c *Client) Unlike(ctx context.Context, id string) (north.LikeState, *north.Response, error) {
	state, response, err := c.web.Unlike(ctx, id)

	return state, publicResponse(response), err
}

func (c *Client) Repost(ctx context.Context, id string) (north.RepostState, *north.Response, error) {
	state, response, err := c.web.Repost(ctx, id)

	return state, publicResponse(response), err
}

func (c *Client) UndoRepost(ctx context.Context, id string) (north.RepostState, *north.Response, error) {
	state, response, err := c.web.UndoRepost(ctx, id)

	return state, publicResponse(response), err
}

func (c *Client) Notifications(ctx context.Context, tab north.NotificationTab, cursor string) (north.NotificationPage, *north.Response, error) {
	page, response, err := c.web.Notifications(ctx, unofficial.NotificationTab(tab), cursor)
	items := make([]north.Notification, len(page.Items))
	for index, item := range page.Items {
		items[index] = north.Notification{
			ID:          item.ID,
			Kind:        item.Kind,
			Read:        item.Read,
			Actors:      publicUsers(item.Actors),
			ActorCount:  item.ActorCount,
			GroupCount:  item.GroupCount,
			TargetCount: item.TargetCount,
			CreatedAt:   item.CreatedAt,
		}
		if item.Post != nil {
			post := item.Post.PublicPost()
			items[index].Post = &post
		}
	}

	return north.NotificationPage{Items: items, NextCursor: page.NextCursor}, publicResponse(response), err
}

func (c *Client) NotificationUnreadCount(ctx context.Context) (int, *north.Response, error) {
	count, response, err := c.web.NotificationUnreadCount(ctx)

	return count, publicResponse(response), err
}

func (c *Client) MarkNotificationsRead(ctx context.Context) (int, *north.Response, error) {
	response, err := c.web.MarkNotificationsRead(ctx)

	return 0, publicResponse(response), err
}
