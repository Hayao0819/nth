package user

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
)

const profileHeaderRows = 4

func profileLines(theme ui.Theme, user north.User, width int) []string {
	return profileLinesWithDetails(theme, user, width, true)
}

func profileLinesWithDetails(theme ui.Theme, user north.User, width int, detailsKnown bool) []string {
	return profileLinesWithImages(theme, user, width, detailsKnown, nil)
}

func profileLinesWithImages(theme ui.Theme, user north.User, width int, detailsKnown bool, images *termimage.Renderer) []string {
	lines, _ := profileLinesAndStats(theme, user, width, detailsKnown, images)

	return lines
}

func profileLinesAndStats(theme ui.Theme, user north.User, width int, detailsKnown bool, images *termimage.Renderer) ([]string, int) {
	if width <= 0 {
		return nil, -1
	}

	var lines []string
	if user.HeaderURL != nil {
		if rendered := images.View(*user.HeaderURL, profileHeaderColumns(width), profileHeaderRows); rendered != "" {
			lines = append(lines, strings.Split(rendered, "\n")...)
			lines = append(lines, "")
		}
	}
	if user.AvatarURL != nil {
		if rendered := images.View(*user.AvatarURL, 6, 3); rendered != "" {
			lines = append(lines, strings.Split(rendered, "\n")...)
			lines = append(lines, "")
		}
	}
	lines = append(lines, identityLines(theme, user, width)...)
	if !detailsKnown {
		return lines, -1
	}
	lines = append(lines, "")
	if user.Bio != nil && strings.TrimSpace(*user.Bio) != "" {
		lines = append(lines, ui.WrappedLines(*user.Bio, width)...)
	} else {
		lines = append(lines, theme.Dim.Render("No bio yet"))
	}

	lines = append(lines, "")
	statsRow := len(lines)
	lines = append(lines,
		ui.Columns([]string{
			theme.Heading.Render(strconv.Itoa(user.PostCount)),
			theme.Heading.Render(strconv.Itoa(user.FollowingCount)),
			theme.Heading.Render(strconv.Itoa(user.FollowerCount)),
		}, width),
		ui.Columns([]string{
			theme.Dim.Render("Posts"),
			theme.Dim.Render("Following"),
			theme.Dim.Render("Followers"),
		}, width),
	)

	details := detailItems(theme, user)
	if len(details) > 0 {
		lines = append(lines, "")
		lines = append(lines, packDetails(details, width)...)
	}

	return lines, statsRow
}

func profileHeaderColumns(width int) int {
	return max(1, width)
}

func identityLines(theme ui.Theme, user north.User, width int) []string {
	badges := ""
	if user.Verified {
		badges += " ✓"
	}
	if user.Protected {
		badges += " 🔒"
	}
	name := ui.SafeInline(user.Name)
	if name == "" {
		name = "@" + strings.TrimPrefix(ui.SafeInline(user.Handle), "@")
	}
	identity := theme.Accent.Render("●") + "  " + theme.Name.Render(name+badges)
	relation := relationshipLabel(theme, user)

	lines := make([]string, 0, 3)
	if relation != "" && lipgloss.Width(identity)+lipgloss.Width(relation)+2 <= width {
		lines = append(lines, ui.Sides(identity, relation, width))
	} else {
		lines = append(lines, ui.Clip(identity, width))
	}
	indent := strings.Repeat(" ", min(3, max(0, width-1)))
	handle := theme.Handle.Render("@" + strings.TrimPrefix(ui.SafeInline(user.Handle), "@"))
	lines = append(lines, indent+ui.Clip(handle, max(1, width-lipgloss.Width(indent))))
	if relation != "" && lipgloss.Width(identity)+lipgloss.Width(relation)+2 > width {
		lines = append(lines, indent+ui.Clip(relation, max(1, width-lipgloss.Width(indent))))
	}

	return lines
}

func relationshipLabel(theme ui.Theme, user north.User) string {
	parts := make([]string, 0, 4)
	if user.Blocking {
		parts = append(parts, theme.Bad.Render("Blocked"))
	}
	if user.Muting {
		parts = append(parts, theme.Warn.Render("Muted"))
	}
	if user.Following {
		parts = append(parts, theme.Active.Render("✓ Following"))
	}
	if user.FollowedBy {
		parts = append(parts, theme.Dim.Render("Follows you"))
	}

	return strings.Join(parts, theme.Dim.Render(" · "))
}

func detailItems(theme ui.Theme, user north.User) []string {
	details := make([]string, 0, 3)
	if user.Location != nil && strings.TrimSpace(*user.Location) != "" {
		details = append(details, theme.Accent.Render("⌖")+" "+theme.Dim.Render(ui.SafeInline(*user.Location)))
	}
	if user.Website != nil && strings.TrimSpace(*user.Website) != "" {
		details = append(details, theme.Accent.Render("↗")+" "+theme.Handle.Render(ui.SafeInline(*user.Website)))
	}
	if user.CreatedAt != nil && !user.CreatedAt.IsZero() {
		joined := "Joined " + user.CreatedAt.Local().Format("January 2006")
		details = append(details, theme.Accent.Render("◷")+" "+theme.Dim.Render(joined))
	}

	return details
}

func packDetails(details []string, width int) []string {
	var lines []string
	row := ""
	for _, detail := range details {
		if row == "" {
			row = detail
			continue
		}
		if lipgloss.Width(row)+3+lipgloss.Width(detail) <= width {
			row += "   " + detail
			continue
		}
		lines = append(lines, ui.Clip(row, width))
		row = detail
	}
	if row != "" {
		lines = append(lines, ui.Clip(row, width))
	}

	return lines
}
