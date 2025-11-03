// # Multiline Attributes
//
// Multi-line attribute values are formatted
// with pipe-prefixed indentation for readability.
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
	logger.Info("Request completed",
		"sql", "SELECT *\nFROM users\nWHERE active = true",
		"duration", "45ms",
	)
	// </EXAMPLE>
}
