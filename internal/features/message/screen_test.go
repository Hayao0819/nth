package message

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	messagedomain "github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/testkit"
)

type messageAPI struct {
	conversations messagedomain.MessageConversationPage
	messages      map[string]messagedomain.MessagePage
	messageCalls  []string
	readCalls     []string
}

func (a *messageAPI) DMConversations(context.Context, string, bool) (messagedomain.MessageConversationPage, *north.Response, error) {
	return a.conversations, nil, nil
}

func (a *messageAPI) DMMessages(_ context.Context, conversationID, cursor string) (messagedomain.MessagePage, *north.Response, error) {
	a.messageCalls = append(a.messageCalls, conversationID+":"+cursor)

	return a.messages[cursor], nil, nil
}

func (a *messageAPI) MarkDMRead(_ context.Context, conversationID string) (*north.Response, error) {
	a.readCalls = append(a.readCalls, conversationID)

	return nil, nil
}

func TestScreenOpensConversationAndPaginatesMessages(t *testing.T) {
	t.Parallel()

	next := "next"
	conversation := messagedomain.MessageConversation{
		ID:           "conversation-1",
		Participants: []north.User{{Name: "Alice", Handle: "alice"}},
		LastMessage:  &messagedomain.Message{Text: "latest message"},
		UnreadCount:  2,
	}
	api := &messageAPI{
		conversations: messagedomain.MessageConversationPage{Items: []messagedomain.MessageConversation{conversation}},
		messages: map[string]messagedomain.MessagePage{
			"": {
				Conversation: conversation,
				Items:        []messagedomain.Message{{ID: "message-1", Text: "first message", Sender: north.User{Name: "Alice", Handle: "alice"}}},
				NextCursor:   &next,
			},
			"next": {
				Conversation: conversation,
				Items:        []messagedomain.Message{{ID: "message-2", Text: "older message", Sender: north.User{Name: "Bob", Handle: "bob"}}},
			},
		},
	}
	screen := New(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(70, 18))
	program.Start()
	if plain := testkit.Plain(program); !strings.Contains(plain, "Alice") || !strings.Contains(plain, "latest message") || !strings.Contains(plain, "2 unread") {
		t.Fatalf("conversation list is incomplete:\n%s", plain)
	}

	testkit.SendKeys(program, "enter")
	if plain := testkit.Plain(program); !strings.Contains(plain, "first message") {
		t.Fatalf("conversation did not open:\n%s", plain)
	}
	if got := strings.Join(api.readCalls, ","); got != "conversation-1" {
		t.Fatalf("read calls = %q", got)
	}

	testkit.SendKeys(program, "G")
	if got := strings.Join(api.messageCalls, ","); got != "conversation-1:,conversation-1:next" {
		t.Fatalf("message calls = %q", got)
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "older message") {
		t.Fatalf("older messages were not appended:\n%s", plain)
	}
	testkit.SendKeys(program, "G")
	command := screen.Update(program.Ctx(), testkit.Key("u"))
	message, ok := command().(navigation.OpenUserMsg)
	if !ok || message.User.Handle != "bob" {
		t.Fatalf("open sender = %#v", message)
	}
}

func TestClickingRenderedConversationOpensThatConversation(t *testing.T) {
	t.Parallel()

	conversations := []messagedomain.MessageConversation{
		{
			ID:           "conversation-1",
			Participants: []north.User{{Name: "Alice", Handle: "alice"}},
			LastMessage:  &messagedomain.Message{Text: "from Alice"},
		},
		{
			ID:           "conversation-2",
			Participants: []north.User{{Name: "Bob", Handle: "bob"}},
			LastMessage:  &messagedomain.Message{Text: "from Bob"},
		},
	}
	api := &messageAPI{
		conversations: messagedomain.MessageConversationPage{Items: conversations},
		messages: map[string]messagedomain.MessagePage{
			"": {
				Conversation: conversations[1],
				Items: []messagedomain.Message{
					{ID: "message-1", Text: "newer message", Sender: north.User{Name: "Carol", Handle: "carol"}},
					{ID: "message-2", Text: "older message", Sender: north.User{Name: "Dave", Handle: "dave"}},
				},
			},
		},
	}
	screen := New(api, ui.NewTheme())
	program := reactea.New(screen, reactea.WithSize(70, 18))
	program.Start()

	if !testkit.ClickText(program, "Bob") {
		t.Fatal("Bob was not rendered")
	}
	if screen.current == nil || screen.current.ID != "conversation-2" {
		t.Fatalf("clicked conversation = %#v", screen.current)
	}
	if got := strings.Join(api.messageCalls, ","); got != "conversation-2:" {
		t.Fatalf("message calls = %q", got)
	}

	if !testkit.ClickText(program, "older message") || screen.messageIndex != 1 {
		t.Fatalf("clicked message index = %d", screen.messageIndex)
	}
	positions := testkit.Find(program, "Dave")
	if len(positions) == 0 {
		t.Fatal("Dave was not rendered")
	}
	point := positions[len(positions)-1]
	command := screen.Update(program.Ctx(), tea.MouseClickMsg{X: point.X, Y: point.Y, Button: tea.MouseLeft})
	if command == nil {
		t.Fatal("clicking a message sender returned no command")
	}
	message, ok := command().(navigation.OpenUserMsg)
	if !ok || message.User.Handle != "dave" {
		t.Fatalf("open sender = %#v", message)
	}
}

var _ messagedomain.MessageAPI = (*messageAPI)(nil)
