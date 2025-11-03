// # No Time
//
// ReplaceAttr can suppress the timestamp from log output.
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
		// <EXAMPLE>
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			// Remove the time attribute from log output.
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
		// </EXAMPLE>
	})
	logger := slog.New(handler)

	logger.Debug("Starting background task")
	logger.Info("Server listening on :8080")
	logger.Warn("Connection pool nearing capacity")
	logger.Error("Failed to process request")
}
