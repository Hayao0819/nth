package user

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type profileUpdatedMsg struct {
	target   *Screen
	user     north.User
	response *north.Response
	err      error
}

func (m profileUpdatedMsg) Response() *north.Response { return m.response }

func (d *Screen) editProfile(ctx *reactea.Ctx) tea.Cmd {
	if d.profile == nil || !d.isOwnProfile() || d.profileBusy {
		return nil
	}

	height := 14
	if d.profileMedia {
		height = 18
	}

	return modal.PushAt(ctx, newProfileEditor(d.theme, d.user, d.profileMedia), dialog.Placement(ctx, 66, height))
}

func (d *Screen) saveProfile(ctx context.Context, update domain.ProfileUpdate) tea.Cmd {
	if d.profile == nil || d.profileBusy {
		return nil
	}
	d.profileBusy = true
	d.profileNotice = "Saving profile…"
	d.profileUpdateErr = nil

	return func() tea.Msg {
		user, response, err := d.profile.UpdateProfile(ctx, update)

		return profileUpdatedMsg{target: d, user: user, response: response, err: err}
	}
}

func (d *Screen) applyProfileUpdated(msg profileUpdatedMsg) {
	if msg.target != d {
		return
	}
	d.profileBusy = false
	d.profileUpdateErr = msg.err
	if msg.err != nil {
		d.profileNotice = ui.FriendlyError(msg.err)

		return
	}
	d.user = msg.user
	d.profileLoaded = true
	if d.viewer != nil {
		*d.viewer = msg.user
	}
	d.profileNotice = "Profile updated"
}
