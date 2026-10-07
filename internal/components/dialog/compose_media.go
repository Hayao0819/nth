package dialog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/modal"
)

type mediaPromptKind uint8

const (
	mediaPromptNone mediaPromptKind = iota
	mediaPromptPath
	mediaPromptAlt
)

type mediaAction uint8

const (
	mediaUploaded mediaAction = iota
	mediaDeleted
	mediaAltUpdated
	mediaWarningUpdated
)

type mediaUpdateMsg struct {
	target   *Compose
	action   mediaAction
	media    north.Media
	id       string
	warning  *north.MediaWarning
	alt      string
	response *north.Response
	err      error
}

func (m mediaUpdateMsg) Response() *north.Response { return m.response }

func (d *Compose) updateMedia(ctx *reactea.Ctx, msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case modal.Result[TextResult]:
		if d.mediaPrompt == mediaPromptNone {
			return nil, false
		}
		prompt := d.mediaPrompt
		d.mediaPrompt = mediaPromptNone
		if !msg.Ok() || msg.Value.Canceled {
			return nil, true
		}
		if prompt == mediaPromptAlt {
			return d.setLastAltText(ctx.Context(), msg.Value.Text), true
		}

		return d.uploadMedia(ctx.Context(), msg.Value.Text), true
	case mediaUpdateMsg:
		if msg.target != d {
			return nil, false
		}
		d.applyMediaUpdate(msg)

		return nil, true
	}

	return nil, false
}

func (d *Compose) promptMedia(ctx *reactea.Ctx, prompt mediaPromptKind) tea.Cmd {
	if d.mediaAPI == nil || d.mediaBusy {
		return nil
	}
	if prompt == mediaPromptPath {
		if len(d.mediaIDs) >= 4 {
			d.mediaNotice = "A post can have at most four attachments"

			return nil
		}
		d.mediaPrompt = prompt

		return modal.PushAt(ctx, NewTextPrompt(d.theme, "Attach media", "Path to an image, GIF, or video", "", "Upload"), Placement(ctx, 62, 8))
	}
	if len(d.mediaIDs) == 0 {
		d.mediaNotice = "Attach media first"

		return nil
	}
	d.mediaPrompt = prompt
	initial := ""
	if media := d.lastMedia(); media != nil && media.AltText != nil {
		initial = *media.AltText
	}

	return modal.PushAt(ctx, NewTextPrompt(d.theme, "Alternative text", "Describe the attachment", initial, "Save"), Placement(ctx, 62, 8))
}

func (d *Compose) uploadMedia(ctx context.Context, path string) tea.Cmd {
	if d.mediaAPI == nil || d.mediaBusy || len(d.mediaIDs) >= 4 {
		return nil
	}
	path = expandHome(strings.TrimSpace(path))
	if path == "" {
		return nil
	}
	d.mediaBusy = true
	d.mediaNotice = "Uploading media…"

	return func() tea.Msg {
		media, response, err := d.mediaAPI.UploadMediaFile(ctx, path, north.WithMediaPurpose(north.MediaForPosts))

		return mediaUpdateMsg{target: d, action: mediaUploaded, media: media, response: response, err: err}
	}
}

func (d *Compose) removeLastMedia(ctx context.Context) tea.Cmd {
	if d.mediaAPI == nil || d.mediaBusy || len(d.mediaIDs) == 0 {
		return nil
	}
	id := d.mediaIDs[len(d.mediaIDs)-1]
	if !d.uploaded[id] {
		d.removeMediaID(id)
		d.mediaNotice = "Attachment removed"

		return nil
	}
	d.mediaBusy = true
	d.mediaNotice = "Removing media…"

	return func() tea.Msg {
		_, response, err := d.mediaAPI.DeleteMedia(ctx, id)

		return mediaUpdateMsg{target: d, action: mediaDeleted, id: id, response: response, err: err}
	}
}

func (d *Compose) setLastAltText(ctx context.Context, text string) tea.Cmd {
	if d.mediaAPI == nil || d.mediaBusy || len(d.mediaIDs) == 0 {
		return nil
	}
	id := d.mediaIDs[len(d.mediaIDs)-1]
	d.mediaBusy = true
	d.mediaNotice = "Saving alternative text…"

	return func() tea.Msg {
		_, response, err := d.mediaAPI.SetMediaAltText(ctx, id, text)

		return mediaUpdateMsg{target: d, action: mediaAltUpdated, id: id, alt: text, response: response, err: err}
	}
}

