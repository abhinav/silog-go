package silog_test

import (
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"go.abhg.dev/log/silog"
)

func TestHandler_formatting(t *testing.T) {
	var buffer strings.Builder
	handler := silog.NewHandler(&buffer, &silog.HandlerOptions{
		Level: slog.LevelDebug,
		Style: silog.PlainStyle(nil),
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				// Use a fixed time for deterministic output.
				return slog.Time(slog.TimeKey, time.Date(2025, 0o6, 15, 9, 45, 0, 0, time.UTC))
			}
			return attr
		},
	})
	log := slog.New(handler)

	assertLinesWithTime := func(t *testing.T, lines ...string) bool {
		t.Helper()

		defer buffer.Reset()

		want := strings.Join(lines, "\n") + "\n"
		got := buffer.String()
		return assert.Equal(t, want, got)
	}

	t.Run("Message", func(t *testing.T) {
		log.Info("foo")
		assertLinesWithTime(t, "9:45AM INF foo")
	})

	t.Run("Levels", func(t *testing.T) {
		tests := []struct {
			name  string
			logFn func(string, ...any)
			want  string
		}{
			{"debug", log.Debug, "9:45AM DBG hello"},
			{"info", log.Info, "9:45AM INF hello"},
			{"warn", log.Warn, "9:45AM WRN hello"},
			{"error", log.Error, "9:45AM ERR hello"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.logFn("hello")
				assertLinesWithTime(t, tt.want)
			})
		}
	})

	t.Run("Attrs", func(t *testing.T) {
		someDate := time.Date(2025, 5, 20, 21, 0, 0, 0, time.UTC)

		tests := []struct {
			name  string
			value any
			want  string
		}{
			{"Bool", true, "true"},
			{"Duration", time.Second, "1s"},
			{"Float64", 3.14, "3.14"},
			{"Int64", int64(42), "42"},
			{"String", "foo", "foo"},
			{"Time", someDate, "9:00PM"},
			{"Uint64", uint64(42), "42"},
			{"Stringer", &testStringer{"foo"}, "foo"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				log.Info("foo", "k1", tt.value)
				assertLinesWithTime(t, "9:45AM INF foo  k1="+tt.want)
			})
		}
	})

	t.Run("EmptyAttr", func(t *testing.T) {
		log.Info("foo", slog.Attr{}, "foo", "bar")
		assertLinesWithTime(t, "9:45AM INF foo  foo=bar")
	})

	t.Run("MultilineMessage", func(t *testing.T) {
		log.Info("foo\nbar\nbaz")
		assertLinesWithTime(t,
			"9:45AM INF foo",
			"9:45AM INF bar",
			"9:45AM INF baz",
		)
	})

	t.Run("MultilineMessageWithLeadingSpaces", func(t *testing.T) {
		var s strings.Builder
		s.WriteString("foo\n")
		s.WriteString("  bar\n")
		s.WriteString("    baz\n")
		s.WriteString("qux\n")
		log.Info(s.String())

		assertLinesWithTime(t,
			"9:45AM INF foo",
			"9:45AM INF   bar",
			"9:45AM INF     baz",
			"9:45AM INF qux",
		)
	})

	t.Run("WithAttrs", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2, "k3", 3.0, "k4", "foo")
		log.Info("bar")

		assertLinesWithTime(t, "9:45AM INF bar  k1=true k2=2 k3=3 k4=foo")
	})

	t.Run("WithAttrsEmpty", func(t *testing.T) {
		log := log.With()
		log.Info("bar")
		assertLinesWithTime(t, "9:45AM INF bar")
	})

	t.Run("MultilineMessageWithAttrs", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2, "k3", 3.0, "k4", "foo")
		log.Info("bar\nbaz")

		assertLinesWithTime(t,
			"9:45AM INF bar",
			"9:45AM INF baz  k1=true k2=2 k3=3 k4=foo",
		)
	})

	t.Run("MultilineMessageWithAttrsAndLeadingNewline", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2, "k3", 3.0, "k4", "foo")
		log.Info("bar\nbaz\n")

		assertLinesWithTime(t,
			"9:45AM INF bar",
			"9:45AM INF baz",
			"  k1=true k2=2 k3=3 k4=foo",
		)
	})

	t.Run("Attrs", func(t *testing.T) {
		log.Info("foo", "k1", true, "k2", 2, "k3", 3.0, "k4", "bar")
		assertLinesWithTime(t, "9:45AM INF foo  k1=true k2=2 k3=3 k4=bar")
	})

	t.Run("AttrsWithAttrs", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2)
		log.Info("foo", "k3", 3.0, "k4", "bar")
		log.Warn("baz", "k5", 5.0, "k6", "qux")

		assertLinesWithTime(t,
			"9:45AM INF foo  k1=true k2=2 k3=3 k4=bar",
			"9:45AM WRN baz  k1=true k2=2 k5=5 k6=qux",
		)
	})

	t.Run("WithGroup", func(t *testing.T) {
		log := log.WithGroup("g")
		log.Info("foo", "k1", true, "k2", 2, "k3", 3.0, "k4", "bar")
		assertLinesWithTime(t, "9:45AM INF foo  g.k1=true g.k2=2 g.k3=3 g.k4=bar")
	})

	t.Run("WithGroupEmpty", func(t *testing.T) {
		log := log.WithGroup("")
		log.Info("foo", "k1", true)
		assertLinesWithTime(t, "9:45AM INF foo  k1=true")
	})

	t.Run("WithGroupWithAttrs", func(t *testing.T) {
		log := log.WithGroup("g").With("k1", true, "k2", 2)
		log.Info("foo", "k3", 3.0, "k4", "bar")
		log.Warn("baz", "k5", 5.0, "k6", "qux")

		assertLinesWithTime(t,
			"9:45AM INF foo  g.k1=true g.k2=2 g.k3=3 g.k4=bar",
			"9:45AM WRN baz  g.k1=true g.k2=2 g.k5=5 g.k6=qux",
		)
	})

	t.Run("AttrGroup", func(t *testing.T) {
		log.Info("foo", slog.Group("bar", "k1", true, "k2", 2, "k3", 3.0, "k4", "bar"))
		assertLinesWithTime(t, "9:45AM INF foo  bar.k1=true bar.k2=2 bar.k3=3 bar.k4=bar")
	})

	t.Run("AttrGroupEmptyAttr", func(t *testing.T) {
		log.Info("foo", slog.Group("bar", slog.Attr{}, "k1", true, "k2", 2, "k3", 3.0, "k4", "bar")) //nolint:loggercheck
		assertLinesWithTime(t, "9:45AM INF foo  bar.k1=true bar.k2=2 bar.k3=3 bar.k4=bar")
	})

	t.Run("TrailingNewlineMessage", func(t *testing.T) {
		log.Info("foo\n")
		assertLinesWithTime(t, "9:45AM INF foo")
	})

	t.Run("TrailingNewlineMessageWithAttr", func(t *testing.T) {
		log.Info("foo\n", "k1", true)
		assertLinesWithTime(t,
			"9:45AM INF foo",
			"  k1=true",
		)
	})

	t.Run("MultilineAttrValue", func(t *testing.T) {
		log.Info("foo", "k1", "bar\nbaz\nqux", "k2", "quux")
		assertLinesWithTime(t,
			"9:45AM INF foo  ",
			"  k1=",
			"    | bar",
			"    | baz",
			"    | qux",
			"  k2=quux",
		)
	})

	t.Run("MultlineAttrValueNestedInGroup", func(t *testing.T) {
		log := log.WithGroup("a").WithGroup("b")
		log.Info("foo", slog.Group("c", "d", "foo\nbar\nbaz", "e", "qux"))

		assertLinesWithTime(t,
			"9:45AM INF foo  ",
			"  a.b.c.d=",
			"    | foo",
			"    | bar",
			"    | baz",
			"  a.b.c.e=qux",
		)
	})

	t.Run("LeadingWhitespace", func(t *testing.T) {
		log.Info(" foo")
		assertLinesWithTime(t, "9:45AM INF  foo")
	})

	t.Run("TrailingWhitespace", func(t *testing.T) {
		log.Info("foo ")
		assertLinesWithTime(t, "9:45AM INF foo")
	})

	t.Run("Prefix", func(t *testing.T) {
		log := slog.New(handler.SetPrefix("prefix"))

		log.Info("foo")
		assertLinesWithTime(t, "9:45AM INF prefix: foo")
	})

	t.Run("MultilineMessageWithPrefix", func(t *testing.T) {
		log := slog.New(handler.SetPrefix("prefix"))

		log.Info("foo\nbar\nbaz")
		assertLinesWithTime(t,
			"9:45AM INF prefix: foo",
			"9:45AM INF prefix: bar",
			"9:45AM INF prefix: baz",
		)
	})

	t.Run("WithLevelOffset", func(t *testing.T) {
		downLog := slog.New(handler.WithLevelOffset(-4))

		downLog.Debug("foo")
		downLog.Info("bar")
		downLog.Warn("baz")
		downLog.Error("qux")

		assertLinesWithTime(t,
			"9:45AM DBG bar",
			"9:45AM INF baz",
			"9:45AM WRN qux")

		log.Debug("quux")
		assertLinesWithTime(t,
			"9:45AM DBG quux")

		t.Run("Undo", func(t *testing.T) {
			upLog := slog.New(downLog.Handler().(*silog.Handler).WithLevelOffset(4))

			upLog.Debug("foo")
			upLog.Info("bar")
			upLog.Warn("baz")
			upLog.Error("qux")

			assertLinesWithTime(t,
				"9:45AM DBG foo",
				"9:45AM INF bar",
				"9:45AM WRN baz",
				"9:45AM ERR qux")
		})
	})
}

