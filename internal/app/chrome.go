package app

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/feed"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type timelineHeader struct {
	reactea.BasicComponent
	root *root
}

type compactHeaderRenderKey struct {
	mode   feed.Mode
	query  string
	name   string
	handle string
}

func (h *timelineHeader) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return nil
	}
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	if y < 0 || y > 2 || h.root.feed.CurrentMode() != feed.Home && h.root.feed.CurrentMode() != feed.Ranked {
		return nil
	}
	if x < ctx.Width()/2 {
		return h.root.feed.SetMode(ctx, feed.Ranked, "")
	}

	return h.root.feed.SetMode(ctx, feed.Home, "")
}

func (h *timelineHeader) Render(ctx *reactea.Ctx) string {
	return h.root.renderTimelineHeader(ctx)
}

func (r *root) compactHeaderKey() compactHeaderRenderKey {
	key := compactHeaderRenderKey{mode: r.feed.CurrentMode(), query: r.feed.SearchQuery()}
	if r.me != nil {
		key.name = r.me.Name
		key.handle = r.me.Handle
	}

	return key
}

func (r *root) renderHeader(ctx *reactea.Ctx) string {
	width := ctx.Width()
	mode := r.feed.CurrentMode().Label(r.feed.SearchQuery())
	title := r.theme.Brand.Render("north") + r.theme.Dim.Render("  /  ") + r.theme.Title.Render(mode)

	account := ""
	if r.me != nil {
		account = "@" + ui.SafeInline(r.me.Handle)
		if width >= 52 {
			account = ui.SafeInline(r.me.Name) + " " + account
		}
	}
	first := ui.Sides(title, r.theme.Handle.Render(account), width)

	return first + "\n" + ui.Left(r.renderNavigation(width), width)
}

func (r *root) renderNavigation(width int) string {
	if r.feed.CurrentMode() == feed.Search {
		return r.theme.Active.Render("● Search") + r.theme.Dim.Render("  esc back  ·  tab timelines")
	}

	tabs := []struct {
		mode  feed.Mode
		label string
	}{
		{feed.Ranked, "For you"},
		{feed.Home, "Following"},
	}
	if width < 64 {
		return ui.Columns([]string{
			r.theme.Dim.Render("‹ Previous"),
			r.theme.Dim.Render("Next ›"),
		}, width)
	}

	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := tab.label
		if r.feed.CurrentMode() == tab.mode {
			label = r.theme.Active.Underline(true).Render(label)
		} else {
			label = r.theme.Dim.Render(label)
		}
		parts = append(parts, label)
	}
	return ui.Columns(parts, width)
}

func (r *root) renderTimelineHeader(ctx *reactea.Ctx) string {
	width := ctx.Width()
	mode := r.feed.CurrentMode()
	if mode == feed.Home || mode == feed.Ranked {
		leftWidth := width / 2
		forYou := timelineTab("For you", leftWidth, mode == feed.Ranked, r.theme)
		following := timelineTab("Following", width-leftWidth, mode == feed.Home, r.theme)

		return "\n" + forYou + following + "\n"
	}

	title := mode.Label(r.feed.SearchQuery())
	description := "Latest posts"
	switch mode {
	case feed.Search:
		title = "Search"
		description = "Latest results for “" + ui.SafeInline(r.feed.SearchQuery()) + "”"
	}

	return "\n" + ui.Sides("  "+r.theme.Title.Render(title), r.theme.Dim.Render(description+"  "), width) + "\n"
}

func timelineTab(label string, width int, active bool, theme ui.Theme) string {
	style := theme.Dim
	if active {
		style = theme.Active.Underline(true)
	}

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(style.Render(label))
}

func (r *root) renderStatusBar(ctx *reactea.Ctx) string {
	width := ctx.Width()
	progress := r.feed.Progress()
	if progress == "" {
		progress = r.feed.CurrentMode().Label(r.feed.SearchQuery())
	}
	right := r.statusForWidth(width)
	if width >= 90 {
		right += r.theme.Dim.Render("  ·  ? help  q quit")
	}

	return ui.Sides(r.theme.Dim.Render("  "+progress), right, width)
}

func (r *root) renderFooter(ctx *reactea.Ctx) string {
	width := ctx.Width()
	progress := r.feed.Progress()
	if progress == "" {
		progress = r.feed.CurrentMode().Label(r.feed.SearchQuery())
	}
	first := ui.Sides(r.theme.Dim.Render(progress), r.statusForWidth(width), width)

	var hints string
	switch {
	case width >= 96:
		hints = "j/k move · enter details · u/U profiles · l like · t repost · r reply · Q quote · n post · / search · ? help"
		if r.notifications != nil {
			hints = "j/k move · enter details · u/U profiles · l like · t repost · r reply · Q quote · 2 notices · n post · / search · ? help"
		}
	case width >= 72:
		hints = "j/k move · enter open · u profile · l like · t repost · r reply · n post · / search · ? help"
	case width >= 52:
		hints = "j/k move · enter open · u profile · l like · n post · / search · ? help"
	default:
		hints = "j/k move · enter open · ? help"
	}

	return first + "\n" + r.theme.Dim.Render(ui.Left(hints, width))
}

func (r *root) styledStatus() string {
	status := r.status()
	if r.problem != nil || r.notice == "" && !r.posting && !r.editing && (r.feed.Failed() || r.meErr != nil) {
		return r.theme.Bad.Render(status)
	}
	if r.posting || r.editing || strings.HasSuffix(status, "…") {
		return r.theme.Warn.Render(status)
	}

	return r.theme.Dim.Render(status)
}

func (r *root) statusForWidth(width int) string {
	status := r.styledStatus()
	if width < 72 && status != "" {
		return status
	}

	return statusWithRate(status, r.latestResponse())
}

func (r *root) status() string {
	switch {
	case r.problem != nil:
		message := ui.FriendlyError(r.problem)
		if r.postFailed {
			job := r.failureJob
			if job == "" {
				job = "Post"
			}

			return job + " failed · draft kept — " + message
		}

		return message
	case r.posting:
		return "Posting…"
	case r.editing:
		return "Updating post…"
	case r.notice != "":
		return r.notice
	case r.feed.Failed():
		return r.feed.Status()
	case r.meErr != nil:
		return ui.FriendlyError(r.meErr)
	case r.feed.Status() != "":
		return r.feed.Status()
	default:
		return ""
	}
}

func (r *root) latestResponse() *north.Response {
	if resp := r.feed.LastResponse(); resp != nil && r.feed.LastResponseAt().After(r.respAt) {
		return resp
	}

	return r.resp
}

func (r *root) setResponse(response *north.Response) {
	if response == nil {
		return
	}
	r.resp = response
	r.respAt = time.Now()
}

func statusWithRate(status string, resp *north.Response) string {
	if resp == nil || !resp.RateLimit.Present {
		return status
	}

	rate := fmt.Sprintf("last request: %d/%d left", resp.RateLimit.Remaining, resp.RateLimit.Limit)
	if status == "" {
		return rate
	}

	return status + " · " + rate
}
