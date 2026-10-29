package message

import (
	"reflect"
	"testing"

	messagedomain "github.com/Hayao0819/nth/internal/domain"
)

func TestMergeRefreshedMessagesKeepsHistoryAndUpdatesTheLatestPage(t *testing.T) {
	t.Parallel()

	current := []messagedomain.Message{
		{ID: "old", Text: "old"},
		{ID: "middle", Text: "before refresh"},
		{ID: "latest", Text: "latest"},
	}
	refreshed := []messagedomain.Message{
		{ID: "middle", Text: "after refresh"},
		{ID: "latest", Text: "latest"},
		{ID: "sent", Text: "new message"},
	}
	got := mergeRefreshedMessages(current, refreshed)
	want := []messagedomain.Message{
		{ID: "old", Text: "old"},
		{ID: "middle", Text: "after refresh"},
		{ID: "latest", Text: "latest"},
		{ID: "sent", Text: "new message"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeRefreshedMessages() = %#v, want %#v", got, want)
	}
}
