//go:build integration

package northapi

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

func TestIntegrationCreateEditAndDeletePost(t *testing.T) {
	if os.Getenv("NORTH_INTEGRATION_WRITE") != "1" {
		t.Skip("set NORTH_INTEGRATION_WRITE=1 to run the write integration test")
	}
	token := os.Getenv("NORTH_API_KEY")
	if token == "" {
		t.Skip("NORTH_API_KEY is not set")
	}

	official, err := north.NewClient(
		token,
		north.WithUserAgent("nth/edit-integration-test"),
		north.WithHTTPClient(NewOfficialHTTPClient()),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := NewHybrid(official, nil, "", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	original := fmt.Sprintf("nth edit integration %d", time.Now().UnixNano())
	created, _, err := client.CreatePost(ctx, north.CreatePostRequest{Text: original})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		deleted, _, deleteErr := client.DeletePost(cleanupCtx, created.ID)
		if deleteErr != nil || !deleted {
			t.Errorf("DeletePost cleanup: deleted=%v err=%v", deleted, deleteErr)
		}
	}()

	post, eligible, etag, _, err := client.EditablePost(ctx, created.ID)
	if err != nil {
		t.Fatalf("EditablePost: %v", err)
	}
	if !eligible || etag == "" {
		t.Fatalf("new post is not conditionally editable: eligible=%v etag=%q", eligible, etag)
	}

	edited := original + " edited"
	if _, _, err := client.EditPost(ctx, domain.PostEdit{ID: post.ID, Text: edited, ETag: etag, Base: post}); err != nil {
		t.Fatalf("EditPost: %v", err)
	}
	updated, _, err := client.Post(ctx, post.ID)
	if err != nil {
		t.Fatalf("Post after edit: %v", err)
	}
	if updated.Text != edited || updated.EditCount < 1 || updated.EditedAt == nil {
		t.Fatalf("edited post was not returned: text=%q editCount=%d editedAt=%v", updated.Text, updated.EditCount, updated.EditedAt)
	}
}
