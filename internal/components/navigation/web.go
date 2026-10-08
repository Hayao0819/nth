package navigation

import (
	"net/url"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
)

const (
	BrowserActionLabel = "w  Browser ↗"
	northWebOrigin     = "https://north.rip"
)

type OpenBrowserMsg struct{ URL string }

type BrowserOpenedMsg struct {
	URL string
	Err error
}

func OpenBrowser(rawURL string) tea.Cmd {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil
	}

	return func() tea.Msg { return OpenBrowserMsg{URL: rawURL} }
}

func UserWebURL(user north.User) string {
	handle := strings.TrimPrefix(strings.TrimSpace(user.Handle), "@")
	if handle == "" {
		return ""
	}

	return northWebOrigin + "/" + url.PathEscape(handle)
}

func PostWebURL(post north.Post) string {
	target := post.DisplayPost()
	if target == nil {
		return ""
	}
	handle := strings.TrimPrefix(strings.TrimSpace(target.Author.Handle), "@")
	id := strings.TrimSpace(target.ID)
	if handle == "" || id == "" {
		return ""
	}

	return northWebOrigin + "/" + url.PathEscape(handle) + "/status/" + url.PathEscape(id)
}
