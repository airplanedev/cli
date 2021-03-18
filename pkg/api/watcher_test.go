package api

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func init() {
	fetchInterval = time.Nanosecond
}

func TestWatcher(t *testing.T) {
	t.Run("dedupes logs (a, a, ab)", func(t *testing.T) {
		var ctx = context.Background()
		var assert = require.New(t)
		var lcm = logsClientMock{}

		var a = LogItem{
			Timestamp: time.Now().Add(time.Second),
			InsertID:  "001",
			Text:      "A",
		}
		var b = LogItem{
			Timestamp: time.Now().Add(2 * time.Second),
			InsertID:  "002",
			Text:      "B",
		}

		var responses = []GetLogsResponse{
			{Logs: []LogItem{a}},
			{Logs: []LogItem{a}},
			{Logs: []LogItem{a, b}},
		}

		var reqs int64
		lcm.getLogs = func(string, time.Time) (GetLogsResponse, error) {
			var n int

			if n = int(atomic.AddInt64(&reqs, 1)); n > 3 {
				n = len(responses)
			}

			fmt.Println("get response", responses[n-1])
			return responses[n-1], nil
		}

		lcm.getRun = func(string) (GetRunResponse, error) {
			var run = Run{Status: RunActive}

			if atomic.LoadInt64(&reqs) > 2 {
				run.Status = RunSucceeded
			}

			return GetRunResponse{run}, nil
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var w = newWatcher(ctx, lcm, "run_id")
		var state RunState
		var printed []string

		for {
			if state = w.Next(); state.Err() != nil {
				break
			}

			for _, l := range state.Logs {
				printed = append(printed, l.Text)
			}

			if state.Stopped() {
				break
			}
		}

		assert.NoError(state.Err())
		assert.Equal([]string{"A", "B"}, printed)
	})
}

type logsClientMock struct {
	getLogs func(runID string, s time.Time) (GetLogsResponse, error)
	getRun  func(runID string) (GetRunResponse, error)
}

func (lcm logsClientMock) GetLogs(ctx context.Context, runID string, since time.Time) (GetLogsResponse, error) {
	return lcm.getLogs(runID, since)
}

func (lcm logsClientMock) GetRun(ctx context.Context, runID string) (GetRunResponse, error) {
	return lcm.getRun(runID)
}
