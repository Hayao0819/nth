package dialog

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
	"github.com/mattn/go-runewidth"
)

type Submission struct {
	Text     string
	ReplyTo  string
	QuoteID  string
	EditID   string
	MediaIDs []string
	Canceled bool
}

type Compose struct {
	reactea.Wrapper

	theme    ui.Theme
	input    *reactea.ReactifiedWidget[textarea.Model]
	replyTo  *north.Post
	quote    *north.Post
	problem  string
	restored bool
	editID   string
	mediaIDs []string
}

func NewEdit(theme ui.Theme, post north.Post, draft string) *Compose {
	text := draft
	if text == "" {
		text = post.Text
	}
	compose := NewCompose(theme, nil, nil, text)
	compose.editID = post.ID
	compose.restored = draft != "" && draft != post.Text
	for _, media := range post.Media {
		if media.ID != "" {
			compose.mediaIDs = append(compose.mediaIDs, media.ID)
		}
	}

	return compose
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
	}
}

func (d *Compose) Init(ctx *reactea.Ctx) tea.Cmd {
	return tea.Batch(d.Wrapper.Init(ctx), d.input.Widget.Focus())
}

func (d *Compose) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && y == ctx.Height()-d.theme.Dialog.GetBorderBottomSize()-1 {
			innerWidth := max(1, ctx.Width()-d.theme.Dialog.GetHorizontalFrameSize())
			buttonWidth := lipgloss.Width(d.theme.Button.Padding(0, 1).Render("Post"))
			buttonLeft := d.theme.Dialog.GetBorderLeftSize() + innerWidth - buttonWidth
			if x >= buttonLeft && x < d.theme.Dialog.GetBorderLeftSize()+innerWidth {
				return d.send(ctx)
			}
			if x < d.theme.Dialog.GetBorderLeftSize()+12 {
				return d.cancel(ctx)
			}
		}
	}

	switch {
	case reactea.Key(msg, "esc"):
		return d.cancel(ctx)
	case reactea.Key(msg, "ctrl+s", "ctrl+enter", "ctrl+j", "ctrl+m", "alt+enter"):
		return d.send(ctx)
	}

	d.problem = ""

	return d.Wrapper.Update(ctx, msg)
}

func (d *Compose) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	style := d.theme.Dialog
	innerWidth := max(0, width-style.GetHorizontalFrameSize())
	innerHeight := max(0, height-style.GetVerticalFrameSize())
	inputHeight := max(1, innerHeight-2)

	title := "New post"
	target := d.replyTo
	if d.editID != "" {
		title = "Edit post"
	} else if target != nil {
		title = "Reply to @" + ui.SafeInline(target.Author.Handle)
	} else if d.quote != nil {
		target = d.quote
		title = "Quote @" + ui.SafeInline(target.Author.Handle)
	}
	if d.restored {
		title += " · Draft"
	}
	contextLine := ""
	inputTop := 1
	if target != nil {
		inputHeight = max(1, innerHeight-4)
		inputTop = 3
		preview := ui.SafeInline(target.Text)
		if target.Deleted {
			preview = "Post deleted"
		} else if target.HiddenReason != "" {
			preview = "Hidden: " + string(target.HiddenReason)
		} else if target.Unavailable {
			preview = "Post unavailable"
		}
		if preview == "" && len(target.Media) > 0 {
			preview = "[media]"
		}
		author := ui.SafeInline(target.Author.Name) + "  @" + ui.SafeInline(target.Author.Handle)
		contextLine = d.theme.Dim.Render(ui.Clip(author, innerWidth)) + "\n" +
			d.theme.Dim.Render(ui.Clip(preview, innerWidth)) + "\n"
	}

	inputCtx := ctx.Inset(
		style.GetBorderLeftSize(),
		style.GetBorderTopSize()+inputTop,
		innerWidth,
		inputHeight,
	)
	content := d.theme.ModalTitle.Render(ui.Clip(title, max(1, innerWidth-2))) + "\n" + contextLine +
		ui.Fit(d.Wrapper.Render(inputCtx), innerWidth, inputHeight) + "\n"

	length := northTextLength(d.input.Widget.Value())
	count := fmt.Sprintf("%d/280", length)
	if length > 280 {
		count = d.theme.Bad.Render(count)
	}
	buttonStyle := d.theme.Button
	if length == 0 || length > 280 {
		buttonStyle = d.theme.Dim
	}
	buttonLabel := "Post"
	if d.editID != "" {
		buttonLabel = "Save"
	}
	button := buttonStyle.Padding(0, 1).Render(buttonLabel)
	right := count + "  " + button
	hint := d.theme.Dim.Render("Esc  Close · Ctrl+S sends")
	if d.problem != "" {
		hint = d.theme.Bad.Render(d.problem)
	}
	content += ui.Sides(hint, right, innerWidth)

	return RenderDialog(style, content, width, height)
}

func (d *Compose) send(ctx *reactea.Ctx) tea.Cmd {
	text := d.input.Widget.Value()
	length := northTextLength(text)
	switch {
	case strings.TrimSpace(text) == "":
		d.problem = "Write something before posting"
	case length > 280:
		d.problem = fmt.Sprintf("Post is %d units; the limit is 280", length)
	default:
		d.problem = ""

		return modal.Return(ctx, Submission{
			Text:     text,
			ReplyTo:  postID(d.replyTo),
			QuoteID:  postID(d.quote),
			EditID:   d.editID,
			MediaIDs: append([]string(nil), d.mediaIDs...),
		})
	}

	return nil
}

func (d *Compose) cancel(ctx *reactea.Ctx) tea.Cmd {
	return modal.Return(ctx, Submission{
		Text:     d.input.Widget.Value(),
		ReplyTo:  postID(d.replyTo),
		QuoteID:  postID(d.quote),
		EditID:   d.editID,
		MediaIDs: append([]string(nil), d.mediaIDs...),
		Canceled: true,
	})
}

func postID(post *north.Post) string {
	if post == nil {
		return ""
	}

	return post.ID
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
