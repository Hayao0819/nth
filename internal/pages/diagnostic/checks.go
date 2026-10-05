package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
)

type API interface {
	Me(context.Context) (north.User, *north.Response, error)
	HomeTimeline(context.Context, north.TimelineOptions) (north.PostPage, *north.Response, error)
}

type notificationAPI interface {
	Notifications(context.Context, north.NotificationTab, string) (north.NotificationPage, *north.Response, error)
}

type bookmarkAPI interface {
	Bookmarks(context.Context, string) (north.PostPage, *north.Response, error)
}

type messageAPI interface {
	DMConversations(context.Context, string, bool) (north.DMConversationPage, *north.Response, error)
}

type trendAPI interface {
	Trends(context.Context, string) ([]north.Trend, *north.Response, error)
}

type check struct {
	name string
	run  func(context.Context) (string, error)
}

func checks(api API) []check {
	return []check{
		{name: "Account", run: func(ctx context.Context) (string, error) {
			user, _, err := api.Me(ctx)
			if err != nil {
				return "", err
			}

			return ui.SafeInline(user.Name) + "  @" + ui.SafeInline(user.Handle), nil
		}},
		{name: "Timeline", run: func(ctx context.Context) (string, error) {
			page, _, err := api.HomeTimeline(ctx, north.TimelineOptions{})
			if err != nil {
				return "", err
			}

			return postPageSummary(page.Items), nil
		}},
		{name: "Notifications", run: notificationCheck(api)},
		{name: "Bookmarks", run: bookmarkCheck(api)},
		{name: "Messages", run: messageCheck(api)},
		{name: "Trends", run: trendCheck(api)},
	}
}

func notificationCheck(api API) func(context.Context) (string, error) {
	service, ok := api.(notificationAPI)
	if !ok {
		return unavailable
	}

	return func(ctx context.Context) (string, error) {
		page, _, err := service.Notifications(ctx, north.NotificationsAll, "")
		if err != nil {
			return "", err
		}
		if len(page.Items) == 0 {
			return "No notifications returned", nil
		}
		item := page.Items[0]
		actor := ""
		if len(item.Actors) > 0 {
			actor = " · @" + ui.SafeInline(item.Actors[0].Handle)
		}

		return fmt.Sprintf("%d items · %s%s", len(page.Items), strings.ToLower(string(item.Kind)), actor), nil
	}
}

func bookmarkCheck(api API) func(context.Context) (string, error) {
	service, ok := api.(bookmarkAPI)
	if !ok {
		return unavailable
	}

	return func(ctx context.Context) (string, error) {
		page, _, err := service.Bookmarks(ctx, "")
		if err != nil {
			return "", err
		}

		return postPageSummary(page.Items), nil
	}
}

func messageCheck(api API) func(context.Context) (string, error) {
	service, ok := api.(messageAPI)
	if !ok {
		return unavailable
	}

	return func(ctx context.Context) (string, error) {
		page, _, err := service.DMConversations(ctx, "", false)
		if err != nil {
			return "", err
		}
		if len(page.Items) == 0 {
			return "No conversations returned", nil
		}
		conversation := page.Items[0]
		name := conversationName(conversation)
		preview := "No messages"
		if conversation.LastMessage != nil {
			preview = ui.SafeInline(conversation.LastMessage.Text)
			if preview == "" {
				preview = "Media message"
			}
		}

		return fmt.Sprintf("%d chats · %s: %s", len(page.Items), name, preview), nil
	}
}

func trendCheck(api API) func(context.Context) (string, error) {
	service, ok := api.(trendAPI)
	if !ok {
		return unavailable
	}

	return func(ctx context.Context) (string, error) {
		items, _, err := service.Trends(ctx, "")
		if err != nil {
			return "", err
		}
		if len(items) == 0 {
			return "No trends returned", nil
		}
		trend := items[0]
		tag := ui.SafeInline(trend.Tag)
		if trend.IsHashtag && !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}

		return fmt.Sprintf("%d items · %s · %d posts", len(items), tag, trend.Count), nil
	}
}

func postPageSummary(items []north.Post) string {
	if len(items) == 0 {
		return "No posts returned"
	}
	post := items[0].DisplayPost()
	text := ui.SafeInline(post.Text)
	if text == "" {
		text = "Media post"
	}

	return fmt.Sprintf("%d posts · @%s: %s", len(items), ui.SafeInline(post.Author.Handle), text)
}

func conversationName(conversation north.DMConversation) string {
	if conversation.Name != nil && strings.TrimSpace(*conversation.Name) != "" {
		return ui.SafeInline(*conversation.Name)
	}
	if len(conversation.Participants) > 0 {
		return "@" + ui.SafeInline(conversation.Participants[0].Handle)
	}

	return "Conversation"
}

func unavailable(context.Context) (string, error) {
	return "", errors.New("not available with the current authentication")
}
