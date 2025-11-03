// # Error Highlighting
//
// The "error" attribute is automatically highlighted in red
// when using the default style.
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
	logger.Info("Operation failed",
		"error", "connection timeout",
		"retry_count", 3,
		"timeout", "30s",
	)
	// </EXAMPLE>
}
