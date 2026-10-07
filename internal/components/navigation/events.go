package navigation

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

type OpenPostMsg struct{ Post north.Post }

type OpenUserMsg struct{ User north.User }

type OpenUserConnectionsMsg struct {
	User      north.User
	Following bool
}

type OpenListMsg struct{ List north.List }

type OpenListMembersMsg struct{ List north.List }

type OpenBookmarkFoldersMsg struct{}

type OpenBookmarkFolderMsg struct{ Folder north.BookmarkFolder }

type OpenSavedPostsMsg struct{}

type OpenAccountSafetyMsg struct{}

type ProfileUpdatedMsg struct{ User north.User }

type MessagesReadMsg struct{}

type PostActivity uint8

const (
	PostQuotes PostActivity = iota
	PostLikes
	PostReposts
	PostHistory
)

type OpenPostActivityMsg struct {
	Post     north.Post
	Activity PostActivity
}

func OpenPost(post north.Post) tea.Cmd {
	return func() tea.Msg { return OpenPostMsg{Post: post} }
}

func OpenUser(user north.User) tea.Cmd {
	return func() tea.Msg { return OpenUserMsg{User: user} }
}

func OpenFollowers(user north.User) tea.Cmd {
	return func() tea.Msg { return OpenUserConnectionsMsg{User: user} }
}

func OpenFollowing(user north.User) tea.Cmd {
	return func() tea.Msg { return OpenUserConnectionsMsg{User: user, Following: true} }
}

func OpenList(list north.List) tea.Cmd {
	return func() tea.Msg { return OpenListMsg{List: list} }
}

func OpenListMembers(list north.List) tea.Cmd {
	return func() tea.Msg { return OpenListMembersMsg{List: list} }
}

func OpenBookmarkFolders() tea.Cmd {
	return func() tea.Msg { return OpenBookmarkFoldersMsg{} }
}

func OpenBookmarkFolder(folder north.BookmarkFolder) tea.Cmd {
	return func() tea.Msg { return OpenBookmarkFolderMsg{Folder: folder} }
}

func OpenSavedPosts() tea.Cmd {
	return func() tea.Msg { return OpenSavedPostsMsg{} }
}

func OpenAccountSafety() tea.Cmd {
	return func() tea.Msg { return OpenAccountSafetyMsg{} }
}

func ProfileUpdated(user north.User) tea.Cmd {
	return func() tea.Msg { return ProfileUpdatedMsg{User: user} }
}

func MessagesRead() tea.Cmd {
	return func() tea.Msg { return MessagesReadMsg{} }
}

func OpenPostActivity(post north.Post, activity PostActivity) tea.Cmd {
	return func() tea.Msg { return OpenPostActivityMsg{Post: post, Activity: activity} }
}
