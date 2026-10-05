package message

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type Conversation = north.DMConversation

type ConversationPage = north.DMConversationPage

type Message = north.DMMessage

type MessagePage = north.DMMessagePage

type API interface {
	DMConversations(context.Context, string, bool) (ConversationPage, *north.Response, error)
	DMMessages(context.Context, string, string) (MessagePage, *north.Response, error)
	MarkDMRead(context.Context, string) (*north.Response, error)
}
