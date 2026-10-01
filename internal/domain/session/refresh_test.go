package session

import (
	"context"
	"errors"
	"testing"
)

func TestRefreshRequestRunsAndCompletesOnce(t *testing.T) {
	t.Parallel()

	calls := 0
	request := NewRefreshRequest("Firefox", func(context.Context) error {
		calls++

		return nil
	})
	if request.Profile() != "Firefox" {
		t.Fatalf("profile = %q", request.Profile())
	}
	if err := request.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	request.Complete(nil)
	request.Complete(errors.New("late failure"))
	if err := request.Wait(context.Background()); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if calls != 1 {
		t.Fatalf("refresh calls = %d", calls)
	}
}
