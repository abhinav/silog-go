package silog_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"go.abhg.dev/log/silog"
)

func TestDefaultStyle(t *testing.T) {
	assertHasValue := func(t *testing.T, style lipgloss.Style, msgArgs ...any) {
		t.Helper()

		assert.NotEmpty(t, strings.TrimSpace(style.Value()), msgArgs...)
	}

	style := silog.DefaultStyle(nil)

	assertHasValue(t, style.KeyValueDelimiter, "KeyValueDelimiter")
	assertHasValue(t, style.MultilineValuePrefix, "MultilinePrefix")
	assertHasValue(t, style.PrefixDelimiter, "PrefixDelimiter")

	defaultLevels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, lvl := range defaultLevels {
		t.Run(lvl.String(), func(t *testing.T) {
			assertHasValue(t, style.LevelLabels[lvl], "LevelLabels.Get(%s)", lvl)
		})
	}
}
