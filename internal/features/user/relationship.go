package user

import (
	"context"
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type relationshipAPI interface {
	Follow(context.Context, string) (bool, *north.Response, error)
	Unfollow(context.Context, string) (bool, *north.Response, error)
	Block(context.Context, string) (bool, *north.Response, error)
	Unblock(context.Context, string) (bool, *north.Response, error)
	Mute(context.Context, string) (bool, *north.Response, error)
	Unmute(context.Context, string) (bool, *north.Response, error)
}

type postNotificationAPI interface {
	EnablePostNotifications(context.Context, string) (bool, *north.Response, error)
	DisablePostNotifications(context.Context, string) (bool, *north.Response, error)
}

type relationshipKind uint8

const (
	noRelationship relationshipKind = iota
	followRelationship
	blockRelationship
	muteRelationship
	postNotificationRelationship
)

type relationshipChange struct {
	kind   relationshipKind
	enable bool
}

type relationshipResultMsg struct {
	target *Screen
	change relationshipChange
	resp   *north.Response
	err    error
}

func (m relationshipResultMsg) Response() *north.Response { return m.resp }

func (d *Screen) relationshipKinds() []relationshipKind {
	if !d.profileLoaded || d.viewer == nil || d.isOwnProfile() || strings.TrimPrefix(strings.TrimSpace(d.user.Handle), "@") == "" {
		return nil
	}
	kinds := make([]relationshipKind, 0, 4)
	if d.relationships != nil {
		kinds = append(kinds, followRelationship, blockRelationship, muteRelationship)
	}
	if d.postNotices != nil {
		kinds = append(kinds, postNotificationRelationship)
	}

	return kinds
}

func (d *Screen) isOwnProfile() bool {
	if d.viewer == nil {
		return false
	}
	if d.viewer.ID != "" && d.user.ID != "" {
		return d.viewer.ID == d.user.ID
	}

	return strings.EqualFold(
		strings.TrimPrefix(strings.TrimSpace(d.viewer.Handle), "@"),
		strings.TrimPrefix(strings.TrimSpace(d.user.Handle), "@"),
	)
}

func (d *Screen) requestRelationship(ctx *reactea.Ctx, kind relationshipKind) tea.Cmd {
	if !d.canManageRelationship(kind) || d.relationshipBusy != noRelationship || d.confirmingBlock {
		return nil
	}
	change := d.relationshipChange(kind)
	if change.kind == noRelationship {
		return nil
	}
	if change.kind == followRelationship && change.enable && d.user.Blocking {
		d.relationshipErr = errors.New("unblock this account before following it")
		d.relationshipNotice = ""

		return nil
	}
	if change.kind == blockRelationship && change.enable {
		d.confirmingBlock = true
		handle := strings.TrimPrefix(ui.SafeInline(d.user.Handle), "@")

		return modal.PushAt(
			ctx,
			dialog.NewConfirm(
				d.theme,
				"Block @"+handle+"?",
				"Blocking removes follows in both directions.",
				"Block",
			),
			dialog.Placement(ctx, 48, 7),
		)
	}

	return d.changeRelationship(ctx.Context(), change)
}

func (d *Screen) canManageRelationship(kind relationshipKind) bool {
	for _, available := range d.relationshipKinds() {
		if available == kind {
			return true
		}
	}

	return false
}

func (d *Screen) relationshipChange(kind relationshipKind) relationshipChange {
	switch kind {
	case followRelationship:
		return relationshipChange{kind: kind, enable: !d.user.Following && !d.followRequested}
	case blockRelationship:
		return relationshipChange{kind: kind, enable: !d.user.Blocking}
	case muteRelationship:
		return relationshipChange{kind: kind, enable: !d.user.Muting}
	case postNotificationRelationship:
		return relationshipChange{kind: kind, enable: !d.postNotifications}
	default:
		return relationshipChange{}
	}
}

func (d *Screen) changeRelationship(ctx context.Context, change relationshipChange) tea.Cmd {
	if !d.canManageRelationship(change.kind) || change.kind == noRelationship || d.relationshipBusy != noRelationship {
		return nil
	}
	d.relationshipBusy = change.kind
	d.relationshipErr = nil
	d.relationshipNotice = ""
	handle := strings.TrimPrefix(strings.TrimSpace(d.user.Handle), "@")

	return func() tea.Msg {
		var (
			ok       bool
			response *north.Response
			err      error
		)
		switch change.kind {
		case followRelationship:
			if change.enable {
				ok, response, err = d.relationships.Follow(ctx, handle)
			} else {
				ok, response, err = d.relationships.Unfollow(ctx, handle)
			}
		case blockRelationship:
			if change.enable {
				ok, response, err = d.relationships.Block(ctx, handle)
			} else {
				ok, response, err = d.relationships.Unblock(ctx, handle)
			}
		case muteRelationship:
			if change.enable {
				ok, response, err = d.relationships.Mute(ctx, handle)
			} else {
				ok, response, err = d.relationships.Unmute(ctx, handle)
			}
		case postNotificationRelationship:
			if change.enable {
				ok, response, err = d.postNotices.EnablePostNotifications(ctx, handle)
			} else {
				ok, response, err = d.postNotices.DisablePostNotifications(ctx, handle)
			}
		}
		if err == nil && !ok {
			err = errors.New("north did not apply the relationship change")
		}

		return relationshipResultMsg{target: d, change: change, resp: response, err: err}
	}
}

func (d *Screen) applyRelationshipResult(result relationshipResultMsg) {
	d.relationshipBusy = noRelationship
	d.relationshipErr = result.err
	if result.err != nil {
		return
	}

	switch result.change.kind {
	case followRelationship:
		d.applyFollow(result.change.enable)
	case blockRelationship:
		d.applyBlock(result.change.enable)
	case muteRelationship:
		d.applyMute(result.change.enable)
	case postNotificationRelationship:
		d.postNotifications = result.change.enable
		if result.change.enable {
			d.relationshipNotice = "Post notifications enabled"
		} else {
			d.relationshipNotice = "Post notifications disabled"
		}
	}
}

func (d *Screen) applyFollow(enable bool) {
	if !enable {
		requested := d.followRequested
		if d.user.Following && d.user.FollowerCount > 0 {
			d.user.FollowerCount--
		}
		d.user.Following = false
		d.followRequested = false
		if requested {
			d.relationshipNotice = "Follow request canceled"
		} else {
			d.relationshipNotice = "Unfollowed @" + strings.TrimPrefix(ui.SafeInline(d.user.Handle), "@")
		}

		return
	}
	if d.user.Protected {
		d.followRequested = true
		d.relationshipNotice = "Follow request sent"

		return
	}
	if !d.user.Following {
		d.user.FollowerCount++
	}
	d.user.Following = true
	d.relationshipNotice = "Following @" + strings.TrimPrefix(ui.SafeInline(d.user.Handle), "@")
}

func (d *Screen) applyBlock(enable bool) {
	if !enable {
		d.user.Blocking = false
		d.relationshipNotice = "Account unblocked"

		return
	}
	if d.user.Following && d.user.FollowerCount > 0 {
		d.user.FollowerCount--
	}
	if d.user.FollowedBy && d.user.FollowingCount > 0 {
		d.user.FollowingCount--
	}
	d.user.Following = false
	d.user.FollowedBy = false
	d.user.Blocking = true
	d.followRequested = false
	d.relationshipNotice = "Account blocked"
}

func (d *Screen) applyMute(enable bool) {
	d.user.Muting = enable
	if enable {
		d.relationshipNotice = "Account muted"
	} else {
		d.relationshipNotice = "Account unmuted"
	}
}

func (d *Screen) relationshipRow(width int, kinds []relationshipKind) string {
	labels := make([]string, len(kinds))
	for index, kind := range kinds {
		switch kind {
		case followRelationship:
			labels[index] = "F  Follow"
			if d.followRequested {
				labels[index] = "F  Requested"
			} else if d.user.Following {
				labels[index] = "F  Following"
			}
		case blockRelationship:
			labels[index] = "B  Block"
			if d.user.Blocking {
				labels[index] = "B  Unblock"
			}
		case muteRelationship:
			labels[index] = "M  Mute"
			if d.user.Muting {
				labels[index] = "M  Unmute"
			}
		case postNotificationRelationship:
			labels[index] = "N  Notify"
			if d.postNotifications {
				labels[index] = "N  Notifying"
			}
		}
		style := d.theme.Heading
		switch {
		case d.relationshipBusy != noRelationship && d.relationshipBusy != kind:
			style = d.theme.Dim
		case kind == followRelationship && d.user.Blocking:
			style = d.theme.Dim
		case kind == followRelationship && (d.followRequested || d.user.Following):
			style = d.theme.Active
		case kind == blockRelationship && d.user.Blocking:
			style = d.theme.Bad.Bold(true)
		case kind == muteRelationship && d.user.Muting:
			style = d.theme.Warn.Bold(true)
		case kind == postNotificationRelationship && d.postNotifications:
			style = d.theme.Active
		}
		labels[index] = style.Render(labels[index])
	}

	return ui.Columns(labels, width)
}

func (d *Screen) relationshipStatus() string {
	switch {
	case d.relationshipErr != nil:
		return d.theme.Warn.Render("! " + ui.FriendlyError(d.relationshipErr))
	case d.relationshipBusy != noRelationship:
		return d.theme.Dim.Render("Updating account…")
	case d.relationshipNotice != "":
		return d.theme.Active.Render("✓ " + d.relationshipNotice)
	default:
		return ""
	}
}
