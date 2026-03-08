package silog

import (
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogHandler_slogtest(t *testing.T) {
	var buffer strings.Builder
	slogtest.Run(t, func(*testing.T) slog.Handler {
		buffer.Reset()

		return NewHandler(&buffer, &HandlerOptions{
			Level:      slog.LevelDebug,
			Style:      PlainStyle(),
			TimeFormat: time.RFC3339,
		})
	}, func(t *testing.T) map[string]any {
		attrs := make(map[string]any)

		line := strings.TrimSpace(buffer.String())

		timestr, line, ok := strings.Cut(line, " ")
		require.True(t, ok, "missing time delimiter: %q", buffer.String())

		var lvlstr string
		if ts, err := time.Parse(time.RFC3339, timestr); err == nil {
			attrs[slog.TimeKey] = ts

			lvlstr, line, ok = strings.Cut(line, lvlDelim)
			require.True(t, ok, "missing level delimiter: %q", buffer.String())
		} else {
			// There's no time if the time was a zero value
			// so use the timestr as the lvl string.
			lvlstr = timestr
		}

		switch lvlstr {
		case "DBG":
			attrs[slog.LevelKey] = slog.LevelDebug
		case "INF":
			attrs[slog.LevelKey] = slog.LevelInfo
		case "WRN":
			attrs[slog.LevelKey] = slog.LevelWarn
		case "ERR":
			attrs[slog.LevelKey] = slog.LevelError
		default:
			t.Fatalf("unknown level: %q", lvlstr)
		}

		attrs[slog.MessageKey], line, _ = strings.Cut(line, msgAttrDelim)

		for pair := range strings.SplitSeq(line, attrDelim) {
			if pair == "" {
				continue
			}
			key, value, ok := strings.Cut(pair, "=")
			require.True(t, ok, "missing attribute delimiter: %q", pair)

			curAttrs := attrs
			for len(key) > 0 {
				groupKey, valKey, ok := strings.Cut(key, groupDelim)
				if !ok {
					// No more groups.
					curAttrs[key] = value
					break
				}

				groupAttrs, ok := curAttrs[groupKey].(map[string]any)
				if !ok {
					groupAttrs = make(map[string]any)
					curAttrs[groupKey] = groupAttrs
				}
				curAttrs = groupAttrs
				key = valKey
			}
		}

		t.Logf("buffer: %q", buffer.String())
		t.Logf("attrs: %q", attrs)
		return attrs
	})
}