func (d *Compose) cycleMediaWarning(ctx context.Context) tea.Cmd {
	if d.mediaAPI == nil || d.mediaBusy || len(d.mediaIDs) == 0 {
		return nil
	}
	media := d.lastMedia()
	next := north.WarningSensitive
	var warning *north.MediaWarning = &next
	if media != nil {
		if media.Warning != nil {
			switch *media.Warning {
			case north.WarningSensitive:
				next = north.WarningViolence
			case north.WarningViolence:
				next = north.WarningNudity
			case north.WarningNudity:
				warning = nil
			}
		}
		if media.Warning == nil || *media.Warning != north.WarningNudity {
			warning = &next
		}
	}
	id := d.mediaIDs[len(d.mediaIDs)-1]
	d.mediaBusy = true
	d.mediaNotice = "Updating content warning…"

	return func() tea.Msg {
		updated, response, err := d.mediaAPI.SetMediaWarning(ctx, id, warning)

		return mediaUpdateMsg{target: d, action: mediaWarningUpdated, id: id, warning: updated, response: response, err: err}
	}
}

func (d *Compose) applyMediaUpdate(msg mediaUpdateMsg) {
	d.mediaBusy = false
	if msg.err != nil {
		d.mediaNotice = ui.FriendlyError(msg.err)

		return
	}
	switch msg.action {
	case mediaUploaded:
		if msg.media.ID == "" {
			d.mediaNotice = "Upload returned no media ID"

			return
		}
		d.media = append(d.media, msg.media)
		d.mediaIDs = append(d.mediaIDs, msg.media.ID)
		d.uploaded[msg.media.ID] = true
		d.mediaNotice = "Media attached"
	case mediaDeleted:
		d.removeMediaID(msg.id)
		d.mediaNotice = "Attachment removed"
	case mediaAltUpdated:
		if media := d.mediaWithID(msg.id); media != nil {
			alt := msg.alt
			media.AltText = &alt
		}
		d.mediaNotice = "Alternative text saved"
	case mediaWarningUpdated:
		if media := d.mediaWithID(msg.id); media != nil {
			media.Warning = msg.warning
		}
		d.mediaNotice = "Content warning updated"
	}
}

func (d *Compose) renderMedia(width int) []string {
	if len(d.mediaIDs) == 0 && d.mediaNotice == "" {
		return nil
	}
	lines := make([]string, 0, len(d.mediaIDs)+1)
	for index, id := range d.mediaIDs {
		label := fmt.Sprintf("[%d] Attachment", index+1)
		if media := d.mediaWithID(id); media != nil {
			label = fmt.Sprintf("[%d] %s", index+1, media.Kind)
			if media.AltText != nil && *media.AltText != "" {
				label += " · alt text"
			}
			if media.Warning != nil {
				label += " · " + strings.ToLower(string(*media.Warning))
			}
		}
		lines = append(lines, d.theme.Dim.Render(ui.Clip(label, width)))
	}
	if d.mediaNotice != "" {
		lines = append(lines, d.theme.Dim.Render(ui.Clip(d.mediaNotice+" · Alt+T alt · Alt+W warning · Alt+X remove", width)))
	}

	return lines
}

func (d *Compose) lastMedia() *north.Media {
	if len(d.mediaIDs) == 0 {
		return nil
	}

	return d.mediaWithID(d.mediaIDs[len(d.mediaIDs)-1])
}

func (d *Compose) mediaWithID(id string) *north.Media {
	for index := range d.media {
		if d.media[index].ID == id {
			return &d.media[index]
		}
	}

	return nil
}

func (d *Compose) removeMediaID(id string) {
	for index := range d.mediaIDs {
		if d.mediaIDs[index] == id {
			d.mediaIDs = append(d.mediaIDs[:index], d.mediaIDs[index+1:]...)
			break
		}
	}
	for index := range d.media {
		if d.media[index].ID == id {
			d.media = append(d.media[:index], d.media[index+1:]...)
			break
		}
	}
	delete(d.uploaded, id)
}

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return filepath.Clean(path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(path)
	}
	if path == "~" {
		return home
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}
