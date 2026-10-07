package post

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
)

type userHit struct {
	user        north.User
	row         int
	left, right int
}

type postHit struct {
	post        north.Post
	top, bottom int
	left, right int
	card        bool
	index       int
}

func makeUserHit(user north.User, row, left int, label string, limit int) (userHit, bool) {
	user.Handle = strings.TrimPrefix(strings.TrimSpace(user.Handle), "@")
	right := min(limit, left+lipgloss.Width(label))
	if user.Handle == "" || left >= right {
		return userHit{}, false
	}

	return userHit{user: user, row: row, left: left, right: right}, true
}

func userAt(hits []userHit, x, row int) (north.User, bool) {
	for _, hit := range hits {
		if row == hit.row && x >= hit.left && x < hit.right {
			return hit.user, true
		}
	}

	return north.User{}, false
}

func postAt(hits []postHit, x, row int) (north.Post, bool) {
	for _, hit := range hits {
		if row >= hit.top && row <= hit.bottom && x >= hit.left && x < hit.right {
			return hit.post, true
		}
	}

	return north.Post{}, false
}

func (d *Screen) userAtPosition(x, y, width, room int) (north.User, bool) {
	if room <= 0 {
		return north.User{}, false
	}
	visible := y - pageheader.Height
	if visible < 0 || visible >= room {
		return north.User{}, false
	}
	lines, hits, posts := d.content(width)
	contentRow := clampOffset(d.offset, len(lines), room) + visible
	for _, hit := range posts {
		if !hit.card || contentRow < hit.top || contentRow > hit.bottom || x < hit.left || x >= hit.right {
			continue
		}
		if user, ok := postcomponent.CardUserAtWithImages(
			hit.post,
			x-hit.left,
			contentRow-hit.top,
			hit.right-hit.left,
			d.images,
		); ok {
			d.replyFocused = true
			d.replyIndex = hit.index

			return user, true
		}
	}

	return userAt(hits, x, contentRow)
}

func (d *Screen) postAtPosition(x, y, width, room int) (north.Post, bool) {
	if room <= 0 {
		return north.Post{}, false
	}
	visible := y - pageheader.Height
	if visible < 0 || visible >= room {
		return north.Post{}, false
	}
	lines, _, hits := d.content(width)
	contentRow := clampOffset(d.offset, len(lines), room) + visible
	for _, hit := range hits {
		if contentRow < hit.top || contentRow > hit.bottom || x < hit.left || x >= hit.right {
			continue
		}
		if !hit.card {
			return hit.post, true
		}
		d.replyFocused = true
		d.replyIndex = hit.index
		if quoted, ok := postcomponent.CardQuotedPostAtWithImages(
			hit.post,
			x-hit.left,
			contentRow-hit.top,
			hit.right-hit.left,
			d.images,
		); ok {
			return quoted, true
		}

		return hit.post, true
	}

	return north.Post{}, false
}

func (d *Screen) replyActionAtPosition(x, y, width, room int) (postcomponent.Action, north.Post, bool) {
	if room <= 0 {
		return 0, north.Post{}, false
	}
	visible := y - pageheader.Height
	if visible < 0 || visible >= room {
		return 0, north.Post{}, false
	}
	lines, _, hits := d.content(width)
	contentRow := clampOffset(d.offset, len(lines), room) + visible
	for _, hit := range hits {
		if !hit.card || contentRow != hit.bottom-1 || x < hit.left || x >= hit.right {
			continue
		}
		if !postcomponent.CanInteract(&hit.post) {
			return 0, north.Post{}, false
		}
		d.replyFocused = true
		d.replyIndex = hit.index

		return postcomponent.CardActionAt(x-hit.left, hit.right-hit.left), hit.post, true
	}

	return 0, north.Post{}, false
}
