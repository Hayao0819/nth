package ui

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/charmbracelet/x/ansi"
)

// These pairs occupy two cells in terminals without Unicode core mode, but
// grapheme-based layout counts them as one. The composed forms count as two in
// both modes.
var halfwidthKanaVoicing = strings.NewReplacer(
	"ｳﾞ", "ヴ",
	"ｶﾞ", "ガ", "ｷﾞ", "ギ", "ｸﾞ", "グ", "ｹﾞ", "ゲ", "ｺﾞ", "ゴ",
	"ｻﾞ", "ザ", "ｼﾞ", "ジ", "ｽﾞ", "ズ", "ｾﾞ", "ゼ", "ｿﾞ", "ゾ",
	"ﾀﾞ", "ダ", "ﾁﾞ", "ヂ", "ﾂﾞ", "ヅ", "ﾃﾞ", "デ", "ﾄﾞ", "ド",
	"ﾊﾞ", "バ", "ﾋﾞ", "ビ", "ﾌﾞ", "ブ", "ﾍﾞ", "ベ", "ﾎﾞ", "ボ",
	"ﾊﾟ", "パ", "ﾋﾟ", "ピ", "ﾌﾟ", "プ", "ﾍﾟ", "ペ", "ﾎﾟ", "ポ",
	"ﾜﾞ", "ヷ", "ｲﾞ", "ヸ", "ｴﾞ", "ヹ", "ｦﾞ", "ヺ",
)

func SafeText(value string) string {
	value = ansi.Strip(value)
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\t", "    ")

	value = strings.Map(func(r rune) rune {
		if r == '\n' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, value)

	return halfwidthKanaVoicing.Replace(value)
}

func SafeInline(value string) string {
	return strings.Join(strings.Fields(SafeText(value)), " ")
}

func Clip(value string, width int) string {
	if width <= 0 {
		return ""
	}

	return ansi.Truncate(value, width, "…")
}

func Left(value string, width int) string {
	value = Clip(value, width)
	if gap := width - lipgloss.Width(value); gap > 0 {
		value += strings.Repeat(" ", gap)
	}

	return value
}

func Sides(leftValue, rightValue string, width int) string {
	if width <= 0 {
		return ""
	}
	if rightWidth := lipgloss.Width(rightValue); rightWidth >= width {
		return Clip(rightValue, width)
	} else if gap := width - lipgloss.Width(leftValue) - rightWidth; gap > 0 {
		return leftValue + strings.Repeat(" ", gap) + rightValue
	}

	rightWidth := lipgloss.Width(rightValue)

	return Clip(leftValue, max(0, width-rightWidth-1)) + " " + rightValue
}

func Columns(values []string, width int) string {
	if width <= 0 || len(values) == 0 {
		return ""
	}
	var result strings.Builder
	remaining := width
	for index, value := range values {
		cellWidth := remaining / (len(values) - index)
		remaining -= cellWidth
		result.WriteString(lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center).Render(Clip(value, cellWidth)))
	}

	return result.String()
}

func ColumnAt(x, width, count int) int {
	if width <= 0 || count <= 1 {
		return 0
	}
	x = min(max(x, 0), width-1)
	remaining := width
	left := 0
	for index := range count {
		cellWidth := remaining / (count - index)
		if x < left+cellWidth {
			return index
		}
		left += cellWidth
		remaining -= cellWidth
	}

	return count - 1
}

func TextAt(line string, x int, label string) bool {
	start := strings.Index(line, label)
	if start < 0 {
		return false
	}
	left := lipgloss.Width(line[:start])

	return x >= left && x < left+lipgloss.Width(label)
}

func Fit(value string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	return lipgloss.NewStyle().
		Width(width).Height(height).
		MaxWidth(width).MaxHeight(height).
		Render(value)
}

func WrappedLines(value string, width int) []string {
	if width <= 0 {
		return nil
	}
	value = strings.TrimSpace(SafeText(value))
	if value == "" {
		return nil
	}

	return strings.Split(ansi.Wrap(value, width, " "), "\n")
}

func FriendlyError(err error) string {
	if err == nil {
		return ""
	}

	var apiError *north.APIError
	if errors.As(err, &apiError) {
		switch {
		case apiError.HasCode(32):
			return "Authentication failed — check the API token or run nth setup"
		case apiError.HasCode(87):
			return "The API token does not have permission for this action"
		case apiError.HasCode(88):
			if apiError.Response != nil && !apiError.Response.RateLimit.Reset.IsZero() {
				return "Rate limited until " + apiError.Response.RateLimit.Reset.Local().Format("15:04:05")
			}

			return "Rate limit reached"
		case len(apiError.Errors) > 0:
			return SafeInline(apiError.Errors[0].Message)
		default:
			return fmt.Sprintf("North returned HTTP %d", apiError.StatusCode)
		}
	}

	var webError *unofficial.APIError
	if errors.As(err, &webError) {
		switch {
		case webError.StatusCode == 401:
			return "Browser session expired — sign in to north.rip and restart nth"
		case webError.StatusCode == 429:
			return "Rate limit reached"
		case webError.Message != "":
			return SafeInline(webError.Message)
		default:
			return fmt.Sprintf("North returned HTTP %d", webError.StatusCode)
		}
	}
	if errors.Is(err, unofficial.ErrNotAuthenticated) {
		return "Browser session expired — sign in to north.rip and restart nth"
	}

	return SafeInline(err.Error())
}
