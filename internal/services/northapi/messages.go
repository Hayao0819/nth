package northapi

import (
	"context"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
)

func (c *Client) DMConversations(ctx context.Context, cursor string, requests bool) (north.DMConversationPage, *north.Response, error) {
	page, response, err := c.web.DMConversations(ctx, cursor, requests)

	return publicDMConversationPage(page), publicResponse(response), err
}

func (c *Client) CreateDMConversation(ctx context.Context, handles ...string) (north.DMConversation, *north.Response, error) {
	conversation, response, err := c.web.CreateDMConversation(ctx, handles...)

	return publicDMConversation(conversation), publicResponse(response), err
}

func (c *Client) DMMessages(ctx context.Context, conversationID, cursor string) (north.DMMessagePage, *north.Response, error) {
	page, response, err := c.web.DMMessages(ctx, conversationID, cursor)

	return publicDMMessagePage(page), publicResponse(response), err
}

func (c *Client) SendDM(ctx context.Context, conversationID string, request north.SendDMRequest) (north.DMMessage, *north.Response, error) {
	replyToID := ""
	if request.ReplyToID != nil {
		replyToID = *request.ReplyToID
	}
	response, err := c.web.SendDM(ctx, conversationID, unofficial.SendDMRequest{
		Text:      request.Text,
		MediaIDs:  append([]string(nil), request.MediaIDs...),
		ReplyToID: replyToID,
	})

	return north.DMMessage{ConversationID: conversationID, Text: request.Text}, publicResponse(response), err
}

func (c *Client) EditDM(ctx context.Context, conversationID, messageID, text string) (north.DMMessage, *north.Response, error) {
	response, err := c.web.EditDM(ctx, conversationID, messageID, text)

	return north.DMMessage{ID: messageID, ConversationID: conversationID, Text: text}, publicResponse(response), err
}

func (c *Client) DeleteDM(ctx context.Context, conversationID, messageID string, everyone bool) (bool, *north.Response, error) {
	response, err := c.web.DeleteDM(ctx, conversationID, messageID, everyone)

	return err == nil, publicResponse(response), err
}

func (c *Client) SetDMReaction(ctx context.Context, conversationID, messageID, emoji string) ([]north.DMReaction, *north.Response, error) {
	response, err := c.web.SetDMReaction(ctx, conversationID, messageID, emoji)

	return nil, publicResponse(response), err
}

func (c *Client) RemoveDMReaction(ctx context.Context, conversationID, messageID string) ([]north.DMReaction, *north.Response, error) {
	response, err := c.web.RemoveDMReaction(ctx, conversationID, messageID)

	return nil, publicResponse(response), err
}

func (c *Client) MarkDMRead(ctx context.Context, conversationID string) (*north.Response, error) {
	response, err := c.web.MarkDMRead(ctx, conversationID)

	return publicResponse(response), err
}

func (c *Client) AcceptDMRequest(ctx context.Context, conversationID string) (bool, *north.Response, error) {
	response, err := c.web.AcceptDMRequest(ctx, conversationID)

	return err == nil, publicResponse(response), err
}

func (c *Client) DeleteDMRequest(ctx context.Context, conversationID string) (bool, *north.Response, error) {
	response, err := c.web.DeleteDMRequest(ctx, conversationID)

	return err == nil, publicResponse(response), err
}
