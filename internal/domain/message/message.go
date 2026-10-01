package message

import (
	"context"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
)

type Conversation = unofficial.DMConversation

type ConversationPage = unofficial.DMConversationPage

type Message = unofficial.DMMessage

type MessagePage = unofficial.DMMessagePage

type API interface {
	DMConversations(context.Context, string, bool) (ConversationPage, *north.Response, error)
	DMMessages(context.Context, string, string) (MessagePage, *north.Response, error)
	MarkDMRead(context.Context, string) (*north.Response, error)
}
