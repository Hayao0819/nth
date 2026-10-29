package message

import (
	"context"
	"fmt"
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

	testkit.SendKeys(program, "M")
	if got := strings.Join(api.messageCalls, ","); got != "conversation-1:,conversation-1:next" {
		t.Fatalf("message calls = %q", got)
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "older message") {
		t.Fatalf("older messages were not appended:\n%s", plain)
	}
	testkit.SendKeys(program, "g")
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

	if !testkit.ClickText(program, "older message") || screen.messageIndex != 0 {
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

func TestConversationShowsBubblesAtTheLatestMessage(t *testing.T) {
	t.Parallel()

	conversation := messagedomain.MessageConversation{
		ID:           "conversation-1",
		Participants: []north.User{{Name: "Alice", Handle: "alice"}},
	}
	api := &messageAPI{
		conversations: messagedomain.MessageConversationPage{Items: []messagedomain.MessageConversation{conversation}},
		messages: map[string]messagedomain.MessagePage{
			"": {
				Conversation: conversation,
				Items: []messagedomain.Message{
					{ID: "newest", Text: "latest message", Read: true, Sender: north.User{ID: "viewer", Name: "Hayao", Handle: "hayao"}},
					{ID: "middle", Text: "middle message", Sender: north.User{ID: "alice", Name: "Alice", Handle: "alice"}},
					{ID: "oldest", Text: "oldest message", Sender: north.User{ID: "alice", Name: "Alice", Handle: "alice"}},
				},
			},
		},
	}
	screen := New(api, ui.NewTheme())
	screen.SetViewer(north.User{ID: "viewer", Handle: "hayao"})
	program := reactea.New(screen, reactea.WithSize(70, 16))
	program.Start()
	testkit.SendKeys(program, "enter")

	if len(screen.messages) != 3 || screen.messages[0].ID != "oldest" || screen.messages[2].ID != "newest" {
		t.Fatalf("message order = %#v", screen.messages)
	}
	if screen.messageIndex != 2 || screen.messageTop == 0 {
		t.Fatalf("latest position = index %d, top %d", screen.messageIndex, screen.messageTop)
	}
	plain := testkit.Plain(program)
	if !strings.Contains(plain, "latest message") || !strings.Contains(plain, "✓ Seen") || !strings.Contains(plain, "╭") || !strings.Contains(plain, "╰") {
		t.Fatalf("latest bubble is not visible:\n%s", plain)
	}
	if incoming := screen.renderMessage(messagedomain.Message{Text: "received", Read: true, Sender: north.User{ID: "alice", Handle: "alice"}}, 70, false); strings.Contains(incoming, "Seen") {
		t.Fatalf("incoming message has an outgoing read receipt:\n%s", incoming)
	}
	latest := testkit.Find(program, "latest message")
	if len(latest) == 0 || latest[len(latest)-1].Y < 10 {
		t.Fatalf("latest bubble was not anchored near the bottom: %#v\n%s", latest, plain)
	}
}

func TestLongMessageCanBeReadWithoutTruncation(t *testing.T) {
	t.Parallel()

	conversation := messagedomain.MessageConversation{
		ID:           "conversation-1",
		Participants: []north.User{{Name: "Alice", Handle: "alice"}},
	}
	messageLines := make([]string, 24)
	for index := range messageLines {
		messageLines[index] = fmt.Sprintf("message line %02d", index+1)
	}
	api := &messageAPI{
		conversations: messagedomain.MessageConversationPage{Items: []messagedomain.MessageConversation{conversation}},
		messages: map[string]messagedomain.MessagePage{
			"": {
				Conversation: conversation,
				Items: []messagedomain.Message{{
					ID:     "long-message",
					Text:   strings.Join(messageLines, "\n"),
					Sender: north.User{Name: "Alice", Handle: "alice"},
				}},
			},
		},
	}
	screen := New(api, ui.NewTheme())
	screen.write = &messageWriter{}
	program := reactea.New(screen, reactea.WithSize(70, 16))
	program.Start()
	testkit.SendKeys(program, "enter")

	plain := testkit.Plain(program)
	if !strings.Contains(plain, "message line 01") || strings.Contains(plain, "message line 24") {
		t.Fatalf("long message did not start at its first line:\n%s", plain)
	}
	if !strings.Contains(plain, "Ctrl+D/U scroll message") {
		t.Fatalf("long message has no scroll hint:\n%s", plain)
	}

	testkit.SendKeys(program, "ctrl+d", "ctrl+d", "ctrl+d", "ctrl+d")
	plain = testkit.Plain(program)
	if !strings.Contains(plain, "message line 24") {
		t.Fatalf("long message cannot be scrolled to its final line:\n%s", plain)
	}
	if screen.messageOffset == 0 {
		t.Fatal("long message offset did not change")
	}
}

func TestConversationShowsPersistentComposer(t *testing.T) {
	t.Parallel()

	conversation := messagedomain.MessageConversation{
		ID:           "conversation-1",
		Participants: []north.User{{Name: "Alice", Handle: "alice"}},
	}
	api := &messageAPI{
		conversations: messagedomain.MessageConversationPage{Items: []messagedomain.MessageConversation{conversation}},
		messages: map[string]messagedomain.MessagePage{
			"": {Conversation: conversation},
		},
	}
	writer := &messageWriter{}
	screen := New(api, ui.NewTheme())
	screen.write = writer
	screen.SetViewer(north.User{ID: "viewer", Name: "Hayao", Handle: "hayao"})
	program := reactea.New(screen, reactea.WithSize(70, 16))
	program.Start()
	testkit.SendKeys(program, "enter")

	plain := testkit.Plain(program)
	if !strings.Contains(plain, "Write a message") || !strings.Contains(plain, "Enter  Write") {
		t.Fatalf("persistent composer is not visible:\n%s", plain)
	}
	if screen.composing || program.InputCaptured() {
		t.Fatalf("composer unexpectedly captured input: open %v, captured %v", screen.composing, program.InputCaptured())
	}
	testkit.SendKeys(program, "s")
	if screen.composing || screen.composer.Widget.Value() != "" {
		t.Fatalf("s still activates the composer: open %v, value %q", screen.composing, screen.composer.Widget.Value())
	}
	testkit.SendKeys(program, "enter")
	if !program.InputCaptured() {
		t.Fatal("composer did not capture global input")
	}
	testkit.SendKeys(program, "h", "e", "l", "l", "o", "enter")
	if got := strings.Join(writer.sent, ","); got != "hello" {
		t.Fatalf("sent messages = %q", got)
	}
	if !screen.composing || screen.composer.Widget.Value() != "" || !program.InputCaptured() {
		t.Fatalf("composer state after send = open %v, value %q, captured %v", screen.composing, screen.composer.Widget.Value(), program.InputCaptured())
	}
	if len(screen.messages) != 1 || screen.messages[0].Text != "hello" {
		t.Fatalf("sent message was not added in place: %#v", screen.messages)
	}
	testkit.SendKeys(program, "a", "g", "a", "i", "n", "shift+enter", "l", "i", "n", "e", "2")
	testkit.Click(program, 69, 15)
	if got := strings.Join(writer.sent, ","); got != "hello,again\nline2" {
		t.Fatalf("messages after clicking Send = %q", got)
	}
	testkit.SendKeys(program, "esc")
	if screen.composing || program.InputCaptured() {
		t.Fatalf("composer state after close = open %v, captured %v", screen.composing, program.InputCaptured())
	}
	if plain := testkit.Plain(program); !strings.Contains(plain, "Write a message") || !strings.Contains(plain, "Enter  Write") {
		t.Fatalf("composer disappeared after losing focus:\n%s", plain)
	}
}

func TestSendingWhileReadingHistoryPreservesTheViewport(t *testing.T) {
	t.Parallel()

	conversation := messagedomain.MessageConversation{
		ID:           "conversation-1",
		Participants: []north.User{{Name: "Alice", Handle: "alice"}},
	}
	api := &messageAPI{
		conversations: messagedomain.MessageConversationPage{Items: []messagedomain.MessageConversation{conversation}},
		messages: map[string]messagedomain.MessagePage{
			"": {
				Conversation: conversation,
				Items: []messagedomain.Message{
					{ID: "newest", Text: "newest", Sender: north.User{ID: "alice", Handle: "alice"}},
					{ID: "middle", Text: "middle", Sender: north.User{ID: "alice", Handle: "alice"}},
					{ID: "oldest", Text: "oldest", Sender: north.User{ID: "alice", Handle: "alice"}},
				},
			},
		},
	}
	writer := &messageWriter{}
	screen := New(api, ui.NewTheme())
	screen.write = writer
	screen.SetViewer(north.User{ID: "viewer", Name: "Hayao", Handle: "hayao"})
	program := reactea.New(screen, reactea.WithSize(70, 18))
	program.Start()
	testkit.SendKeys(program, "enter", "g")
	selectedID := screen.messages[screen.messageIndex].ID
	top := screen.messageTop

	testkit.SendKeys(program, "enter", "h", "e", "l", "l", "o", "enter")
	if screen.messages[screen.messageIndex].ID != selectedID || screen.messageTop != top {
		t.Fatalf("send reset history position: selected %q, top %d", screen.messages[screen.messageIndex].ID, screen.messageTop)
	}
	if len(screen.messages) != 4 || screen.messages[len(screen.messages)-1].Text != "hello" {
		t.Fatalf("sent message was not appended: %#v", screen.messages)
	}
}

type messageWriter struct {
	sent []string
}

func (*messageWriter) CreateDMConversation(context.Context, ...string) (messagedomain.MessageConversation, *north.Response, error) {
	return messagedomain.MessageConversation{}, nil, nil
}

func (w *messageWriter) SendDM(_ context.Context, _ string, request north.SendDMRequest) (messagedomain.Message, *north.Response, error) {
	w.sent = append(w.sent, request.Text)

	return messagedomain.Message{
		ID:     fmt.Sprintf("sent-%d", len(w.sent)),
		Text:   request.Text,
		Sender: north.User{ID: "viewer", Name: "Hayao", Handle: "hayao"},
	}, nil, nil
}

func (*messageWriter) EditDM(context.Context, string, string, string) (messagedomain.Message, *north.Response, error) {
	return messagedomain.Message{}, nil, nil
}

func (*messageWriter) DeleteDM(context.Context, string, string, bool) (bool, *north.Response, error) {
	return true, nil, nil
}

func (*messageWriter) SetDMReaction(context.Context, string, string, string) ([]north.DMReaction, *north.Response, error) {
	return nil, nil, nil
}

func (*messageWriter) RemoveDMReaction(context.Context, string, string) ([]north.DMReaction, *north.Response, error) {
	return nil, nil, nil
}

var _ messagedomain.MessageAPI = (*messageAPI)(nil)
var _ messagedomain.MessageWriterAPI = (*messageWriter)(nil)
