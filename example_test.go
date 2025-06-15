package silog_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss"
	"go.abhg.dev/log/silog"
)

// Demonstrates how to introduce a new log level to the logger.
func Example_customLevel() {
	const LevelTrace = slog.LevelDebug - 4

	style := silog.PlainStyle()
	style.LevelLabels[LevelTrace] = style.LevelLabels[slog.LevelDebug].SetString("TRC")
	style.Messages[LevelTrace] = style.Messages[slog.LevelDebug]

	handler := silog.NewHandler(os.Stdout, &silog.HandlerOptions{
		Style: style,
		Level: LevelTrace,
		// To keep the test output clean easy to test,
		// we will not log the time in this example.
		ReplaceAttr: skipTime,
	})

	logger := slog.New(handler)
	logger.Log(context.Background(), LevelTrace, "This is a trace message")
	logger.Debug("This is a debug message")

	// Output:
	// TRC This is a trace message
	// DBG This is a debug message
}

// Demonstrates reserving a log level to be logged without a label before it.
func Example_noLabel() {
	const LevelPlain = slog.LevelDebug - 1

	style := silog.PlainStyle()
	style.LevelLabels[LevelPlain] = lipgloss.NewStyle() // No label
	style.Messages[LevelPlain] = style.Messages[slog.LevelDebug]

	handler := silog.NewHandler(os.Stdout, &silog.HandlerOptions{
		Style: style,
		Level: LevelPlain,
		// To keep the test output clean easy to test,
		// we will not log the time in this example.
		ReplaceAttr: skipTime,
	})

	logger := slog.New(handler)
	logger.Log(context.Background(), LevelPlain, "This is a plain message")
	logger.Debug("This is a debug message")

	// Output:
	// This is a plain message
	// DBG This is a debug message
}

func ExampleHandler_SetPrefix() {
	handler := silog.NewHandler(os.Stdout, &silog.HandlerOptions{
		Style: silog.PlainStyle(),
		// To keep the test output clean easy to test,
		// we will not log the time in this example.
		ReplaceAttr: skipTime,
	})

	h1 := handler.SetPrefix("example")
	h2 := h1.SetPrefix("")

	slog.New(h1).Info("Single prefix")
	slog.New(h2).Info("No prefix")

	// Output:
	// INF example: Single prefix
	// INF No prefix
}

func ExampleHandler_WithLevelOffset() {
	handler := silog.NewHandler(os.Stdout, &silog.HandlerOptions{
		Level: slog.LevelDebug,
		Style: silog.PlainStyle(),
		// To keep the test output clean easy to test,
		// we will not log the time in this example.
		ReplaceAttr: skipTime,
	})

	logger := slog.New(handler.WithLevelOffset(-4)) // downgrade messages
	logger.Error("Downgraded to warning")
	logger.Warn("Downgraded to info")
	logger.Info("Downgraded to debug")
	logger.Debug("Not logged because below configured level")

	// Output:
	// WRN Downgraded to warning
	// INF Downgraded to info
	// DBG Downgraded to debug
}

func skipTime(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}
