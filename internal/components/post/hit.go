package post

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
)

func CardCursorRow(post north.Post) int {
	if post.RepostOf != nil {
		return 1
	}

	return 0
}

func CardUserAt(post north.Post, x, row, width int) (north.User, bool) {
	return CardUserAtWithImages(post, x, row, width, nil)
}

func CardUserAtWithImages(post north.Post, x, row, width int, images *termimage.Renderer) (north.User, bool) {
	if width <= 0 || x < 0 || x >= width || row < 0 {
		return north.User{}, false
	}
	target := post.DisplayPost()
	if target == nil {
		return north.User{}, false
	}

	currentRow := 0
	if post.RepostOf != nil {
		left := 6
		if row == currentRow {
			if user, ok := cardUserHit(post.Author, x, left, "@"+ui.SafeInline(post.Author.Handle), width); ok {
				return user, true
			}
		}
		currentRow++
	}

	badges := ""
	if target.Author.Verified {
		badges += " ✓"
	}
	if target.Author.Protected {
		badges += " 🔒"
	}
	label := ui.SafeInline(target.Author.Name) + badges + " @" + ui.SafeInline(target.Author.Handle)
	if row == currentRow {
		if user, ok := cardUserHit(target.Author, x, 3+termimage.AvatarColumns, label, width); ok {
			return user, true
		}
	}
	currentRow++

	if target.InReplyToHandle != nil && *target.InReplyToHandle != "" {
		left := 4 + lipgloss.Width("Replying to ")
		if row == currentRow {
			user := north.User{Handle: *target.InReplyToHandle}
			if user, ok := cardUserHit(user, x, left, "@"+ui.SafeInline(*target.InReplyToHandle), width); ok {
				return user, true
			}
		}
		currentRow++
	}
	concealed := Concealed(target)
	if target.Deleted || target.HiddenReason != "" || target.Unavailable {
		currentRow++
	}
	if !concealed {
		body := ui.WrappedLines(target.Text, max(1, width-4))
		currentRow += len(body)
		if len(body) == 0 && len(target.Media) == 0 {
			currentRow++
		}
		for _, media := range target.Media {
			bodyRows := len(cardMediaBody(media, max(1, width-2), images))
			currentRow += bodyRows + 2
		}
	}

	if !concealed && target.Quoted != nil {
		quoted := target.Quoted.DisplayPost()
		label := ui.SafeInline(quoted.Author.Name) + " @" + ui.SafeInline(quoted.Author.Handle)
		if row == currentRow {
			if user, ok := cardUserHit(quoted.Author, x, 7, label, width); ok {
				return user, true
			}
		}
	}

	return north.User{}, false
}

// CardQuotedPostAt reports whether a position belongs to the quoted post.
// Check CardUserAt first so the quoted author's label remains a user link.
func CardQuotedPostAt(post north.Post, x, row, width int) (north.Post, bool) {
	return CardQuotedPostAtWithImages(post, x, row, width, nil)
}

func CardQuotedPostAtWithImages(post north.Post, x, row, width int, images *termimage.Renderer) (north.Post, bool) {
	quoted, first, last, ok := cardQuotedPostRows(post, width, images)
	if !ok || x < 4 || x >= width || row < first || row > last {
		return north.Post{}, false
	}

	return quoted, true
}

func cardQuotedPostRows(post north.Post, width int, images *termimage.Renderer) (north.Post, int, int, bool) {
	if width <= 0 {
		return north.Post{}, 0, 0, false
	}
	target := post.DisplayPost()
	if target == nil || Concealed(target) || target.Quoted == nil {
		return north.Post{}, 0, 0, false
	}
	quoted := target.Quoted.DisplayPost()
	if Concealed(quoted) {
		return north.Post{}, 0, 0, false
	}

	currentRow := 0
	if post.RepostOf != nil {
		currentRow++
	}
	currentRow++
	if target.InReplyToHandle != nil && *target.InReplyToHandle != "" {
		currentRow++
	}
	body := ui.WrappedLines(target.Text, max(1, width-4))
	currentRow += len(body)
	if len(body) == 0 && len(target.Media) == 0 {
		currentRow++
	}
	for _, media := range target.Media {
		bodyRows := len(cardMediaBody(media, max(1, width-2), images))
		currentRow += bodyRows + 2
	}

	quoteWidth := max(4, width-4)
	quoteBody := ui.WrappedLines(quoted.Text, max(1, quoteWidth-4))
	quoteHeight := len(quoteBody) + 2

	return *quoted, currentRow, currentRow + quoteHeight - 1, true
}

func cardUserHit(user north.User, x, left int, label string, limit int) (north.User, bool) {
	user.Handle = strings.TrimPrefix(strings.TrimSpace(user.Handle), "@")
	if user.Handle == "" || !hitText(x, left, label, limit) {
		return north.User{}, false
	}

	return user, true
}

func hitText(x, left int, label string, limit int) bool {
	return x >= left && x < min(limit, left+lipgloss.Width(label))
}
