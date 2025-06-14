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
	handler := silog.NewHandler(&buffer, &silog.Options{
		Level: slog.LevelDebug,
	})
	log := slog.New(handler)

	assertLines := func(t *testing.T, lines ...string) bool {
		t.Helper()

		defer buffer.Reset()

		want := strings.Join(lines, "\n") + "\n"
		got := buffer.String()
		return assert.Equal(t, want, got)
	}

	t.Run("Message", func(t *testing.T) {
		log.Info("foo")
		assertLines(t, "INF foo")
	})

	t.Run("Levels", func(t *testing.T) {
		tests := []struct {
			name  string
			logFn func(string, ...any)
			want  string
		}{
			{"debug", log.Debug, "DBG hello"},
			{"info", log.Info, "INF hello"},
			{"warn", log.Warn, "WRN hello"},
			{"error", log.Error, "ERR hello"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.logFn("hello")
				assertLines(t, tt.want)
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
				assertLines(t, "INF foo  k1="+tt.want)
			})
		}
	})

	t.Run("EmptyAttr", func(t *testing.T) {
		log.Info("foo", slog.Attr{}, "foo", "bar")
		assertLines(t, "INF foo  foo=bar")
	})

	t.Run("MultilineMessage", func(t *testing.T) {
		log.Info("foo\nbar\nbaz")
		assertLines(t,
			"INF foo",
			"INF bar",
			"INF baz",
		)
	})

	t.Run("MultilineMessageWithLeadingSpaces", func(t *testing.T) {
		var s strings.Builder
		s.WriteString("foo\n")
		s.WriteString("  bar\n")
		s.WriteString("    baz\n")
		s.WriteString("qux\n")
		log.Info(s.String())

		assertLines(t,
			"INF foo",
			"INF   bar",
			"INF     baz",
			"INF qux",
		)
	})

	t.Run("WithAttrs", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2, "k3", 3.0, "k4", "foo")
		log.Info("bar")

		assertLines(t, "INF bar  k1=true k2=2 k3=3 k4=foo")
	})

	t.Run("WithAttrsEmpty", func(t *testing.T) {
		log := log.With()
		log.Info("bar")
		assertLines(t, "INF bar")
	})

	t.Run("MultilineMessageWithAttrs", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2, "k3", 3.0, "k4", "foo")
		log.Info("bar\nbaz")

		assertLines(t,
			"INF bar",
			"INF baz  k1=true k2=2 k3=3 k4=foo",
		)
	})

	t.Run("MultilineMessageWithAttrsAndLeadingNewline", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2, "k3", 3.0, "k4", "foo")
		log.Info("bar\nbaz\n")

		assertLines(t,
			"INF bar",
			"INF baz",
			"  k1=true k2=2 k3=3 k4=foo",
		)
	})

	t.Run("Attrs", func(t *testing.T) {
		log.Info("foo", "k1", true, "k2", 2, "k3", 3.0, "k4", "bar")
		assertLines(t, "INF foo  k1=true k2=2 k3=3 k4=bar")
	})

	t.Run("AttrsWithAttrs", func(t *testing.T) {
		log := log.With("k1", true, "k2", 2)
		log.Info("foo", "k3", 3.0, "k4", "bar")
		log.Warn("baz", "k5", 5.0, "k6", "qux")

		assertLines(t,
			"INF foo  k1=true k2=2 k3=3 k4=bar",
			"WRN baz  k1=true k2=2 k5=5 k6=qux",
		)
	})

	t.Run("WithGroup", func(t *testing.T) {
		log := log.WithGroup("g")
		log.Info("foo", "k1", true, "k2", 2, "k3", 3.0, "k4", "bar")
		assertLines(t, "INF foo  g.k1=true g.k2=2 g.k3=3 g.k4=bar")
	})

	t.Run("WithGroupEmpty", func(t *testing.T) {
		log := log.WithGroup("")
		log.Info("foo", "k1", true)
		assertLines(t, "INF foo  k1=true")
	})

	t.Run("WithGroupWithAttrs", func(t *testing.T) {
		log := log.WithGroup("g").With("k1", true, "k2", 2)
		log.Info("foo", "k3", 3.0, "k4", "bar")
		log.Warn("baz", "k5", 5.0, "k6", "qux")

		assertLines(t,
			"INF foo  g.k1=true g.k2=2 g.k3=3 g.k4=bar",
			"WRN baz  g.k1=true g.k2=2 g.k5=5 g.k6=qux",
		)
	})

	t.Run("AttrGroup", func(t *testing.T) {
		log.Info("foo", slog.Group("bar", "k1", true, "k2", 2, "k3", 3.0, "k4", "bar"))
		assertLines(t, "INF foo  bar.k1=true bar.k2=2 bar.k3=3 bar.k4=bar")
	})

	t.Run("AttrGroupEmptyAttr", func(t *testing.T) {
		log.Info("foo", slog.Group("bar", slog.Attr{}, "k1", true, "k2", 2, "k3", 3.0, "k4", "bar")) //nolint:loggercheck
		assertLines(t, "INF foo  bar.k1=true bar.k2=2 bar.k3=3 bar.k4=bar")
	})

	t.Run("TrailingNewlineMessage", func(t *testing.T) {
		log.Info("foo\n")
		assertLines(t, "INF foo")
	})

	t.Run("TrailingNewlineMessageWithAttr", func(t *testing.T) {
		log.Info("foo\n", "k1", true)
		assertLines(t,
			"INF foo",
			"  k1=true",
		)
	})

	t.Run("MultilineAttrValue", func(t *testing.T) {
		log.Info("foo", "k1", "bar\nbaz\nqux", "k2", "quux")
		assertLines(t,
			"INF foo  ",
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

		assertLines(t,
			"INF foo  ",
			"  a.b.c.d=",
			"    | foo",
			"    | bar",
			"    | baz",
			"  a.b.c.e=qux",
		)
	})

	t.Run("LeadingWhitespace", func(t *testing.T) {
		log.Info(" foo")
		assertLines(t, "INF  foo")
	})

	t.Run("TrailingWhitespace", func(t *testing.T) {
		log.Info("foo ")
		assertLines(t, "INF foo")
	})

	t.Run("Prefix", func(t *testing.T) {
		log := slog.New(handler.WithPrefix("prefix"))

		log.Info("foo")
		assertLines(t, "INF prefix: foo")
	})

	t.Run("MultilineMessageWithPrefix", func(t *testing.T) {
		log := slog.New(handler.WithPrefix("prefix"))

		log.Info("foo\nbar\nbaz")
		assertLines(t,
			"INF prefix: foo",
			"INF prefix: bar",
			"INF prefix: baz",
		)
	})

	t.Run("Downgrade", func(t *testing.T) {
		downLog := slog.New(handler.WithDowngrade(4))

		downLog.Debug("foo")
		downLog.Info("bar")
		downLog.Warn("baz")
		downLog.Error("qux")

		assertLines(t,
			"DBG bar",
			"INF baz",
			"WRN qux")

		log.Debug("quux")
		assertLines(t,
			"DBG quux")
	})
}

