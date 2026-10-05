package northapi

import (
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
)

func TestPublicDMMessagePage(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	edited := created.Add(time.Minute)
	name := "North team"
	bob := unofficial.User{User: north.User{ID: "user-2", Handle: "bob"}}
	message := unofficial.DMMessage{
		ID:             "message-1",
		ConversationID: "conversation-1",
		Sender:         unofficial.User{User: north.User{ID: "user-1", Handle: "alice"}},
		Text:           "hello",
		CreatedAt:      created,
		EditedAt:       &edited,
		Read:           true,
		Media: []unofficial.Media{{
			Media: north.Media{ID: "media-1", Kind: north.MediaPhoto},
		}},
		Reactions: []unofficial.DMReaction{{
			Emoji:           "👍",
			Count:           1,
			ReactedByViewer: true,
			Users:           []unofficial.User{bob},
		}},
		ReplyTo: &unofficial.DMReply{
			ID:       "message-0",
			Sender:   &bob,
			Text:     "previous",
			HasMedia: true,
		},
		Post: &unofficial.Post{Post: north.Post{ID: "post-1", Text: "shared"}},
	}
	page := unofficial.DMMessagePage{
		Items: []unofficial.DMMessage{message},
		Conversation: unofficial.DMConversation{
			ID:           "conversation-1",
			Name:         &name,
			Group:        true,
			Participants: []unofficial.User{message.Sender, bob},
			LastMessage:  &message,
			UnreadCount:  2,
			UpdatedAt:    edited,
		},
	}

	got := publicDMMessagePage(page)
	if len(got.Items) != 1 || got.Items[0].Sender.Handle != "alice" || got.Items[0].ConversationID != "conversation-1" {
		t.Fatalf("messages = %#v", got.Items)
	}
	converted := got.Items[0]
	if len(converted.Media) != 1 || converted.Media[0].ID != "media-1" || converted.Post == nil || converted.Post.ID != "post-1" {
		t.Fatalf("message attachments = %#v", converted)
	}
	if len(converted.Reactions) != 1 || len(converted.Reactions[0].Users) != 1 || converted.Reactions[0].Users[0].Handle != "bob" {
		t.Fatalf("message reactions = %#v", converted.Reactions)
	}
	if converted.ReplyTo == nil || converted.ReplyTo.Sender == nil || converted.ReplyTo.Sender.Handle != "bob" || !converted.ReplyTo.HasMedia {
		t.Fatalf("message reply = %#v", converted.ReplyTo)
	}
	if got.Conversation.Name == nil || *got.Conversation.Name != name || len(got.Conversation.Participants) != 2 || got.Conversation.LastMessage == nil || got.Conversation.LastMessage.ID != "message-1" {
		t.Fatalf("conversation = %#v", got.Conversation)
	}
}

func TestPublicTrends(t *testing.T) {
	t.Parallel()

	got := publicTrends([]unofficial.Trend{{Tag: "north", Count: 42, IsHashtag: true}})
	if len(got) != 1 || got[0] != (north.Trend{Tag: "north", Count: 42, IsHashtag: true}) {
		t.Fatalf("trends = %#v", got)
	}
}
