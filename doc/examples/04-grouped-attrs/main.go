// # Grouped Attributes
//
// Nested grouped attributes are created using WithGroup and slog.Group,
// rendered with dot-notation for hierarchical attribute organization.
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

	// <EXAMPLE>
	logger := slog.New(handler).WithGroup("request").WithGroup("headers")

	logger.Info("Incoming request",
		slog.Group("body",
			"method", "POST",
			"path", "/api/users",
		),
	)
	// </EXAMPLE>
}
