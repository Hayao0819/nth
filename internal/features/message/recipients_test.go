package message

import (
	"context"
	"strings"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type recipientAPI struct {
	queries []string
	users   []north.User
}

func (a *recipientAPI) DMRecipients(_ context.Context, query, _ string) (north.UserPage, *north.Response, error) {
	a.queries = append(a.queries, query)

	return north.UserPage{Items: a.users}, nil, nil
}

func TestRecipientPickerSearchesAndSelectsAccounts(t *testing.T) {
	t.Parallel()

	api := &recipientAPI{users: []north.User{
		{ID: "alice", Name: "Alice", Handle: "Alice"},
		{ID: "bob", Name: "Bob", Handle: "bob"},
	}}
	picker := newRecipientPicker(api, ui.NewTheme(), "New message", "Start")
	program := reactea.New(picker, reactea.WithSize(64, 20))
	program.Start()
	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Alice") || !strings.Contains(plain, "Bob") {
		t.Fatalf("recipient results are incomplete:\n%s", plain)
	}
	if len(api.queries) != 1 || api.queries[0] != "" {
		t.Fatalf("queries = %#v", api.queries)
	}

	picker.toggleSelected()
	picker.selected = 1
	picker.toggleSelected()
	if len(picker.chosen) != 2 || picker.chosen["alice"].ID != "alice" || picker.chosen["bob"].ID != "bob" {
		t.Fatalf("chosen accounts = %#v", picker.chosen)
	}
	picker.toggleSelected()
	if len(picker.chosen) != 1 {
		t.Fatalf("toggle did not remove the account: %#v", picker.chosen)
	}
}
