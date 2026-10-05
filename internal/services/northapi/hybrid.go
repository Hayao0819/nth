package northapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/domain/conversation"
	"github.com/Hayao0819/nth/internal/domain/notification"
	"github.com/Hayao0819/nth/internal/domain/session"
)

type OfficialAPI interface {
	Me(context.Context) (north.User, *north.Response, error)
	User(context.Context, string) (north.User, *north.Response, error)
	HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error)
	SearchPosts(context.Context, string, north.SearchOptions) (north.PostPage, *north.Response, error)
	Mentions(context.Context, string, string) (north.PostPage, *north.Response, error)
	UserPosts(context.Context, string, string) (north.PostPage, *north.Response, error)
	Post(context.Context, string) (north.Post, *north.Response, error)
	CreatePost(context.Context, north.CreatePostRequest) (north.CreatedPost, *north.Response, error)
	DeletePost(context.Context, string) (bool, *north.Response, error)
	Like(context.Context, string) (north.LikeState, *north.Response, error)
	Unlike(context.Context, string) (north.LikeState, *north.Response, error)
	Repost(context.Context, string) (north.RepostState, *north.Response, error)
	UndoRepost(context.Context, string) (north.RepostState, *north.Response, error)
}

type officialBookmarkAPI interface {
	Bookmarks(context.Context, string) (north.BookmarkPage, *north.Response, error)
}

type officialConversationAPI interface {
	Conversation(context.Context, string, string) (north.Conversation, *north.Response, error)
}

type officialEditorAPI interface {
	EditPost(context.Context, string, north.EditPostRequest) (north.Post, *north.Response, error)
}

type officialMessageAPI interface {
	DMConversations(context.Context, string, bool) (north.DMConversationPage, *north.Response, error)
	DMMessages(context.Context, string, string) (north.DMMessagePage, *north.Response, error)
	MarkDMRead(context.Context, string) (bool, *north.Response, error)
}

type officialTrendAPI interface {
	Trends(context.Context, string) ([]north.Trend, *north.Response, error)
}

type RefreshFunc func(context.Context) (*Client, error)

type Hybrid struct {
	official OfficialAPI
	profile  string
	refresh  RefreshFunc

	webMu sync.RWMutex
	web   *Client

	handlerMu sync.RWMutex
	handler   func(*session.RefreshRequest)

	refreshMu sync.Mutex
	request   *session.RefreshRequest
}

func NewHybrid(official OfficialAPI, web *Client, profile string, refresh RefreshFunc) *Hybrid {
	return &Hybrid{
		official: official,
		web:      web,
		profile:  profile,
		refresh:  refresh,
	}
}

func (c *Hybrid) SetCookieRefreshHandler(handler func(*session.RefreshRequest)) {
	c.handlerMu.Lock()
	c.handler = handler
	c.handlerMu.Unlock()
}

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

func (c *Hybrid) Bookmarks(ctx context.Context, cursor string) (north.PostPage, *north.Response, error) {
	if official, ok := c.official.(officialBookmarkAPI); ok {
		page, response, err := official.Bookmarks(ctx, cursor)

		return north.PostPage{Items: page.Items, NextCursor: page.NextCursor}, response, err
	}

	return withWeb(ctx, c, func(client *Client) (north.PostPage, *north.Response, error) {
		return client.Bookmarks(ctx, cursor)
	})
}

func (c *Hybrid) Post(ctx context.Context, id string) (north.Post, *north.Response, error) {
	if c.official != nil {
		return c.official.Post(ctx, id)
	}

	return withWeb(ctx, c, func(client *Client) (north.Post, *north.Response, error) {
		return client.Post(ctx, id)
	})
}

