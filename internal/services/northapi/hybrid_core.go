package northapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/domain"
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

type RefreshFunc func(context.Context) (*Client, error)

type Hybrid struct {
	official OfficialAPI
	profile  string
	refresh  RefreshFunc

	webMu sync.RWMutex
	web   *Client

	handlerMu sync.RWMutex
	handler   func(*domain.RefreshRequest)

	refreshMu sync.Mutex
	request   *domain.RefreshRequest
}

func NewHybrid(official OfficialAPI, web *Client, profile string, refresh RefreshFunc) *Hybrid {
	return &Hybrid{official: official, web: web, profile: profile, refresh: refresh}
}

func (c *Hybrid) SetCookieRefreshHandler(handler func(*domain.RefreshRequest)) {
	c.handlerMu.Lock()
	c.handler = handler
	c.handlerMu.Unlock()
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
		request = domain.NewRefreshRequest(c.profile, func(refreshCtx context.Context) error {
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
