package domain

import (
	"context"

	"github.com/Hayao0819/go-north"
)

type MessageConversation = north.DMConversation

type MessageConversationPage = north.DMConversationPage

type Message = north.DMMessage

type MessagePage = north.DMMessagePage

type MessageAPI interface {
	DMConversations(context.Context, string, bool) (MessageConversationPage, *north.Response, error)
	DMMessages(context.Context, string, string) (MessagePage, *north.Response, error)
	MarkDMRead(context.Context, string) (*north.Response, error)
}

type MessageWriterAPI interface {
	CreateDMConversation(context.Context, ...string) (MessageConversation, *north.Response, error)
	SendDM(context.Context, string, north.SendDMRequest) (Message, *north.Response, error)
	EditDM(context.Context, string, string, string) (Message, *north.Response, error)
	DeleteDM(context.Context, string, string, bool) (bool, *north.Response, error)
	SetDMReaction(context.Context, string, string, string) ([]north.DMReaction, *north.Response, error)
	RemoveDMReaction(context.Context, string, string) ([]north.DMReaction, *north.Response, error)
}

type MessageRequestAPI interface {
	AcceptDMRequest(context.Context, string) (bool, *north.Response, error)
	DeleteDMRequest(context.Context, string) (bool, *north.Response, error)
}

type MessageGroupAPI interface {
	RenameDMConversation(context.Context, string, string) (bool, *north.Response, error)
	AddDMConversationMembers(context.Context, string, ...string) (bool, *north.Response, error)
	LeaveDMConversation(context.Context, string) (bool, *north.Response, error)
}

type MessageRecipientAPI interface {
	DMRecipients(context.Context, string, string) (north.UserPage, *north.Response, error)
}

type MessageUnreadAPI interface {
	DMUnreadCount(context.Context) (int, *north.Response, error)
}