func (c *Hybrid) PostConversation(ctx context.Context, id, cursor string) (conversation.Page, *north.Response, error) {
	if official, ok := c.official.(officialConversationAPI); ok {
		page, response, err := official.Conversation(ctx, id, cursor)

		return conversation.Page{
			Ancestors:  page.Ancestors,
			Post:       page.Post,
			Replies:    page.Replies,
			NextCursor: page.NextCursor,
		}, response, err
	}

	return withWeb(ctx, c, func(client *Client) (conversation.Page, *north.Response, error) {
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
	if official, ok := c.official.(officialEditorAPI); ok {
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

func (c *Hybrid) Notifications(ctx context.Context, tab north.NotificationTab, cursor string) (north.NotificationPage, *north.Response, error) {
	if official, ok := c.official.(notification.API); ok {
		return official.Notifications(ctx, tab, cursor)
	}

	return withWeb(ctx, c, func(client *Client) (north.NotificationPage, *north.Response, error) {
		return client.Notifications(ctx, tab, cursor)
	})
}

func (c *Hybrid) NotificationUnreadCount(ctx context.Context) (int, *north.Response, error) {
	if official, ok := c.official.(notification.API); ok {
		return official.NotificationUnreadCount(ctx)
	}

	return withWeb(ctx, c, func(client *Client) (int, *north.Response, error) {
		return client.NotificationUnreadCount(ctx)
	})
}

func (c *Hybrid) MarkNotificationsRead(ctx context.Context) (int, *north.Response, error) {
	if official, ok := c.official.(notification.API); ok {
		return official.MarkNotificationsRead(ctx)
	}

	return withWeb(ctx, c, func(client *Client) (int, *north.Response, error) {
		return client.MarkNotificationsRead(ctx)
	})
}

func (c *Hybrid) DMConversations(ctx context.Context, cursor string, requests bool) (north.DMConversationPage, *north.Response, error) {
	if official, ok := c.official.(officialMessageAPI); ok {
		return official.DMConversations(ctx, cursor, requests)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMConversationPage, *north.Response, error) {
		return client.DMConversations(ctx, cursor, requests)
	})
}

func (c *Hybrid) DMMessages(ctx context.Context, conversationID, cursor string) (north.DMMessagePage, *north.Response, error) {
	if official, ok := c.official.(officialMessageAPI); ok {
		return official.DMMessages(ctx, conversationID, cursor)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMMessagePage, *north.Response, error) {
		return client.DMMessages(ctx, conversationID, cursor)
	})
}

func (c *Hybrid) MarkDMRead(ctx context.Context, conversationID string) (*north.Response, error) {
	if official, ok := c.official.(officialMessageAPI); ok {
		_, response, err := official.MarkDMRead(ctx, conversationID)

		return response, err
	}

	_, response, err := withWeb(ctx, c, func(client *Client) (struct{}, *north.Response, error) {
		response, err := client.MarkDMRead(ctx, conversationID)

		return struct{}{}, response, err
	})

	return response, err
}

func (c *Hybrid) Trends(ctx context.Context, cursor string) ([]north.Trend, *north.Response, error) {
	if official, ok := c.official.(officialTrendAPI); ok {
		return official.Trends(ctx, cursor)
	}

	return withWeb(ctx, c, func(client *Client) ([]north.Trend, *north.Response, error) {
		return client.Trends(ctx, cursor)
	})
}

func withWeb[T any](ctx context.Context, client *Hybrid, call func(*Client) (T, *north.Response, error)) (T, *north.Response, error) {
	web := client.currentWeb()
	if web == nil {
		var err error
		web, err = client.refreshWeb(ctx, nil)
		if err != nil {
			var zero T

			return zero, nil, err
		}
	}

	value, response, err := call(web)
	if !cookieExpired(err) {
		return value, response, err
	}

	web, refreshErr := client.refreshWeb(ctx, web)
	if refreshErr != nil {
		var zero T

		return zero, response, refreshErr
	}

	return call(web)
}

func (c *Hybrid) currentWeb() *Client {
	c.webMu.RLock()
	defer c.webMu.RUnlock()

	return c.web
}

func (c *Hybrid) refreshWeb(ctx context.Context, stale *Client) (*Client, error) {
	if current := c.currentWeb(); current != nil && current != stale {
		return current, nil
	}
	if c.refresh == nil {
		return nil, errors.New("browser session is not configured")
	}

	c.refreshMu.Lock()
	request := c.request
	first := request == nil
	if first {
		request = session.NewRefreshRequest(c.profile, func(refreshCtx context.Context) error {
			web, err := c.refresh(refreshCtx)
			if err != nil {
				return err
			}
			c.webMu.Lock()
			c.web = web
			c.webMu.Unlock()

			return nil
		})
		c.request = request
	}
	c.refreshMu.Unlock()

	if first {
		c.handlerMu.RLock()
		handler := c.handler
		c.handlerMu.RUnlock()
		if handler == nil {
			err := request.Refresh(ctx)
			request.Complete(err)
		} else {
			handler(request)
		}
	}

	err := request.Wait(ctx)
	c.refreshMu.Lock()
	if c.request == request {
		if _, complete := request.Result(); complete {
			c.request = nil
		}
	}
	c.refreshMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("refresh browser session: %w", err)
	}
	web := c.currentWeb()
	if web == nil {
		return nil, errors.New("refresh browser session: no client was created")
	}

	return web, nil
}

func cookieExpired(err error) bool {
	if errors.Is(err, unofficial.ErrNotAuthenticated) {
		return true
	}
	var apiError *unofficial.APIError

	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusUnauthorized
}
