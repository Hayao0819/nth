package post

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
)

const maxAncestorPosts = 32

type PostAPI interface {
	Post(context.Context, string) (north.Post, *north.Response, error)
}

type ancestorsLoadedMsg struct {
	target    *Screen
	posts     []north.Post
	resp      *north.Response
	err       error
	truncated bool
}

func (m ancestorsLoadedMsg) Response() *north.Response { return m.resp }

func (d *Screen) loadAncestors(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		id := replyParentID(d.post)
		seen := map[string]bool{d.post.ID: true}
		if target := d.post.DisplayPost(); target != nil {
			seen[target.ID] = true
		}
		posts := make([]north.Post, 0, 4)
		var response *north.Response
		for id != "" && len(posts) < maxAncestorPosts {
			if seen[id] {
				return ancestorsLoadedMsg{target: d, posts: reversePosts(posts), resp: response, truncated: true}
			}
			seen[id] = true
			post, currentResponse, err := d.api.Post(ctx, id)
			if currentResponse != nil {
				response = currentResponse
			}
			if err != nil {
				return ancestorsLoadedMsg{target: d, posts: reversePosts(posts), resp: response, err: err}
			}
			posts = append(posts, post)
			id = replyParentID(post)
		}

		return ancestorsLoadedMsg{
			target:    d,
			posts:     reversePosts(posts),
			resp:      response,
			truncated: id != "",
		}
	}
}

func (d *Screen) threadLines(width int) []string {
	lines, _, _ := d.threadContent(width)

	return lines
}

func (d *Screen) threadContent(width int) ([]string, []userHit, []postHit) {
	if width <= 0 || len(d.ancestors) == 0 && !d.loadingAncestors && d.ancestorErr == nil && !d.ancestorsTruncated {
		return nil, nil, nil
	}

	lines := make([]string, 0, len(d.ancestors)*4+2)
	hits := make([]userHit, 0, len(d.ancestors))
	postHits := make([]postHit, 0, len(d.ancestors))
	for _, post := range d.ancestors {
		first := len(lines)
		ancestorLines, hit, ok := d.ancestorContent(post, width)
		if ok {
			hit.row += len(lines)
			hits = append(hits, hit)
		}
		lines = append(lines, ancestorLines...)
		if target := post.DisplayPost(); target != nil && strings.TrimSpace(target.ID) != "" && len(ancestorLines) > 0 {
			postHits = append(postHits, postHit{post: post, top: first, bottom: len(lines) - 1, right: width})
		}
	}
	guide := d.theme.Dim.Render("│ ")
	switch {
	case d.loadingAncestors:
		lines = append(lines, guide+d.theme.Dim.Render("Loading earlier replies…"), guide)
	case d.ancestorErr != nil:
		message := "Earlier replies unavailable: " + ui.FriendlyError(d.ancestorErr)
		for _, line := range ui.WrappedLines(message, max(1, width-2)) {
			lines = append(lines, guide+d.theme.Warn.Render(line))
		}
		lines = append(lines, guide)
	case d.ancestorsTruncated:
		lines = append(lines, guide+d.theme.Dim.Render("Earlier replies omitted"), guide)
	}

	return lines, hits, postHits
}

func (d *Screen) ancestorContent(post north.Post, width int) ([]string, userHit, bool) {
	target := post.DisplayPost()
	if target == nil {
		return []string{d.theme.Dim.Render("│ ") + d.theme.Dim.Render("Post unavailable"), d.theme.Dim.Render("│")}, userHit{}, false
	}

	contentWidth := max(1, width-2)
	guide := d.theme.Dim.Render("│ ")
	label := ui.SafeInline(target.Author.Name) + " @" + ui.SafeInline(target.Author.Handle)
	identity := d.theme.Name.Render(ui.SafeInline(target.Author.Name)) + " " + d.theme.Handle.Render("@"+ui.SafeInline(target.Author.Handle))
	if !target.CreatedAt.IsZero() {
		identity += d.theme.Dim.Render(" · " + postcomponent.RelativeTime(target.CreatedAt, time.Now()))
	}
	lines := []string{guide + ui.Clip(identity, contentWidth)}
	hit, clickable := makeUserHit(target.Author, 0, 2, label, width)

	text := target.Text
	if target.Deleted {
		text = "Post deleted"
	} else if target.HiddenReason != "" {
		text = "Hidden: " + string(target.HiddenReason)
	} else if target.Unavailable {
		text = "Post unavailable"
	}
	for _, line := range ui.WrappedLines(text, contentWidth) {
		lines = append(lines, guide+line)
	}
	if !postcomponent.Concealed(target) {
		for _, media := range target.Media {
			label := "[" + postcomponent.MediaLabel(media) + "]"
			lines = append(lines, guide+d.theme.Dim.Render(ui.Clip(label, contentWidth)))
		}
	}
	lines = append(lines, d.theme.Dim.Render("│"))

	return lines, hit, clickable
}

func replyParentID(post north.Post) string {
	target := post.DisplayPost()
	if target == nil || target.InReplyToID == nil {
		return ""
	}

	return strings.TrimSpace(*target.InReplyToID)
}

func reversePosts(posts []north.Post) []north.Post {
	for left, right := 0, len(posts)-1; left < right; left, right = left+1, right-1 {
		posts[left], posts[right] = posts[right], posts[left]
	}

	return posts
}
