package silog_test

import (
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.abhg.dev/log/silog"
)

func TestHandler_concurrentWrites(t *testing.T) {
	var buffer strings.Builder

	handler := silog.NewHandler(&buffer, &silog.HandlerOptions{
		Style: silog.PlainStyle(),
	})
	logger := slog.New(handler)

	const (
		NumWorkers, NumMessages = 10, 100
		TotalMessages           = NumWorkers * NumMessages
	)

	var wg sync.WaitGroup
	wg.Add(NumWorkers)
	for workerIdx := range NumWorkers {
		go func() {
			defer wg.Done()

			logger := logger.With(slog.Int("worker", workerIdx))

			for msgIdx := range NumMessages {
				logger.Info("Hello",
					slog.Int("message", msgIdx),
				)
			}
		}()
	}
	wg.Wait()

	output := buffer.String()
	assert.Equal(t, TotalMessages, strings.Count(output, "Hello"),
		"incorrect number of messages logged")

	for workerIdx := range NumWorkers {
		re := regexp.MustCompile(`worker=` + strconv.Itoa(workerIdx) + `\b`)
		assert.Equal(t, NumMessages, len(re.FindAllString(output, -1)),
			"incorrect number of messages logged for worker %d", workerIdx)
	}

	for msgIdx := range NumMessages {
		re := regexp.MustCompile(`message=` + strconv.Itoa(msgIdx) + `\b`)
		assert.Equal(t, NumWorkers, len(re.FindAllString(output, -1)),
			"incorrect number of messages logged for message %d", msgIdx)
	}
}
