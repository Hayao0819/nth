package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/components/sessiondialog"
	"github.com/Hayao0819/nth/internal/domain"
	notificationfeature "github.com/Hayao0819/nth/internal/features/notification"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

func (r *root) handleEvent(ctx *reactea.Ctx, event tea.Msg) (tea.Cmd, bool) {
	switch msg := event.(type) {
	case *domain.RefreshRequest:
		if msg == nil {
			return nil, true
		}

		return modal.Push(ctx, sessiondialog.New(r.theme, msg)), true

	case notificationfeature.PageLoadedMsg:
		return r.Wrapper.Update(ctx, msg), true

	case notificationfeature.ReadMsg:
		if msg.Error() == nil {
			r.unread = 0
			r.unreadErr = nil
		}

		return r.Wrapper.Update(ctx, msg), true

	case pageheader.BackMsg:
		return r.backPage(ctx), true

	case navigation.OpenPostMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openPost(ctx, msg.Post), true

	case navigation.OpenPostActivityMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openPostActivity(ctx, msg.Post, msg.Activity), true

	case navigation.OpenUserMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openUser(ctx, msg.User), true

	case navigation.OpenUserConnectionsMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openUserConnections(ctx, msg.User, msg.Following), true

	case navigation.OpenListMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openList(ctx, msg.List), true

	case navigation.OpenListMembersMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openListMembers(ctx, msg.List), true

	case navigation.OpenBookmarkFoldersMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openBookmarkFolders(ctx), true

	case navigation.OpenBookmarkFolderMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.openBookmarkFolder(ctx, msg.Folder), true

	case navigation.OpenSavedPostsMsg:
		return r.openSavedPosts(ctx), true

	case navigation.OpenAccountSafetyMsg:
		return r.openAccountSafety(ctx), true

	case navigation.OpenBrowserMsg:
		return r.openBrowser(msg.URL), true

	case navigation.BrowserOpenedMsg:
		return r.handleBrowserOpened(ctx, msg), true

	case navigation.ProfileUpdatedMsg:
		user := msg.User
		r.me = &user
		r.users[strings.ToLower(user.Handle)] = user

		return r.Wrapper.Update(ctx, msg), true

	case navigation.MessagesReadMsg:
		return r.loadDMUnread(ctx), true

	case postcomponent.ActionMsg:
		if r.timelineInputCaptured(ctx) {
			return nil, true
		}

		return r.handlePostAction(ctx, msg.Action, msg.Post), true

	case postCreatedMsg:
		return r.handlePostCreated(ctx, msg), true

	case postDeletedMsg:
		return r.handlePostDeleted(ctx, msg), true

	case postEditedMsg:
		return r.handlePostEdited(ctx, msg), true

	case bookmarkChangedMsg:
		return r.handleBookmarkChanged(ctx, msg), true

	case modal.Result[dialog.Submission]:
		return r.handleSubmission(ctx, msg), true

	case navigationMsg:
		return r.handleNavigation(ctx, msg), true
	}

	return nil, false
}

func (r *root) timelineInputCaptured(ctx *reactea.Ctx) bool {
	return r.page.kind == timelinePage && ctx.InputCaptured()
}

func (r *root) handleNavigation(ctx *reactea.Ctx, msg navigationMsg) tea.Cmd {
	if ctx.InputCaptured() {
		return nil
	}

	switch msg.action {
	case navigateSearch:
		if msg.query != "" {
			return r.searchFor(ctx, msg.query)
		}

		return r.search(ctx)
	case navigateCompose:
		return r.compose(ctx, nil, nil)
	case navigateNotifications:
		return r.openNotifications(ctx)
	case navigateMessages:
		return r.openMessages(ctx)
	case navigateBookmarks:
		return r.openBookmarks(ctx)
	case navigateLists:
		return r.openLists(ctx)
	case navigateProfile:
		if r.me != nil {
			return r.openUser(ctx, *r.me)
		}

		return nil
	default:
		return r.openTimeline(ctx, msg.mode)
	}
}
