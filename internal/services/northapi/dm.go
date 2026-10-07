package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
)

type officialMessageAPI interface {
	DMConversations(context.Context, string, bool) (north.DMConversationPage, *north.Response, error)
	DMMessages(context.Context, string, string) (north.DMMessagePage, *north.Response, error)
	MarkDMRead(context.Context, string) (bool, *north.Response, error)
}

type officialMessageWriterAPI interface {
	CreateDMConversation(context.Context, ...string) (north.DMConversation, *north.Response, error)
	SendDM(context.Context, string, north.SendDMRequest) (north.DMMessage, *north.Response, error)
	EditDM(context.Context, string, string, string) (north.DMMessage, *north.Response, error)
	DeleteDM(context.Context, string, string, bool) (bool, *north.Response, error)
	SetDMReaction(context.Context, string, string, string) ([]north.DMReaction, *north.Response, error)
	RemoveDMReaction(context.Context, string, string) ([]north.DMReaction, *north.Response, error)
}

type officialMessageRequestAPI interface {
	AcceptDMRequest(context.Context, string) (bool, *north.Response, error)
	DeleteDMRequest(context.Context, string) (bool, *north.Response, error)
}

type officialMessageGroupAPI interface {
	RenameDMConversation(context.Context, string, string) (bool, *north.Response, error)
	AddDMConversationMembers(context.Context, string, ...string) (bool, *north.Response, error)
	LeaveDMConversation(context.Context, string) (bool, *north.Response, error)
}

type officialMessageRecipientAPI interface {
	DMRecipients(context.Context, string, string) (north.UserPage, *north.Response, error)
}

type officialMessageUnreadAPI interface {
	DMUnreadCount(context.Context) (int, *north.Response, error)
}

var errMessageOperationUnavailable = errors.New("message operation is not available with the configured credentials")

func (c *Hybrid) DMUnreadCount(ctx context.Context) (int, *north.Response, error) {
	api, ok := c.official.(officialMessageUnreadAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return 0, nil, errMessageOperationUnavailable
	}

	return api.DMUnreadCount(ctx)
}

func (c *Hybrid) DMConversations(ctx context.Context, cursor string, requests bool) (north.DMConversationPage, *north.Response, error) {
	if official, ok := c.official.(officialMessageAPI); ok && c.officialSupportsScopedAPI() {
		return official.DMConversations(ctx, cursor, requests)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMConversationPage, *north.Response, error) {
		return client.DMConversations(ctx, cursor, requests)
	})
}

func (c *Hybrid) CreateDMConversation(ctx context.Context, handles ...string) (north.DMConversation, *north.Response, error) {
	if official, ok := c.official.(officialMessageWriterAPI); ok && c.officialSupportsScopedAPI() {
		return official.CreateDMConversation(ctx, handles...)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMConversation, *north.Response, error) {
		return client.CreateDMConversation(ctx, handles...)
	})
}

func (c *Hybrid) DMMessages(ctx context.Context, conversationID, cursor string) (north.DMMessagePage, *north.Response, error) {
	if official, ok := c.official.(officialMessageAPI); ok && c.officialSupportsScopedAPI() {
		return official.DMMessages(ctx, conversationID, cursor)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMMessagePage, *north.Response, error) {
		return client.DMMessages(ctx, conversationID, cursor)
	})
}

func (c *Hybrid) SendDM(ctx context.Context, conversationID string, request north.SendDMRequest) (north.DMMessage, *north.Response, error) {
	if official, ok := c.official.(officialMessageWriterAPI); ok && c.officialSupportsScopedAPI() {
		return official.SendDM(ctx, conversationID, request)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMMessage, *north.Response, error) {
		return client.SendDM(ctx, conversationID, request)
	})
}

func (c *Hybrid) EditDM(ctx context.Context, conversationID, messageID, text string) (north.DMMessage, *north.Response, error) {
	if official, ok := c.official.(officialMessageWriterAPI); ok && c.officialSupportsScopedAPI() {
		return official.EditDM(ctx, conversationID, messageID, text)
	}

	return withWeb(ctx, c, func(client *Client) (north.DMMessage, *north.Response, error) {
		return client.EditDM(ctx, conversationID, messageID, text)
	})
}

