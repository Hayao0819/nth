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
	meCalls             int
	postCalls           int
	bookmarkCalls       int
	conversationCalls   int
	editCalls           int
	notificationCalls   int
	unreadCalls         int
	markReadCalls       int
	dmConversationCalls int
	dmMessageCalls      int
	markDMReadCalls     int
	trendCalls          int
	editRequest         north.EditPostRequest
}

func (s *officialSpy) Post(context.Context, string) (north.Post, *north.Response, error) {
	s.postCalls++

	return north.Post{ID: "official-post", EditEligible: true}, nil, nil
}

func (s *officialSpy) Bookmarks(context.Context, string) (north.BookmarkPage, *north.Response, error) {
	s.bookmarkCalls++

	return north.BookmarkPage{Items: []north.Post{{ID: "official-bookmark"}}}, nil, nil
}

func (s *officialSpy) Conversation(context.Context, string, string) (north.Conversation, *north.Response, error) {
	s.conversationCalls++

	return north.Conversation{Post: north.Post{ID: "official-conversation"}}, nil, nil
}

func (s *officialSpy) EditPost(_ context.Context, _ string, request north.EditPostRequest) (north.Post, *north.Response, error) {
	s.editCalls++
	s.editRequest = request

	return north.Post{ID: "official-post"}, nil, nil
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

func (s *officialSpy) DMConversations(context.Context, string, bool) (north.DMConversationPage, *north.Response, error) {
	s.dmConversationCalls++

	return north.DMConversationPage{Items: []north.DMConversation{{ID: "official-dm"}}}, nil, nil
}

func (s *officialSpy) DMMessages(context.Context, string, string) (north.DMMessagePage, *north.Response, error) {
	s.dmMessageCalls++

	return north.DMMessagePage{Items: []north.DMMessage{{ID: "official-message"}}}, nil, nil
}

func (s *officialSpy) MarkDMRead(context.Context, string) (bool, *north.Response, error) {
	s.markDMReadCalls++

	return true, nil, nil
}

func (s *officialSpy) Trends(context.Context, string) ([]north.Trend, *north.Response, error) {
	s.trendCalls++

	return []north.Trend{{Tag: "official-trend"}}, nil, nil
}

func TestHybridPrefersOfficialForSupportedFeatures(t *testing.T) {
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
	bookmarks, _, err := client.Bookmarks(context.Background(), "")
	if err != nil || len(bookmarks.Items) != 1 || bookmarks.Items[0].ID != "official-bookmark" {
		t.Fatalf("Bookmarks = %#v, %v", bookmarks, err)
	}
	conversation, _, err := client.PostConversation(context.Background(), "post-1", "")
	if err != nil || conversation.Post.ID != "official-conversation" {
		t.Fatalf("PostConversation = %#v, %v", conversation, err)
	}
	post, eligible, _, err := client.EditablePost(context.Background(), "post-1")
	if err != nil || !eligible || post.ID != "official-post" {
		t.Fatalf("EditablePost = %#v, %t, %v", post, eligible, err)
	}
	if _, err := client.EditPost(context.Background(), "post-1", "updated", []string{"media-1"}); err != nil {
		t.Fatalf("EditPost: %v", err)
	}
	if official.editRequest.Text == nil || *official.editRequest.Text != "updated" || official.editRequest.MediaIDs == nil || len(*official.editRequest.MediaIDs) != 1 || (*official.editRequest.MediaIDs)[0] != "media-1" {
		t.Fatalf("edit request = %#v", official.editRequest)
	}
	dmConversations, _, err := client.DMConversations(context.Background(), "", false)
	if err != nil || len(dmConversations.Items) != 1 || dmConversations.Items[0].ID != "official-dm" {
		t.Fatalf("DMConversations = %#v, %v", dmConversations, err)
	}
	dmMessages, _, err := client.DMMessages(context.Background(), "official-dm", "")
	if err != nil || len(dmMessages.Items) != 1 || dmMessages.Items[0].ID != "official-message" {
		t.Fatalf("DMMessages = %#v, %v", dmMessages, err)
	}
	if _, err := client.MarkDMRead(context.Background(), "official-dm"); err != nil {
		t.Fatalf("MarkDMRead: %v", err)
	}
	trends, _, err := client.Trends(context.Background(), "")
	if err != nil || len(trends) != 1 || trends[0].Tag != "official-trend" {
		t.Fatalf("Trends = %#v, %v", trends, err)
	}
	if official.bookmarkCalls != 1 || official.conversationCalls != 1 || official.postCalls != 1 || official.editCalls != 1 || official.dmConversationCalls != 1 || official.dmMessageCalls != 1 || official.markDMReadCalls != 1 || official.trendCalls != 1 {
		t.Fatalf("official calls = %#v", official)
	}
	if webRequests.Load() != 0 {
		t.Fatalf("web requests = %d", webRequests.Load())
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
