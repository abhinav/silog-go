// # Multiline Prefixes
//
// Prefixes are preserved across multi-line messages,
// appearing on each line of output.
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
	// <EXAMPLE>
	log := slog.New(h.SetPrefix("worker"))

	log.Info("Task completed:\n- Processed 1000 items\n- Generated 50 reports\n- Sent 25 notifications")
	// </EXAMPLE>
}
