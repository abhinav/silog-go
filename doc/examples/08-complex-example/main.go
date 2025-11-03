// # Complex Example
//
// A complex real-world example demonstrating combining
// multi-line error attributes,
// structured key-value pairs,
// and error highlighting.
package main

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

func main() {
	h := silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Style: silog.DefaultStyle(nil),
		Level: slog.LevelDebug,
	})
	log := slog.New(h)

	// <EXAMPLE>
	log.Error("API request failed",
		"method", "POST",
		"path", "/api/orders",
		"error", "validation failed:\n  - invalid email\n  - missing required field: address",
		"user_id", "12345",
		"duration", "125ms",
	)
	// </EXAMPLE>
}
