package northapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
)

func publicPostPage(page unofficial.PostPage) north.PostPage {
	return page.PublicPostPage()
}

func publicPostPageResult(page unofficial.PostPage, response *unofficial.Response, err error) (north.PostPage, *north.Response, error) {
	return publicPostPage(page), publicResponse(response), err
}

func publicPosts(posts []unofficial.Post) []north.Post {
	result := make([]north.Post, len(posts))
	for index := range posts {
		result[index] = posts[index].PublicPost()
	}

	return result
}

func publicUsers(users []unofficial.User) []north.User {
	result := make([]north.User, len(users))
	for index := range users {
		result[index] = users[index].User
	}

	return result
}

func publicTrends(trends []unofficial.Trend) []north.Trend {
	result := make([]north.Trend, len(trends))
	for index, trend := range trends {
		result[index] = north.Trend{
			Tag:       trend.Tag,
			Count:     trend.Count,
			IsHashtag: trend.IsHashtag,
		}
	}

	return result
}

func publicDMConversationPage(page unofficial.DMConversationPage) north.DMConversationPage {
	result := north.DMConversationPage{
		Items:        make([]north.DMConversation, len(page.Items)),
		NextCursor:   page.NextCursor,
		RequestCount: page.RequestCount,
	}
	for index := range page.Items {
		result.Items[index] = publicDMConversation(page.Items[index])
	}

	return result
}

func publicDMConversation(conversation unofficial.DMConversation) north.DMConversation {
	result := north.DMConversation{
		ID:           conversation.ID,
		Name:         conversation.Name,
		Group:        conversation.Group,
		Participants: publicUsers(conversation.Participants),
		UnreadCount:  conversation.UnreadCount,
		Request:      conversation.Request,
		UpdatedAt:    conversation.UpdatedAt,
	}
	if conversation.LastMessage != nil {
		message := publicDMMessage(*conversation.LastMessage)
		result.LastMessage = &message
	}

	return result
}

func publicDMMessagePage(page unofficial.DMMessagePage) north.DMMessagePage {
	result := north.DMMessagePage{
		Items:        make([]north.DMMessage, len(page.Items)),
		NextCursor:   page.NextCursor,
		Conversation: publicDMConversation(page.Conversation),
	}
	for index := range page.Items {
		result.Items[index] = publicDMMessage(page.Items[index])
	}

	return result
}

func publicDMMessage(message unofficial.DMMessage) north.DMMessage {
	result := north.DMMessage{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Sender:         message.Sender.User,
		Text:           message.Text,
		CreatedAt:      message.CreatedAt,
		EditedAt:       message.EditedAt,
		Read:           message.Read,
		System:         message.System,
		Media:          publicMedia(message.Media),
		Reactions:      publicDMReactions(message.Reactions),
	}
	if message.Post != nil {
		post := message.Post.PublicPost()
		result.Post = &post
	}
	if message.ReplyTo != nil {
		reply := publicDMReply(*message.ReplyTo)
		result.ReplyTo = &reply
	}

	return result
}

func publicMedia(media []unofficial.Media) []north.Media {
	result := make([]north.Media, len(media))
	for index := range media {
		result[index] = media[index].Media
	}

	return result
}

func publicDMReactions(reactions []unofficial.DMReaction) []north.DMReaction {
	result := make([]north.DMReaction, len(reactions))
	for index, reaction := range reactions {
		result[index] = north.DMReaction{
			Emoji:           reaction.Emoji,
			Count:           reaction.Count,
			ReactedByViewer: reaction.ReactedByViewer,
			Users:           publicUsers(reaction.Users),
		}
	}

	return result
}

func publicDMReply(reply unofficial.DMReply) north.DMReply {
	result := north.DMReply{
		ID:       reply.ID,
		Deleted:  reply.Deleted,
		Text:     reply.Text,
		HasMedia: reply.HasMedia,
	}
	if reply.Sender != nil {
		sender := reply.Sender.User
		result.Sender = &sender
	}

	return result
}

func publicResponse(response *unofficial.Response) *north.Response {
	if response == nil {
		return nil
	}

	return &north.Response{
		StatusCode: response.StatusCode,
		Header:     response.Header,
		RateLimit:  rateLimit(response.Header),
	}
}

func rateLimit(header http.Header) north.RateLimit {
	limitValue := header.Get("X-Rate-Limit-Limit")
	remainingValue := header.Get("X-Rate-Limit-Remaining")
	resetValue := header.Get("X-Rate-Limit-Reset")
	if limitValue == "" && remainingValue == "" && resetValue == "" {
		return north.RateLimit{}
	}
	limit, _ := strconv.Atoi(limitValue)
	remaining, _ := strconv.Atoi(remainingValue)
	reset, _ := strconv.ParseInt(resetValue, 10, 64)
	rate := north.RateLimit{Present: true, Limit: limit, Remaining: remaining}
	if reset > 0 {
		rate.Reset = time.Unix(reset, 0)
	}

	return rate
}
