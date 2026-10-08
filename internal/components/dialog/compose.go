package dialog

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/navigation"
	"github.com/Hayao0819/nth/internal/domain"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/mattn/go-runewidth"
)

type Submission struct {
	Text        string
	ReplyTo     string
	QuoteID     string
	EditID      string
	EditETag    string
	EditBase    north.Post
	MediaIDs    []string
	Poll        *north.CreatePoll
	ThreadItems []north.ThreadItem
	Canceled    bool
}

type Compose struct {
	reactea.Wrapper

	theme       ui.Theme
	input       *reactea.ReactifiedWidget[textarea.Model]
	replyTo     *north.Post
	quote       *north.Post
	problem     string
	restored    bool
	editID      string
	editETag    string
	editBase    north.Post
	mediaIDs    []string
	media       []north.Media
	mediaAPI    domain.MediaAPI
	mediaPrompt mediaPromptKind
	mediaBusy   bool
	mediaNotice string
	uploaded    map[string]bool
	poll        *north.CreatePoll
	pollForm    bool
	thread      []north.ThreadItem
	threadable  bool
}

func NewEdit(theme ui.Theme, post north.Post) *Compose {
	compose := NewCompose(theme, nil, nil, post.Text)
	compose.editID = post.ID
	compose.editBase = cloneEditBase(post)
	compose.restored = false
	for _, media := range post.Media {
		if media.ID != "" {
			compose.mediaIDs = append(compose.mediaIDs, media.ID)
			compose.media = append(compose.media, media)
		}
	}

	return compose
}

func (d *Compose) RestoreEditDraft(text string, mediaIDs []string, base north.Post) *Compose {
	d.input.Widget.SetValue(text)
	d.mediaIDs = append([]string(nil), mediaIDs...)
	if base.ID != "" {
		d.editBase = cloneEditBase(base)
	}
	d.restored = true

	return d
}

func (d *Compose) SetEditETag(etag string) *Compose {
	d.editETag = etag

	return d
}

func NewCompose(theme ui.Theme, replyTo, quote *north.Post, draft string) *Compose {
	model := textarea.New()
	model.Prompt = ""
	model.Placeholder = "What is happening?"
	model.ShowLineNumbers = false
	model.CharLimit = 0
	model.SetVirtualCursor(false)
	styles := model.Styles()
	styles.Focused.CursorLine = styles.Focused.CursorLine.UnsetBackground()
	model.SetStyles(styles)
	model.SetValue(draft)
	model.Focus()

	widget := reactea.ReactifyWidget(model).OnResize(func(model textarea.Model, width, height int) textarea.Model {
		model.SetWidth(max(1, width))
		model.SetHeight(max(1, height))

		return model
	})

	return &Compose{
		Wrapper:  reactea.Wrap(widget),
		theme:    theme,
		input:    widget,
		replyTo:  replyTo,
		quote:    quote,
		restored: strings.TrimSpace(draft) != "",
		uploaded: make(map[string]bool),
	}
}

func (d *Compose) SetMediaAPI(api domain.MediaAPI) *Compose {
	d.mediaAPI = api

	return d
}

func (d *Compose) RestoreMediaIDs(ids []string) *Compose {
	d.mediaIDs = append([]string(nil), ids...)
	d.restored = d.restored || len(ids) > 0

	return d
}

func (d *Compose) RestorePoll(poll *north.CreatePoll) *Compose {
	d.poll = clonePoll(poll)
	d.restored = d.restored || poll != nil

	return d
}

func (d *Compose) RestoreThread(items []north.ThreadItem) *Compose {
	d.thread = cloneThreadItems(items)
	d.restored = d.restored || len(items) > 0

	return d
}

func (d *Compose) SetThreadEnabled(enabled bool) *Compose {
	d.threadable = enabled && d.replyTo == nil && d.quote == nil && d.editID == ""

	return d
}

func (d *Compose) Init(ctx *reactea.Ctx) tea.Cmd {
	return tea.Batch(d.Wrapper.Init(ctx), d.input.Widget.Focus())
}

func (d *Compose) openSaved(ctx *reactea.Ctx) tea.Cmd {
	return tea.Sequence(modal.Dismiss(ctx), navigation.OpenSavedPosts())
}

func (d *Compose) send(ctx *reactea.Ctx) tea.Cmd {
	text := d.input.Widget.Value()
	length := northTextLength(text)
	switch {
	case !d.canSend():
		d.problem = "Write something or attach media before posting"
	case length > 280:
		d.problem = fmt.Sprintf("Post is %d units; the limit is 280", length)
	default:
		d.problem = ""

		submission := Submission{
			Text:     text,
			ReplyTo:  postID(d.replyTo),
			QuoteID:  postID(d.quote),
			EditID:   d.editID,
			EditETag: d.editETag,
			EditBase: cloneEditBase(d.editBase),
			MediaIDs: append([]string(nil), d.mediaIDs...),
			Poll:     clonePoll(d.poll),
		}
		if len(d.thread) > 0 {
			submission.ThreadItems = cloneThreadItems(d.thread)
			if !currentThreadItemEmpty(text, d.mediaIDs, d.poll) {
				submission.ThreadItems = append(submission.ThreadItems, d.currentThreadItem())
			}
		}

		return modal.Return(ctx, submission)
	}

	return nil
}

func (d *Compose) cancel(ctx *reactea.Ctx) tea.Cmd {
	return modal.Return(ctx, Submission{
		Text:        d.input.Widget.Value(),
		ReplyTo:     postID(d.replyTo),
		QuoteID:     postID(d.quote),
		EditID:      d.editID,
		EditETag:    d.editETag,
		EditBase:    cloneEditBase(d.editBase),
		MediaIDs:    append([]string(nil), d.mediaIDs...),
		Poll:        clonePoll(d.poll),
		ThreadItems: cloneThreadItems(d.thread),
		Canceled:    true,
	})
}

func (d *Compose) canSend() bool {
	return !currentThreadItemEmpty(d.input.Widget.Value(), d.mediaIDs, d.poll) || len(d.thread) > 0
}

func postID(post *north.Post) string {
	if post == nil {
		return ""
	}

	return post.ID
}

func cloneEditBase(post north.Post) north.Post {
	post.Media = append([]north.Media(nil), post.Media...)

	return post
}

func northTextLength(value string) int {
	length := 0
	for _, r := range value {
		width := runewidth.RuneWidth(r)
		if width < 1 {
			width = 1
		}
		length += min(width, 2)
	}

	return length
}

func TextLength(value string) int { return northTextLength(value) }
