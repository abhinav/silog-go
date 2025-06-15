package silog

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
)

// Style defines the output styling for the logger.
type Style struct {
	Key lipgloss.Style

	KeyValueDelimiter lipgloss.Style                // required
	LevelLabels       map[slog.Level]lipgloss.Style // required
	MultilinePrefix   lipgloss.Style                // required
	PrefixDelimiter   lipgloss.Style                // required

	Messages map[slog.Level]lipgloss.Style
	Values   map[string]lipgloss.Style
}

// TODO: lipgloss.Renderer-based style

// DefaultStyle returns the default style for the logger.
func DefaultStyle() *Style {
	return &Style{
		Key:               lipgloss.NewStyle().Faint(true),
		KeyValueDelimiter: lipgloss.NewStyle().SetString("=").Faint(true),
		MultilinePrefix:   lipgloss.NewStyle().SetString("| ").Faint(true),
		PrefixDelimiter:   lipgloss.NewStyle().SetString(": "),
		LevelLabels: map[slog.Level]lipgloss.Style{
			slog.LevelDebug: lipgloss.NewStyle().SetString("DBG"),                                  // default
			slog.LevelInfo:  lipgloss.NewStyle().SetString("INF").Foreground(lipgloss.Color("10")), // green
			slog.LevelWarn:  lipgloss.NewStyle().SetString("WRN").Foreground(lipgloss.Color("11")), // yellow
			slog.LevelError: lipgloss.NewStyle().SetString("ERR").Foreground(lipgloss.Color("9")),  // red
		},
		Messages: map[slog.Level]lipgloss.Style{
			slog.LevelDebug: lipgloss.NewStyle().Faint(true),
		},
		Values: map[string]lipgloss.Style{
			"error": lipgloss.NewStyle().Foreground(lipgloss.Color("9")), // red
		},
	}
}

// PlainStyle returns a style for the logger without any colors.
func PlainStyle() *Style {
	return &Style{
		KeyValueDelimiter: lipgloss.NewStyle().SetString("="),
		MultilinePrefix:   lipgloss.NewStyle().SetString("  | "),
		PrefixDelimiter:   lipgloss.NewStyle().SetString(": "),
		LevelLabels: map[slog.Level]lipgloss.Style{
			slog.LevelDebug: lipgloss.NewStyle().SetString("DBG"),
			slog.LevelInfo:  lipgloss.NewStyle().SetString("INF"),
			slog.LevelWarn:  lipgloss.NewStyle().SetString("WRN"),
			slog.LevelError: lipgloss.NewStyle().SetString("ERR"),
		},
		Messages: map[slog.Level]lipgloss.Style{},
		Values:   map[string]lipgloss.Style{},
	}
}
