package user

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
)

type postPosition struct {
	index  int
	top    int
	height int
	left   int
	width  int
}

type pageContent struct {
	lines []string
	posts []postPosition
}

func (c pageContent) postAt(row int) (postPosition, bool) {
	for _, position := range c.posts {
		if row >= position.top && row < position.top+position.height {
			return position, true
		}
	}

	return postPosition{}, false
}

func (d *Screen) content(width int) pageContent {
	if width <= 0 {
		return pageContent{}
	}
	inset := 0
	if width >= 12 {
		inset = 2
	}
	contentWidth := max(1, width-inset*2)
	content := pageContent{}
	appendLines := func(lines ...string) {
		padding := strings.Repeat(" ", inset)
		for _, line := range lines {
			content.lines = append(content.lines, padding+ui.Left(line, contentWidth))
		}
	}

	if d.profileErr != nil {
		message := "Profile details unavailable: " + ui.FriendlyError(d.profileErr)
		for index, line := range ui.WrappedLines(message, max(1, contentWidth-2)) {
			prefix := "  "
			if index == 0 {
				prefix = "! "
			}
			appendLines(d.theme.Warn.Render(prefix + line))
		}
		appendLines("")
	}
	appendLines(profileLinesWithImages(d.theme, d.user, contentWidth, d.profileLoaded, d.images)...)
	appendLines("", d.theme.Heading.Render("Posts"), "")

	switch {
	case d.postsLoading && len(d.posts) == 0:
		appendLines(d.theme.Dim.Render("Loading posts…"))
	case d.postsErr != nil && len(d.posts) == 0:
		message := "Posts unavailable: " + ui.FriendlyError(d.postsErr)
		for index, line := range ui.WrappedLines(message, max(1, contentWidth-2)) {
			prefix := "  "
			if index == 0 {
				prefix = "! "
			}
			appendLines(d.theme.Warn.Render(prefix + line))
		}
	case len(d.posts) == 0:
		appendLines(d.theme.Dim.Render("No posts yet"))
	}

	for index, post := range d.posts {
		card := postcomponent.RenderCardWithImages(post, contentWidth, index == d.selected, d.theme, time.Now(), d.images)
		cardLines := strings.Split(card, "\n")
		position := postPosition{
			index:  index,
			top:    len(content.lines),
			height: len(cardLines),
			left:   inset,
			width:  contentWidth,
		}
		appendLines(cardLines...)
		content.posts = append(content.posts, position)
	}
	if d.postsErr != nil && len(d.posts) > 0 {
		appendLines(d.theme.Warn.Render("! Could not load older posts: " + ui.FriendlyError(d.postsErr)))
	} else if d.loadingMore {
		appendLines(d.theme.Dim.Render("  Loading older posts…"))
	}

	return content
}

func (d *Screen) selectedPost() *north.Post {
	if d.selected < 0 || d.selected >= len(d.posts) {
		return nil
	}

	return &d.posts[d.selected]
}

func (d *Screen) requestSelected(action postcomponent.Action) tea.Cmd {
	post := d.selectedPost()
	if post == nil {
		return nil
	}
	return postcomponent.Request(action, *post)
}

func (d *Screen) moveSelection(delta, width, room int) {
	if len(d.posts) == 0 {
		return
	}
	d.selected = min(max(0, d.selected+delta), len(d.posts)-1)
	d.ensureSelectionVisible(width, room)
}

func (d *Screen) ensureSelectionVisible(width, room int) {
	if room <= 0 || len(d.posts) == 0 {
		return
	}
	d.clampSelection()
	content := d.content(width)
	for _, position := range content.posts {
		if position.index != d.selected {
			continue
		}
		cursor := position.top + postcomponent.CardCursorRow(d.posts[d.selected])
		if cursor < d.offset {
			d.offset = cursor
		} else if cursor-d.offset > room/2 {
			d.offset = cursor - room/2
		}
		break
	}
	d.clampOffset(width, room)
}

func (d *Screen) clampSelection() {
	if len(d.posts) == 0 {
		d.selected = 0
		return
	}
	d.selected = min(max(0, d.selected), len(d.posts)-1)
}

func (d *Screen) clampOffset(width, room int) {
	d.offset = min(max(0, d.offset), max(0, len(d.content(width).lines)-room))
}

func (d *Screen) applyReaction(update postcomponent.ReactionUpdate) {
	for index := range d.posts {
		target := d.posts[index].DisplayPost()
		if target == nil || target.ID != update.PostID || update.Err != nil {
			continue
		}
		if update.Like != nil {
			target.Liked = update.Like.Liked
			target.LikeCount = update.Like.LikeCount
		}
		if update.Repost != nil {
			target.Reposted = update.Repost.Reposted
			target.RepostCount = update.Repost.RepostCount
		}
	}
}