func TestHandler_WithLevel(t *testing.T) {
	var buffer strings.Builder
	handler := silog.NewHandler(&buffer, nil)
	rootLogger := slog.New(handler)

	rootLogger.Debug("foo")
	assert.Empty(t, buffer.String())

	debugLogger := slog.New(handler.WithLeveler(slog.LevelDebug))
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
			h := silog.NewHandler(io.Discard, &silog.Options{
				Level: tt.leveler,
				Style: silog.PlainStyle(),
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
	log := slog.New(silog.NewHandler(&buffer, &silog.Options{
		Level: slog.LevelDebug,
		Style: silog.PlainStyle(),
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
	// Force colored output even if the terminal doesn't support it.
	t.Setenv("CLICOLOR_FORCE", "1")
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.EnvColorProfile())

	style := silog.PlainStyle()
	style.Messages[slog.LevelInfo] = lipgloss.NewStyle().Bold(true)

	var buffer strings.Builder
	log := slog.New(silog.NewHandler(&buffer, &silog.Options{
		Level: slog.LevelDebug,
		Style: style,
	}))

	log.Info("foo\nbar")

	assert.Equal(t,
		"INF \x1b[1mfoo\x1b[0m\n"+
			"INF \x1b[1mbar\x1b[0m\n",
		buffer.String())
}

type testStringer struct{ v string }

func (s *testStringer) String() string { return s.v }
