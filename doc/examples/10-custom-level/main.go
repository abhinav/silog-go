// # Custom Level
//
// A custom log level can be defined with its own label and message style.
package main

import (
	"context"
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

const LevelTrace = slog.LevelDebug - 4

func main() {
	style := silog.DefaultStyle()
	// <EXAMPLE>
	style.LevelLabels[LevelTrace] = style.LevelLabels[slog.LevelDebug].SetString("TRC")
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