func (c *Hybrid) DeleteDM(ctx context.Context, conversationID, messageID string, everyone bool) (bool, *north.Response, error) {
	if official, ok := c.official.(officialMessageWriterAPI); ok && c.officialSupportsScopedAPI() {
		return official.DeleteDM(ctx, conversationID, messageID, everyone)
	}

	return withWeb(ctx, c, func(client *Client) (bool, *north.Response, error) {
		return client.DeleteDM(ctx, conversationID, messageID, everyone)
	})
}

func (c *Hybrid) SetDMReaction(ctx context.Context, conversationID, messageID, emoji string) ([]north.DMReaction, *north.Response, error) {
	if official, ok := c.official.(officialMessageWriterAPI); ok && c.officialSupportsScopedAPI() {
		return official.SetDMReaction(ctx, conversationID, messageID, emoji)
	}

	return withWeb(ctx, c, func(client *Client) ([]north.DMReaction, *north.Response, error) {
		return client.SetDMReaction(ctx, conversationID, messageID, emoji)
	})
}

func (c *Hybrid) RemoveDMReaction(ctx context.Context, conversationID, messageID string) ([]north.DMReaction, *north.Response, error) {
	if official, ok := c.official.(officialMessageWriterAPI); ok && c.officialSupportsScopedAPI() {
		return official.RemoveDMReaction(ctx, conversationID, messageID)
	}

	return withWeb(ctx, c, func(client *Client) ([]north.DMReaction, *north.Response, error) {
		return client.RemoveDMReaction(ctx, conversationID, messageID)
	})
}

func (c *Hybrid) MarkDMRead(ctx context.Context, conversationID string) (*north.Response, error) {
	if official, ok := c.official.(officialMessageAPI); ok && c.officialSupportsScopedAPI() {
		_, response, err := official.MarkDMRead(ctx, conversationID)

		return response, err
	}

	_, response, err := withWeb(ctx, c, func(client *Client) (struct{}, *north.Response, error) {
		response, err := client.MarkDMRead(ctx, conversationID)

		return struct{}{}, response, err
	})

	return response, err
}

func (c *Hybrid) AcceptDMRequest(ctx context.Context, conversationID string) (bool, *north.Response, error) {
	if official, ok := c.official.(officialMessageRequestAPI); ok && c.officialSupportsScopedAPI() {
		return official.AcceptDMRequest(ctx, conversationID)
	}

	return withWeb(ctx, c, func(client *Client) (bool, *north.Response, error) {
		return client.AcceptDMRequest(ctx, conversationID)
	})
}

func (c *Hybrid) DeleteDMRequest(ctx context.Context, conversationID string) (bool, *north.Response, error) {
	if official, ok := c.official.(officialMessageRequestAPI); ok && c.officialSupportsScopedAPI() {
		return official.DeleteDMRequest(ctx, conversationID)
	}

	return withWeb(ctx, c, func(client *Client) (bool, *north.Response, error) {
		return client.DeleteDMRequest(ctx, conversationID)
	})
}

func (c *Hybrid) RenameDMConversation(ctx context.Context, conversationID, name string) (bool, *north.Response, error) {
	official, ok := c.official.(officialMessageGroupAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errMessageOperationUnavailable
	}

	return official.RenameDMConversation(ctx, conversationID, name)
}

func (c *Hybrid) AddDMConversationMembers(ctx context.Context, conversationID string, handles ...string) (bool, *north.Response, error) {
	official, ok := c.official.(officialMessageGroupAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errMessageOperationUnavailable
	}

	return official.AddDMConversationMembers(ctx, conversationID, handles...)
}

func (c *Hybrid) LeaveDMConversation(ctx context.Context, conversationID string) (bool, *north.Response, error) {
	official, ok := c.official.(officialMessageGroupAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errMessageOperationUnavailable
	}

	return official.LeaveDMConversation(ctx, conversationID)
}

func (c *Hybrid) DMRecipients(ctx context.Context, query, cursor string) (north.UserPage, *north.Response, error) {
	official, ok := c.official.(officialMessageRecipientAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.UserPage{}, nil, errMessageOperationUnavailable
	}

	return official.DMRecipients(ctx, query, cursor)
}
