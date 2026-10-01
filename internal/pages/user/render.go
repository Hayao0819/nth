package user

import (
	"fmt"
	"strings"

	"github.com/Hayao0819/nth/internal/components/pageheader"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (d *Screen) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	innerWidth, room := d.layout(width, height)
	content := d.content(innerWidth)
	d.clampOffset(innerWidth, room)
	end := min(len(content.lines), d.offset+room)

	right := ""
	if d.profileLoading || d.postsLoading {
		right = d.theme.Dim.Render("Loading…")
	}
	header := pageheader.Render(d.theme, "Profile", right, innerWidth)
	body := ""
	if d.offset < end {
		body = strings.Join(content.lines[d.offset:end], "\n")
	}

	footerLeft := ""
	if len(d.posts) > 0 {
		footerLeft = "j/k posts · enter open"
	} else if len(content.lines) > room {
		footerLeft = "j/k · ↑/↓ scroll"
	}
	footerRight := "Esc/← back"
	if d.canRetry() {
		footerRight = ". retry · " + footerRight
	} else if len(d.posts) > 0 {
		progress := fmt.Sprintf("%d of %d", d.selected+1, len(d.posts))
		if d.loadingMore {
			progress += " · loading more…"
		} else if d.nextCursor != nil {
			progress += " · more below"
		}
		footerRight = progress + " · " + footerRight
	}
	footer := ui.Sides(footerLeft, footerRight, innerWidth)

	return ui.Fit(header+"\n"+ui.Fit(body, innerWidth, room)+"\n"+d.theme.Dim.Render(footer), width, height)
}

func (d *Screen) layout(width, height int) (int, int) {
	return max(0, width), max(0, height-pageheader.Height-1)
}
