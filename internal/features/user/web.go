package user

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/nth/internal/components/navigation"
)

func (d *Screen) browserURL() string {
	return navigation.UserWebURL(d.user)
}

func (d *Screen) browserAction() string {
	if d.browserURL() == "" {
		return ""
	}

	return navigation.BrowserActionLabel
}

func (d *Screen) openBrowser() tea.Cmd {
	return navigation.OpenBrowser(d.browserURL())
}
