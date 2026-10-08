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
	} else if d.profileNotice != "" {
		if d.profileBusy {
			right = d.theme.Dim.Render(ui.Clip(d.profileNotice, max(1, innerWidth-12)))
		} else if d.profileUpdateErr != nil {
			right = d.theme.Warn.Render(ui.Clip(d.profileNotice, max(1, innerWidth-12)))
		} else {
			right = d.theme.Active.Render(ui.Clip(d.profileNotice, max(1, innerWidth-12)))
		}
	}
	header := pageheader.RenderWithAction(d.theme, "Profile", right, d.browserAction(), innerWidth)
	body := ""
	if d.offset < end {
		body = strings.Join(content.lines[d.offset:end], "\n")
	}

	footerLeft := ""
	if d.browserURL() != "" {
		footerLeft = "w browser"
	}
	if d.activity != nil {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "Tab posts/likes"
	}
	if len(d.relationshipKinds()) > 0 {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "F/B/M/N account"
	}
	if d.profile != nil && d.isOwnProfile() {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "E edit profile"
	}
	if d.accountSafety && d.isOwnProfile() {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "S privacy"
	}
	if d.connections != nil {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "o/O following/followers"
	}
	if d.listMembers != nil && !d.isOwnProfile() {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "A lists"
	}
	if len(d.posts) > 0 {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "j/k posts · enter open"
	} else if len(content.lines) > room {
		if footerLeft != "" {
			footerLeft += " · "
		}
		footerLeft += "j/k · ↑/↓ scroll"
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
