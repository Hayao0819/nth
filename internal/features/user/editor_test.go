package user

import (
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
)

func TestProfileEditorOnlyShowsMediaFieldsWhenSupported(t *testing.T) {
	t.Parallel()

	user := north.User{Name: "Alice"}
	withoutMedia := newProfileEditor(ui.NewTheme(), user, false)
	if len(withoutMedia.inputs) != 4 {
		t.Fatalf("text-only fields = %d", len(withoutMedia.inputs))
	}
	withMedia := newProfileEditor(ui.NewTheme(), user, true)
	if len(withMedia.inputs) != 6 {
		t.Fatalf("media fields = %d", len(withMedia.inputs))
	}
	if withMedia.labels[4] != "Avatar file · blank keeps current" || withMedia.labels[5] != "Header file · blank keeps current" {
		t.Fatalf("media labels = %#v", withMedia.labels[4:])
	}
}