func TestHandler_WithLevel(t *testing.T) {
	var buffer strings.Builder
	handler := silog.NewHandler(&buffer, &silog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
		Style: silog.PlainStyle(nil),
	})
	rootLogger := slog.New(handler)

	rootLogger.Debug("foo")
	assert.Empty(t, buffer.String())

	debugLogger := slog.New(handler.WithLevel(slog.LevelDebug))
	debugLogger.Debug("foo")
	assert.Equal(t, "DBG foo\n", buffer.String())
	buffer.Reset()

	rootLogger.Debug("foo")
	assert.Empty(t, buffer.String())
}

func TestHandler_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		leveler  slog.Leveler
		enabled  []slog.Level
		disabled []slog.Level
	}{
		{
			name:    "debug",
			leveler: slog.LevelDebug,
			enabled: []slog.Level{slog.LevelDebug, slog.LevelInfo},
		},
		{
			name:     "info",
			leveler:  slog.LevelInfo,
			enabled:  []slog.Level{slog.LevelInfo, slog.LevelWarn},
			disabled: []slog.Level{slog.LevelDebug},
		},
		{
			name:     "warn",
			leveler:  slog.LevelWarn,
			enabled:  []slog.Level{slog.LevelWarn, slog.LevelError},
			disabled: []slog.Level{slog.LevelDebug, slog.LevelInfo},
		},
		{
			name:     "error",
			leveler:  slog.LevelError,
			enabled:  []slog.Level{slog.LevelError, slog.LevelError + 4},
			disabled: []slog.Level{slog.LevelInfo, slog.LevelWarn},
		},
		{
			name:     "fatal",
			leveler:  slog.LevelError + 4,
			enabled:  []slog.Level{slog.LevelError + 4},
			disabled: []slog.Level{slog.LevelWarn, slog.LevelError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := silog.NewHandler(io.Discard, &silog.HandlerOptions{
				Level: tt.leveler,
				Style: silog.PlainStyle(nil),
			})
			for _, level := range tt.enabled {
				assert.True(t, h.Enabled(t.Context(), level.Level()), "level %s should be enabled", level)
			}
			for _, level := range tt.disabled {
				assert.False(t, h.Enabled(t.Context(), level.Level()), "level %s should be disabled", level)
			}
		})
	}
}

