// # Multiline Errors
//
// Multi-line error messages are formatted
// with each line receiving the timestamp and level prefix.
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
	logger.Error("Database connection failed:\nConnection refused\n  at db.Connect()\n  at main.startup()")
	// </EXAMPLE>
}
