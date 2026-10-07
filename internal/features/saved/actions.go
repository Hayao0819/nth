package saved

import (
	"context"
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/dialog"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

const (
	scheduleTextKey = "text"
	scheduleTimeKey = "time"
)

type actionKind uint8

const (
	actionSaveDraft actionKind = iota
	actionSaveScheduled
	actionCreatePending
	actionDelete
	actionSendPending
	actionUndoPending
)

type actionMsg struct {
	target   *Screen
	action   actionKind
	response *north.Response
	err      error
}

func (m actionMsg) Response() *north.Response { return m.response }

func (s *Screen) openEditor(ctx *reactea.Ctx, create bool) tea.Cmd {
	if s.api == nil || s.acting {
		return nil
	}
	initial, id := "", ""
	if !create {
		if s.selected < 0 || s.selected >= s.count() || s.tab == pendingTab {
			return nil
		}
		draft := s.draftAt(s.selected)
		id = draft.ID
		if len(draft.Items) > 0 {
			initial = draft.Items[0].Text
		}
	}
	s.editingID = id
	switch s.tab {
	case scheduledTab:
		s.editor = editorScheduled
		when := time.Now().Add(time.Hour).Local()
		if !create {
			when = s.scheduled[s.selected].ScheduledAt.Local()
		}
		form := dialog.NewForm(s.theme, "Schedule post", "Save",
			dialog.Field{Key: scheduleTextKey, Label: "Post", Placeholder: "What is happening?", Value: initial, Required: true},
			dialog.Field{Key: scheduleTimeKey, Label: "Publish at", Placeholder: "YYYY-MM-DD HH:MM", Value: when.Format("2006-01-02 15:04"), Required: true},
		)

		return modal.PushAt(ctx, form, dialog.Placement(ctx, 66, 11))
	case pendingTab:
		if !create {
			return nil
		}
		s.editor = editorPending

		return modal.PushAt(ctx, dialog.NewTextEditor(s.theme, "Add to undo queue", "What is happening?", "", "Queue"), dialog.Placement(ctx, 68, 14))
	default:
		s.editor = editorDraft

		return modal.PushAt(ctx, dialog.NewTextEditor(s.theme, "Draft", "What is happening?", initial, "Save"), dialog.Placement(ctx, 68, 14))
	}
}

func (s *Screen) saveText(ctx context.Context, text string) tea.Cmd {
	kind, id := s.editor, s.editingID
	s.editor, s.editingID = editorNone, ""
	if s.api == nil || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Saving…"

	return func() tea.Msg {
		message := actionMsg{target: s}
		switch kind {
		case editorDraft:
			request := draftRequest(s.draftByID(id), text)
			if id == "" {
				_, message.response, message.err = s.api.CreateDraft(ctx, request)
			} else {
				_, message.response, message.err = s.api.UpdateDraft(ctx, id, request)
			}
			message.action = actionSaveDraft
		case editorPending:
			_, message.response, message.err = s.api.CreatePendingPost(ctx, north.CreatePendingPostRequest{
				Item: north.DraftItemInput{Text: text},
			})
			message.action = actionCreatePending
		default:
			message.err = errors.New("saved post editor is no longer active")
		}

		return message
	}
}

func (s *Screen) saveSchedule(ctx context.Context, result dialog.FormResult) tea.Cmd {
	id := s.editingID
	s.editor, s.editingID = editorNone, ""
	when, err := time.ParseInLocation("2006-01-02 15:04", strings.TrimSpace(result.Values[scheduleTimeKey]), time.Local)
	if err != nil {
		s.notice = "Use YYYY-MM-DD HH:MM for the publish time"

		return nil
	}
	if s.api == nil || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Saving schedule…"
	request := north.SchedulePostRequest{
		Item:        north.DraftItemInput{Text: strings.TrimSpace(result.Values[scheduleTextKey])},
		ScheduledAt: when,
	}
	if existing := s.scheduledByID(id); existing != nil {
		request.Item = draftItemInput(existing.Draft, request.Item.Text)
		request.QuotedID, request.InReplyToID = draftContext(existing.Draft)
	}

	return func() tea.Msg {
		var response *north.Response
		var err error
		if id == "" {
			_, response, err = s.api.SchedulePost(ctx, request)
		} else {
			_, response, err = s.api.UpdateScheduledPost(ctx, id, request)
		}

		return actionMsg{target: s, action: actionSaveScheduled, response: response, err: err}
	}
}

func (s *Screen) confirmDelete(ctx *reactea.Ctx) tea.Cmd {
	if s.tab == pendingTab || s.selected < 0 || s.selected >= s.count() || s.acting {
		return nil
	}
	s.confirm = true
	s.deleteID = s.draftAt(s.selected).ID
	s.deleteTab = s.tab

	return modal.PushAt(ctx, dialog.NewConfirm(s.theme, "Delete saved post?", "This cannot be undone.", "Delete"), dialog.Placement(ctx, 52, 8))
}

func (s *Screen) delete(ctx context.Context) tea.Cmd {
	id, source := s.deleteID, s.deleteTab
	s.deleteID = ""
	if id == "" || s.api == nil || s.acting {
		return nil
	}
	s.acting = true
	s.notice = "Deleting…"

	return func() tea.Msg {
		var response *north.Response
		var err error
		if source == scheduledTab {
			_, response, err = s.api.DeleteScheduledPosts(ctx, id)
		} else {
			_, response, err = s.api.DeleteDrafts(ctx, id)
		}

		return actionMsg{target: s, action: actionDelete, response: response, err: err}
	}
}

func (s *Screen) pendingAction(ctx context.Context, undo bool) tea.Cmd {
	if s.tab != pendingTab || s.selected < 0 || s.selected >= len(s.pending) || s.acting {
		return nil
	}
	id := s.pending[s.selected].ID
	s.acting = true
	if undo {
		s.notice = "Moving to drafts…"
	} else {
		s.notice = "Sending…"
	}

	return func() tea.Msg {
		message := actionMsg{target: s, action: actionSendPending}
		if undo {
			message.action = actionUndoPending
			_, message.response, message.err = s.api.UndoPendingPost(ctx, id)
		} else {
			_, message.response, message.err = s.api.SendPendingPost(ctx, id)
		}

		return message
	}
}

func (s *Screen) applyAction(ctx context.Context, msg actionMsg) tea.Cmd {
	if msg.target != s {
		return nil
	}
	s.acting = false
	if msg.err != nil {
		s.notice = ui.FriendlyError(msg.err)

		return nil
	}
	s.notice = map[actionKind]string{
		actionSaveDraft:     "Draft saved",
		actionSaveScheduled: "Schedule saved",
		actionCreatePending: "Added to undo queue",
		actionDelete:        "Saved post deleted",
		actionSendPending:   "Post sent",
		actionUndoPending:   "Moved to drafts",
	}[msg.action]

	return s.load(ctx)
}

func draftRequest(existing *north.Draft, text string) north.DraftRequest {
	request := north.DraftRequest{Items: []north.DraftItemInput{{Text: text}}}
	if existing != nil {
		request.Items[0] = draftItemInput(*existing, text)
		request.QuotedID, request.InReplyToID = draftContext(*existing)
	}

	return request
}

func draftItemInput(draft north.Draft, text string) north.DraftItemInput {
	item := north.DraftItemInput{Text: text}
	if len(draft.Items) == 0 {
		return item
	}
	item.ReplyPolicy = draft.Items[0].ReplyPolicy
	for _, media := range draft.Items[0].Media {
		if media.ID != "" {
			item.MediaIDs = append(item.MediaIDs, media.ID)
		}
	}

	return item
}

func draftContext(draft north.Draft) (quoted, reply *north.NullableString) {
	if draft.QuotedPost != nil && draft.QuotedPost.ID != "" {
		quoted = north.NewNullableString(draft.QuotedPost.ID)
	}
	if draft.InReplyToPost != nil && draft.InReplyToPost.ID != "" {
		reply = north.NewNullableString(draft.InReplyToPost.ID)
	}

	return quoted, reply
}

func (s *Screen) draftByID(id string) *north.Draft {
	for index := range s.drafts {
		if s.drafts[index].ID == id {
			return &s.drafts[index]
		}
	}

	return nil
}

func (s *Screen) scheduledByID(id string) *north.ScheduledPost {
	for index := range s.scheduled {
		if s.scheduled[index].ID == id {
			return &s.scheduled[index]
		}
	}

	return nil
}
