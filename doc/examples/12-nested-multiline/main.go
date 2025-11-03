// # Nested Multiline
//
// Multi-line attribute values within deeply nested groups,
// show how formatting is preserved through multiple levels.
package main

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

func main() {
	baseLogger := slog.New(silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Style: silog.DefaultStyle(nil),
		Level: slog.LevelDebug,
	}))

	// <EXAMPLE>
	logger := baseLogger.WithGroup("service").WithGroup("database")

	logger.Info("Query executed",
		slog.Group("details",
			"query", "SELECT id, name\nFROM users\nORDER BY created_at",
			"rows", 42))
	// </EXAMPLE>
}
