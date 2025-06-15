package silog

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
)

// Style defines the output styling for the logger.
//
// Any fields of style that are not set will use the default style
// (no color, no special formatting).
//
// # Customization
//
// It's best to construct a style with [DefaultStyle] or [PlainStyle]
// and then modify the fields you want to change.
// For example:
//
//	style := silog.DefaultStyle()
//	style.KeyValueDelimiter = lipgloss.NewStyle().SetString(": ")
type Style struct {
	// Key is the style used for the key in key-value pairs.
	Key lipgloss.Style

	// KeyValueDelimiter is the style used for the delimiter
	// separating keys and values in key-value pairs.
	//
	// This SHOULD have a non-empty value (e.g. "=", ": ", etc.)
	// set with lipgloss.Style.SetString.
	//
	// The default value is "=".
	KeyValueDelimiter lipgloss.Style

	// LevelLabels is a map of slog.Level to style
	// for the label of that level.
	//
	// Each style here SHOULD have a non-empty value
	// (e.g. "DBG", "INF", etc.) set with lipgloss.Style.SetString.
	//
	// If a record has a level that is not present in this map,
	// messages of that level will not be labeled.
	LevelLabels map[slog.Level]lipgloss.Style

	// MultilineValuePrefix defines the style for the prefix that is
	// prepended to each line of an indented multi-line attribute value.
	//
	// This SHOULD have a non-empty value (e.g. "| ").
	// The default value is "| ".
	MultilineValuePrefix lipgloss.Style

	// PrefixDelimiter defines the style separating a prefix
	// (specified with Handler.WithPrefix) from the rest of the log message.
	//
	// This SHOULD have a non-empty value (e.g. ": ", " - ", etc.)
	// The default value is ": ".
	PrefixDelimiter lipgloss.Style

	// Time defines the style used for the time of a log record.
	//
	// If ReplaceAttr is used to change the time attribute,
	// the style is also used for the replacement value.
	Time lipgloss.Style

	// Messages defines styling for messages logged at different levels.
	//
	// If a log record has a level that is not present in this map,
	// the message will use plain text style.
	Messages map[slog.Level]lipgloss.Style

	// Values defines the styling for attributes matched by their keys.
	// Attributes with keys that are not present in this map
	// will use a plain text style for their values.
	//
	// DefaultStyle uses this to style the "error" key in red.
	Values map[string]lipgloss.Style
}

// TODO: lipgloss.Renderer-based style

// DefaultStyle is the default style used by [Handler].
// It provides colored output, faint text for debug messages, red errors, etc.
func DefaultStyle() *Style {
	return &Style{
		Key:                  lipgloss.NewStyle().Faint(true),
		KeyValueDelimiter:    lipgloss.NewStyle().SetString("=").Faint(true),
		MultilineValuePrefix: lipgloss.NewStyle().SetString("| ").Faint(true),
		PrefixDelimiter:      lipgloss.NewStyle().SetString(": "),
		Time:                 lipgloss.NewStyle().Faint(true),
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

// PlainStyle is a style for [Handler] that performs no styling.
// It's best when writing to a non-terminal destination.
func PlainStyle() *Style {
	return &Style{
		KeyValueDelimiter:    lipgloss.NewStyle().SetString("="),
		MultilineValuePrefix: lipgloss.NewStyle().SetString("  | "),
		Time:                 lipgloss.NewStyle(),
		PrefixDelimiter:      lipgloss.NewStyle().SetString(": "),
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
