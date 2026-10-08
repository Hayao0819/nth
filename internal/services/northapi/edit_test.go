package northapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

func TestHybridUsesGoNorthConditionalEdit(t *testing.T) {
	t.Parallel()

	const etag = `"version-1"`
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/2/tweets/post-1":
			writer.Header().Set("ETag", etag)
			writeJSON(t, writer, `{"data":{"id":"post-1","text":"original","editEligible":true,"media":[{"id":"media-1","kind":"PHOTO"}]}}`)
		case request.Method == http.MethodPut && request.URL.Path == "/2/tweets/post-1/edit":
			if got := request.Header.Get("If-Match"); got != etag {
				t.Errorf("If-Match = %q", got)
			}
			var body north.EditPostRequest
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Text == nil || *body.Text != "updated" || body.MediaIDs == nil || len(*body.MediaIDs) != 1 || (*body.MediaIDs)[0] != "media-1" {
				t.Errorf("edit body = %#v", body)
			}
			writeJSON(t, writer, `{"data":{"id":"post-1","text":"updated","editEligible":true}}`)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	official, err := north.NewClient(
		"nth_oat_test",
		north.WithBaseURL(server.URL),
		north.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := NewHybrid(official, nil, "", nil)
	post, eligible, gotETag, _, err := client.EditablePost(context.Background(), "post-1")
	if err != nil || post.ID != "post-1" || !eligible || gotETag != etag {
		t.Fatalf("EditablePost = %#v, %v, %q, %v", post, eligible, gotETag, err)
	}
	updated, _, err := client.EditPost(context.Background(), domain.PostEdit{
		ID: post.ID, Text: "updated", MediaIDs: []string{"media-1"}, ETag: gotETag, Base: post,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != "post-1" || updated.Text != "updated" {
		t.Fatalf("EditPost returned %#v", updated)
	}
}

func TestHybridSendsEmptyMediaArrayWhenRemovingAllMedia(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/2/tweets/post-1/edit" {
			http.NotFound(writer, request)
			return
		}
		var body struct {
			MediaIDs *[]string `json:"mediaIds"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.MediaIDs == nil || len(*body.MediaIDs) != 0 {
			t.Errorf("mediaIds = %#v, want an empty array", body.MediaIDs)
		}
		writeJSON(t, writer, `{"data":{"id":"post-1","text":"updated","editEligible":true}}`)
	}))
	t.Cleanup(server.Close)

	official, err := north.NewClient(
		"nth_oat_test",
		north.WithBaseURL(server.URL),
		north.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := NewHybrid(official, nil, "", nil)
	if _, _, err := client.EditPost(context.Background(), domain.PostEdit{ID: "post-1", Text: "updated"}); err != nil {
		t.Fatal(err)
	}
}

func TestHybridRefreshesETagAndRetriesUnchangedPost(t *testing.T) {
	t.Parallel()

	const (
		oldETag = `"version-1"`
		newETag = `"version-2"`
	)
	putCalls := 0
	getCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodPut && request.URL.Path == "/2/tweets/post-1/edit":
			putCalls++
			if putCalls == 1 {
				if got := request.Header.Get("If-Match"); got != oldETag {
					t.Errorf("first If-Match = %q", got)
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusPreconditionFailed)
				_, _ = writer.Write([]byte(`{"errors":[{"code":89,"message":"stale"}]}`))

				return
			}
			if got := request.Header.Get("If-Match"); got != newETag {
				t.Errorf("retried If-Match = %q", got)
			}
			writeJSON(t, writer, `{"data":{"id":"post-1","text":"updated","media":[]}}`)
		case request.Method == http.MethodGet && request.URL.Path == "/2/tweets/post-1":
			getCalls++
			writer.Header().Set("ETag", newETag)
			writeJSON(t, writer, `{"data":{"id":"post-1","text":"original","media":[]}}`)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	official, err := north.NewClient(
		"nth_oat_test",
		north.WithBaseURL(server.URL),
		north.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := NewHybrid(official, nil, "", nil)
	updated, _, err := client.EditPost(context.Background(), domain.PostEdit{
		ID: "post-1", Text: "updated", ETag: oldETag,
		Base: north.Post{ID: "post-1", Text: "original"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Text != "updated" || putCalls != 2 || getCalls != 1 {
		t.Fatalf("updated = %#v, PUTs = %d, GETs = %d", updated, putCalls, getCalls)
	}
}

func TestHybridDoesNotOverwriteConcurrentPostEdit(t *testing.T) {
	t.Parallel()

	putCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodPut && request.URL.Path == "/2/tweets/post-1/edit":
			putCalls++
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusPreconditionFailed)
			_, _ = writer.Write([]byte(`{"errors":[{"code":89,"message":"stale"}]}`))
		case request.Method == http.MethodGet && request.URL.Path == "/2/tweets/post-1":
			writer.Header().Set("ETag", `"version-2"`)
			writeJSON(t, writer, `{"data":{"id":"post-1","text":"changed elsewhere","media":[]}}`)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	official, err := north.NewClient(
		"nth_oat_test",
		north.WithBaseURL(server.URL),
		north.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := NewHybrid(official, nil, "", nil)
	latest, response, err := client.EditPost(context.Background(), domain.PostEdit{
		ID: "post-1", Text: "my draft", ETag: `"version-1"`,
		Base: north.Post{ID: "post-1", Text: "original"},
	})
	if !errors.Is(err, errPostEditConflict) {
		t.Fatalf("EditPost error = %v", err)
	}
	if latest.Text != "changed elsewhere" || response == nil || response.StatusCode != http.StatusPreconditionFailed ||
		response.Header.Get("ETag") != `"version-2"` || putCalls != 1 {
		t.Fatalf("latest = %#v, response = %#v, PUTs = %d", latest, response, putCalls)
	}
}
