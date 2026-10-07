package account

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

const (
	keywordWordKey          = "word"
	keywordDurationKey      = "duration"
	keywordHomeKey          = "home"
	keywordNotificationsKey = "notifications"
	keywordAnyoneKey        = "anyone"
)

type changedMsg struct {
	target   *Screen
	kind     changeKind
	response *north.Response
	err      error
}

func (m changedMsg) Response() *north.Response { return m.response }

func (s *Screen) change(ctx context.Context, kind changeKind) tea.Cmd {
	if s.acting || s.selected < 0 || s.selected >= s.count() {
		return nil
	}
	s.acting = true
	s.notice = actionProgress(kind)
	var handle, id string
	if s.tab == keywordsTab {
		id = s.words[s.selected].ID
	} else {
		handle = s.users[s.selected].Handle
	}

	return func() tea.Msg {
		var response *north.Response
		var ok bool
		var err error
		switch kind {
		case changeAccept:
			ok, response, err = s.api.AcceptFollowRequest(ctx, handle)
		case changeReject:
			ok, response, err = s.api.RejectFollowRequest(ctx, handle)
		case changeUnblock:
			ok, response, err = s.api.Unblock(ctx, handle)
		case changeUnmute:
			ok, response, err = s.api.Unmute(ctx, handle)
		case changeDeleteKeyword:
			ok, response, err = s.keywords.DeleteMutedKeyword(ctx, id)
		}
		if err == nil && !ok {
			err = errors.New("north did not apply the change")
		}

		return changedMsg{target: s, kind: kind, response: response, err: err}
	}
}

func (s *Screen) confirmChange(ctx *reactea.Ctx, kind changeKind) tea.Cmd {
	if s.acting || s.selected < 0 || s.selected >= s.count() {
		return nil
	}
	title, message, button := "Remove item?", "This item will be removed.", "Remove"
	switch kind {
	case changeReject:
		title, message, button = "Reject follow request?", "The pending follow request will be rejected.", "Reject"
	case changeUnblock:
		title, message, button = "Unblock account?", "This account will be able to interact with you again.", "Unblock"
	case changeUnmute:
		title, message, button = "Unmute account?", "Posts from this account may appear again.", "Unmute"
	case changeDeleteKeyword:
		title, message, button = "Delete muted word?", "This word will no longer be filtered.", "Delete"
	}
	s.confirm = kind

	return modal.PushAt(ctx, dialog.NewConfirm(s.theme, title, message, button), dialog.Placement(ctx, 54, 8))
}

func (s *Screen) openKeywordEditor(ctx *reactea.Ctx) tea.Cmd {
	if s.keywords == nil || s.acting {
		return nil
	}
	form := dialog.NewForm(s.theme, "Mute a word", "Mute",
		dialog.Field{Key: keywordWordKey, Label: "Word or phrase", Placeholder: "Word or phrase", Required: true},
		dialog.Field{Key: keywordDurationKey, Label: "Duration", Placeholder: "forever, 24h, 7d, or 30d", Value: "forever", Required: true},
		dialog.Field{Key: keywordHomeKey, Label: "Home timeline", Kind: dialog.ToggleField, Checked: true},
		dialog.Field{Key: keywordNotificationsKey, Label: "Notifications", Kind: dialog.ToggleField, Checked: true},
		dialog.Field{Key: keywordAnyoneKey, Label: "From anyone", Kind: dialog.ToggleField, Checked: true},
	)

	return modal.PushAt(ctx, form, dialog.Placement(ctx, 64, 16))
}

func (s *Screen) createKeyword(ctx context.Context, result dialog.FormResult) tea.Cmd {
	if s.keywords == nil || s.acting {
		return nil
	}
	duration := north.MuteDuration(strings.ToLower(strings.TrimSpace(result.Values[keywordDurationKey])))
	switch duration {
	case north.MuteForever, north.Mute24Hours, north.Mute7Days, north.Mute30Days:
	default:
		s.notice = "Duration must be forever, 24h, 7d, or 30d"

		return nil
	}
	request := north.CreateMutedKeywordRequest{
		Word:          strings.TrimSpace(result.Values[keywordWordKey]),
		Duration:      duration,
		Home:          result.Toggles[keywordHomeKey],
		Notifications: result.Toggles[keywordNotificationsKey],
		FromAnyone:    result.Toggles[keywordAnyoneKey],
	}
	s.acting = true
	s.notice = "Adding muted word…"

	return func() tea.Msg {
		_, response, err := s.keywords.CreateMutedKeyword(ctx, request)

		return changedMsg{target: s, kind: changeCreateKeyword, response: response, err: err}
	}
}

func (s *Screen) applyChanged(ctx context.Context, msg changedMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}
	s.notice = actionDone(msg.kind)

	return s.load(ctx, false)
}

func actionProgress(kind changeKind) string {
	switch kind {
	case changeAccept:
		return "Accepting request…"
	case changeReject:
		return "Rejecting request…"
	case changeUnblock:
		return "Unblocking account…"
	case changeUnmute:
		return "Unmuting account…"
	case changeDeleteKeyword:
		return "Deleting muted word…"
	default:
		return "Updating…"
	}
}

func actionDone(kind changeKind) string {
	switch kind {
	case changeAccept:
		return "Follow request accepted"
	case changeReject:
		return "Follow request rejected"
	case changeUnblock:
		return "Account unblocked"
	case changeUnmute:
		return "Account unmuted"
	case changeCreateKeyword:
		return "Muted word added"
	case changeDeleteKeyword:
		return "Muted word deleted"
	default:
		return "Updated"
	}
}
