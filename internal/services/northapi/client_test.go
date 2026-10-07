package northapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
)

func TestClientAdaptsWebAPI(t *testing.T) {
	t.Parallel()

	webPost := `{"id":"post-1","text":"hello","author":{"id":"user-1","handle":"alice"},"media":[{"id":"media-1","kind":"PHOTO","status":"READY","url":"/media/1.jpg","position":0}],"poll":{"id":"poll-1","endsAt":"2026-10-02T00:00:00Z","ended":false,"totalVotes":3,"viewerOptionId":"option-2","options":[{"id":"option-1","label":"One","position":0,"voteCount":1,"percent":33},{"id":"option-2","label":"Two","position":1,"voteCount":2,"percent":67}]},"quoted":{"id":"quoted-1","text":"quoted","author":{"id":"user-2","handle":"bob"},"media":[]},"editEligible":true}`
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Rate-Limit-Limit", "100")
		writer.Header().Set("X-Rate-Limit-Remaining", "99")
		if got := request.Header.Get("Cookie"); got != "session=secret" {
			t.Errorf("Cookie = %q", got)
		}
		switch request.URL.Path {
		case "/api/auth/me":
			writeJSON(t, writer, `{"user":{"id":"user-1","handle":"alice"}}`)
		case "/api/users/bob":
			writeJSON(t, writer, `{"id":"user-2","handle":"bob","name":"Bob"}`)
		case "/api/timeline/home":
			if request.URL.Query().Get("ranked") != "1" {
				t.Errorf("timeline query = %q", request.URL.RawQuery)
			}
			writeJSON(t, writer, `{"items":[`+webPost+`],"nextCursor":"next"}`)
		case "/api/tweets/post-1":
			writeJSON(t, writer, webPost)
		case "/api/tweets/post-1/conversation":
			if request.URL.Query().Get("cursor") != "replies-next" {
				t.Errorf("conversation query = %q", request.URL.RawQuery)
			}
			writeJSON(t, writer, `{"ancestors":[{"id":"parent","text":"parent","author":{"id":"user-2","handle":"bob"},"media":[]}],"tweet":`+webPost+`,"replies":[{"id":"reply-1","text":"reply","author":{"id":"user-3","handle":"carol"},"media":[]}],"nextCursor":"next","readerRootId":"parent"}`)
		case "/api/tweets/post-1/edit":
			if request.Method != http.MethodPut {
				t.Errorf("edit method = %s", request.Method)
			}
			var body unofficial.EditPostRequest
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode edit post: %v", err)
			}
			if body.Text != "updated" || len(body.MediaIDs) != 1 || body.MediaIDs[0] != "media-1" {
				t.Errorf("edit body = %#v", body)
			}
		case "/api/notifications":
			writeJSON(t, writer, `{"items":[{"id":"notice-1","kind":"LIKE","read":false,"actorCount":1,"groupCount":2,"actors":[{"id":"user-2","handle":"bob","name":"Bob"}],"tweet":`+webPost+`}],"nextCursor":"notices-next"}`)
		case "/api/notifications/unread-count":
			writeJSON(t, writer, `{"count":2}`)
		case "/api/notifications/read":
			if request.Method != http.MethodPost {
				t.Errorf("mark-read method = %s", request.Method)
			}
		case "/api/tweets":
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode create post: %v", err)
			}
			if body["text"] != "reply" || body["inReplyToId"] != "post-1" || body["quotedId"] != "quoted" {
				t.Errorf("create post body = %#v", body)
			}
			media, ok := body["mediaIds"].([]any)
			if !ok || len(media) != 1 || media[0] != "media-1" {
				t.Errorf("media IDs = %#v", body["mediaIds"])
			}
			poll, ok := body["poll"].(map[string]any)
			if !ok || poll["durationMinutes"] != float64(60) {
				t.Errorf("poll = %#v", body["poll"])
			}
			options, ok := poll["options"].([]any)
			if !ok || len(options) != 2 || options[0] != "yes" || options[1] != "no" {
				t.Errorf("poll options = %#v", poll["options"])
			}
			writeJSON(t, writer, `{"id":"post-2","text":"reply"}`)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	client, err := New(
		"session=secret",
		unofficial.WithBaseURL(server.URL),
		unofficial.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	user, response, err := client.Me(ctx)
	if err != nil || user.Handle != "alice" || response.StatusCode != http.StatusOK || !response.RateLimit.Present || response.RateLimit.Remaining != 99 {
		t.Fatalf("Me = %#v, %#v, %v", user, response, err)
	}
	user, _, err = client.User(ctx, "bob")
	if err != nil || user.ID != "user-2" || user.Name != "Bob" {
		t.Fatalf("User = %#v, %v", user, err)
	}
	page, _, err := client.HomeTimeline(ctx, north.TimelineOptions{Ranked: true})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "post-1" || page.NextCursor == nil || page.Items[0].Poll == nil || page.Items[0].Quoted == nil {
		t.Fatalf("HomeTimeline = %#v, %v", page, err)
	}
	if page.Items[0].Poll.ViewerOptionID == nil || *page.Items[0].Poll.ViewerOptionID != "option-2" || page.Items[0].Poll.Options[1].VoteCount != 2 || len(page.Items[0].Media) != 1 {
		t.Fatalf("converted timeline post = %#v", page.Items[0])
	}
	post, _, err := client.Post(ctx, "post-1")
	if err != nil || post.Text != "hello" || post.Quoted == nil || post.Quoted.ID != "quoted-1" || post.Poll == nil {
		t.Fatalf("Post = %#v, %v", post, err)
	}
	conversation, _, err := client.PostConversation(ctx, "post-1", "replies-next")
	if err != nil || len(conversation.Ancestors) != 1 || conversation.Ancestors[0].ID != "parent" || conversation.Post.ID != "post-1" || len(conversation.Replies) != 1 || conversation.Replies[0].ID != "reply-1" {
		t.Fatalf("PostConversation = %#v, %v", conversation, err)
	}
	if conversation.NextCursor == nil || *conversation.NextCursor != "next" || conversation.ReaderRootID == nil || *conversation.ReaderRootID != "parent" {
		t.Fatalf("PostConversation cursors = %#v", conversation)
	}
	created, _, err := client.CreatePost(ctx, north.CreatePostRequest{
		Text:        "reply",
		Media:       &north.CreatePostMedia{MediaIDs: []string{"media-1"}},
		Reply:       &north.CreatePostReply{InReplyToPostID: "post-1"},
		QuotePostID: "quoted",
		Poll:        &north.CreatePoll{Options: []string{"yes", "no"}, DurationMinutes: 60},
	})
	if err != nil || created.ID != "post-2" {
		t.Fatalf("CreatePost = %#v, %v", created, err)
	}
	post, eligible, _, err := client.EditablePost(ctx, "post-1")
	if err != nil || !eligible || post.ID != "post-1" {
		t.Fatalf("EditablePost = %#v, %v, %v", post, eligible, err)
	}
	if _, err := client.EditPost(ctx, "post-1", "updated", []string{"media-1"}); err != nil {
		t.Fatalf("EditPost: %v", err)
	}
	notifications, _, err := client.Notifications(ctx, north.NotificationsAll, "")
	if err != nil || len(notifications.Items) != 1 || notifications.Items[0].Kind != "LIKE" || notifications.Items[0].GroupCount != 2 || notifications.Items[0].Post == nil || notifications.Items[0].Post.ID != "post-1" || notifications.Items[0].Post.Poll == nil || notifications.NextCursor == nil {
		t.Fatalf("Notifications = %#v, %v", notifications, err)
	}
	if len(notifications.Items[0].Actors) != 1 || notifications.Items[0].Actors[0].Handle != "bob" {
		t.Fatalf("notification actors = %#v", notifications.Items[0].Actors)
	}
	unread, _, err := client.NotificationUnreadCount(ctx)
	if err != nil || unread != 2 {
		t.Fatalf("NotificationUnreadCount = %d, %v", unread, err)
	}
	if _, _, err := client.MarkNotificationsRead(ctx); err != nil {
		t.Fatalf("MarkNotificationsRead: %v", err)
	}
}

func writeJSON(t *testing.T, writer http.ResponseWriter, body string) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Errorf("write response: %v", err)
	}
}
