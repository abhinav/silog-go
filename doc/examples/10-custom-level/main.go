// # Custom Level
//
// A custom log level can be defined with its own label and message style.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss"
	"go.abhg.dev/log/silog"
)

const LevelTrace = slog.LevelDebug - 4

func main() {
	renderer := lipgloss.DefaultRenderer()
	style := silog.DefaultStyle(renderer)
	// <EXAMPLE>
	style.LevelLabels[LevelTrace] = renderer.NewStyle().SetString("TRC")
	style.Messages[LevelTrace] = style.Messages[slog.LevelDebug]
	// </EXAMPLE>

	handler := silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Style: style,
		Level: LevelTrace,
	})
	logger := slog.New(handler)

	// <EXAMPLE>
	logger.Log(context.Background(), LevelTrace, "Entering function")
	// </EXAMPLE>
	logger.Debug("Processing data")
}
