package northapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/domain/session"
)

type officialSpy struct {
	OfficialAPI
	meCalls           int
	notificationCalls int
	unreadCalls       int
	markReadCalls     int
}

func (s *officialSpy) Me(context.Context) (north.User, *north.Response, error) {
	s.meCalls++

	return north.User{ID: "official-user", Handle: "official"}, nil, nil
}

func (s *officialSpy) Notifications(context.Context, north.NotificationTab, string) (north.NotificationPage, *north.Response, error) {
	s.notificationCalls++

	return north.NotificationPage{Items: []north.Notification{{ID: "official-notice"}}}, nil, nil
}

func (s *officialSpy) NotificationUnreadCount(context.Context) (int, *north.Response, error) {
	s.unreadCalls++

	return 2, nil, nil
}

func (s *officialSpy) MarkNotificationsRead(context.Context) (int, *north.Response, error) {
	s.markReadCalls++

	return 2, nil, nil
}

func TestHybridPrefersOfficialAndKeepsBrowserOnlyFeatures(t *testing.T) {
	t.Parallel()

	var webRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		webRequests.Add(1)
		switch request.URL.Path {
		case "/api/notifications":
			writeJSON(t, writer, `{"items":[]}`)
		case "/api/tweets/post-1/conversation":
			writeJSON(t, writer, `{"ancestors":[],"tweet":{"id":"post-1","media":[]},"replies":[]}`)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	web := newTestWebClient(t, server, "session=web")
	official := &officialSpy{}
	client := NewHybrid(official, web, "Firefox · default", nil)
	user, _, err := client.Me(context.Background())
	if err != nil || user.ID != "official-user" {
		t.Fatalf("Me = %#v, %v", user, err)
	}
	if official.meCalls != 1 || webRequests.Load() != 0 {
		t.Fatalf("official calls = %d, web requests = %d", official.meCalls, webRequests.Load())
	}
	page, _, err := client.Notifications(context.Background(), north.NotificationsAll, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "official-notice" {
		t.Fatalf("Notifications = %#v, %v", page, err)
	}
	if official.notificationCalls != 1 || webRequests.Load() != 0 {
		t.Fatalf("notification calls: official = %d, web = %d", official.notificationCalls, webRequests.Load())
	}
	unread, _, err := client.NotificationUnreadCount(context.Background())
	if err != nil || unread != 2 {
		t.Fatalf("NotificationUnreadCount = %d, %v", unread, err)
	}
	marked, _, err := client.MarkNotificationsRead(context.Background())
	if err != nil || marked != 2 {
		t.Fatalf("MarkNotificationsRead = %d, %v", marked, err)
	}
	if official.unreadCalls != 1 || official.markReadCalls != 1 || webRequests.Load() != 0 {
		t.Fatalf("notification state calls: unread = %d, read = %d, web = %d", official.unreadCalls, official.markReadCalls, webRequests.Load())
	}
	conversation, _, err := client.PostConversation(context.Background(), "post-1", "")
	if err != nil || conversation.Post.ID != "post-1" {
		t.Fatalf("PostConversation = %#v, %v", conversation, err)
	}
	if webRequests.Load() != 1 {
		t.Fatalf("browser-only web requests = %d", webRequests.Load())
	}
}

func TestHybridRefreshesAnExpiredBrowserSessionAndRetries(t *testing.T) {
	t.Parallel()

	var staleRequests atomic.Int32
	staleServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		staleRequests.Add(1)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte(`{"error":"UNAUTHORIZED","message":"sign in again"}`))
	}))
	t.Cleanup(staleServer.Close)

	var freshRequests atomic.Int32
	freshServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		freshRequests.Add(1)
		if got := request.Header.Get("Cookie"); got != "session=fresh" {
			t.Errorf("fresh Cookie = %q", got)
		}
		writeJSON(t, writer, `{"items":[{"id":"notice-1","kind":"LIKE"}]}`)
	}))
	t.Cleanup(freshServer.Close)

	stale := newTestWebClient(t, staleServer, "session=stale")
	fresh := newTestWebClient(t, freshServer, "session=fresh")
	var refreshCalls atomic.Int32
	client := NewHybrid(nil, stale, "Firefox · default", func(context.Context) (*Client, error) {
		refreshCalls.Add(1)
		return fresh, nil
	})
	var prompts atomic.Int32
	client.SetCookieRefreshHandler(func(request *session.RefreshRequest) {
		prompts.Add(1)
		request.Complete(request.Refresh(context.Background()))
	})

	page, _, err := client.Notifications(context.Background(), north.NotificationsAll, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "notice-1" {
		t.Fatalf("notifications = %#v", page)
	}
	if staleRequests.Load() != 1 || freshRequests.Load() != 1 || refreshCalls.Load() != 1 || prompts.Load() != 1 {
		t.Fatalf(
			"stale = %d, fresh = %d, refresh = %d, prompts = %d",
			staleRequests.Load(), freshRequests.Load(), refreshCalls.Load(), prompts.Load(),
		)
	}
}

func newTestWebClient(t *testing.T, server *httptest.Server, cookie string) *Client {
	t.Helper()
	client, err := New(
		cookie,
		unofficial.WithBaseURL(server.URL),
		unofficial.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

var _ OfficialAPI = (*officialSpy)(nil)
