// # Basic Levels
//
// Built-in levels are demonstrated with their default color-coded styling.
package main

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

func main() {
	handler := silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Level: slog.LevelDebug,
		Style: silog.DefaultStyle(nil),
	})
	logger := slog.New(handler)

	// <EXAMPLE>
	logger.Debug("Starting background task")
	logger.Info("Server listening on :8080")
	logger.Warn("Connection pool nearing capacity")
	logger.Error("Failed to process request")
	// </EXAMPLE>
}