func TestHandler_withAttrsConcurrent(t *testing.T) {
	const (
		NumWorkers = 10
		NumWrites  = 100
	)

	var buffer strings.Builder
	log := slog.New(silog.NewHandler(&buffer, &silog.HandlerOptions{
		Level: slog.LevelDebug,
		Style: silog.PlainStyle(nil),
	}))

	var ready, done sync.WaitGroup
	ready.Add(NumWorkers)
	for range NumWorkers {
		done.Add(1)
		go func() {
			defer done.Done()

			ready.Done()
			ready.Wait()

			for i := range NumWrites {
				log.Info("message", "i", i)
			}
		}()
	}

	done.Wait()

	assert.Equal(t, NumWorkers*NumWrites, strings.Count(buffer.String(), "INF message"))
}

func TestHandler_multilineMessageStyling(t *testing.T) {
	renderer := lipgloss.NewRenderer(nil, termenv.WithUnsafe())
	renderer.SetColorProfile(termenv.ANSI)

	style := silog.PlainStyle(renderer)
	style.Messages[slog.LevelInfo] = renderer.NewStyle().Bold(true)

	var buffer strings.Builder
	log := slog.New(silog.NewHandler(&buffer, &silog.HandlerOptions{
		Level: slog.LevelDebug,
		Style: style,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
	}))

	log.Info("foo\nbar")

	assert.Equal(t,
		"INF \x1b[1mfoo\x1b[0m\n"+
			"INF \x1b[1mbar\x1b[0m\n",
		buffer.String())
}

type testStringer struct{ v string }

func (s *testStringer) String() string { return s.v }
