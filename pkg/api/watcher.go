package api

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	// fetchInterval is the interval to use for fetching
	// new run states.
	fetchInterval = 1 * time.Second
)

// LogsClient represents a logs client.
type logsClient interface {
	GetLogs(ctx context.Context, runID string, t time.Time) (GetLogsResponse, error)
	GetOutputs(ctx context.Context, runID string) (GetOutputsResponse, error)
	GetRun(ctx context.Context, runID string) (GetRunResponse, error)
}

// RunState represents a run state.
type RunState struct {
	Status  RunStatus
	Logs    []LogItem
	Outputs Outputs
	err     error
}

// Err returns an error if any.
func (r RunState) Err() error {
	return r.err
}

// Stopped returns true if the run has stopped.
func (r RunState) Stopped() bool {
	switch r.Status {
	case RunCancelled, RunFailed, RunSucceeded:
		return true
	default:
		return false
	}
}

// Failed returns true if the task has failed.
func (r RunState) Failed() bool {
	return r.Status == RunFailed
}

// Merge merges run states.
func (r RunState) merge(b RunState) RunState {
	return RunState{
		Status:  b.Status,
		Logs:    DedupeLogs(r.Logs, b.Logs),
		Outputs: b.Outputs,
	}
}

// LastTimestamp returns the last log timestamp.
func (r RunState) lastTimestamp() time.Time {
	if l := len(r.Logs); l > 0 {
		return r.Logs[l-1].Timestamp
	}
	return time.Time{}
}

// Watcher represents a run watcher.
type Watcher struct {
	ctx    context.Context
	client logsClient
	runID  string
	state  chan RunState
}

// NewWatcher returns a new watcher with the given runID and context.
func newWatcher(ctx context.Context, client logsClient, runID string) *Watcher {
	w := &Watcher{
		ctx:    ctx,
		client: client,
		runID:  runID,
		state:  make(chan RunState),
	}
	go w.watch()
	return w
}

// RunID returns the runID.
func (w *Watcher) RunID() string {
	return w.runID
}

// Next returns the next run state.
func (w *Watcher) Next() RunState {
	return <-w.state
}

// Watch implements a watcher go-routine.
//
// On every tick the method attempts to fetch the most recent
// logs and run status and sends them on an internal "state" channel
// on fetch failure, or when the task is canceled a special state
// is sent with an error.
func (w *Watcher) watch() {
	var ticker = time.NewTicker(fetchInterval)
	var prev RunState

	for {
		select {
		case <-w.ctx.Done():
			// TODO(amir): actually send a cancel request
			// and wait for the API state change.
			w.send(w.ctx, RunState{})

		case <-ticker.C:
			state, err := w.fetch(w.ctx, prev)
			if err != nil {
				w.send(w.ctx, RunState{
					err: err,
				})
				return
			}

			w.send(w.ctx, prev.merge(state))
			prev = state
		}
	}
}

// Send sends the given state with context.
func (w *Watcher) send(ctx context.Context, state RunState) {
	select {
	case w.state <- state:
	case <-ctx.Done():
		w.state <- RunState{
			err: ctx.Err(),
		}
	}
}

// Fetch fetches the next state.
func (w *Watcher) fetch(ctx context.Context, prev RunState) (RunState, error) {
	var eg, subctx = errgroup.WithContext(ctx)
	var since = prev.lastTimestamp().UTC()
	var state = new(RunState)

	eg.Go(func() error {
		run, err := w.client.GetRun(subctx, w.runID)
		if err != nil {
			return errors.Wrap(err, "get run")
		}

		state.Status = run.Run.Status

		if state.Stopped() {
			resp, err := w.client.GetOutputs(subctx, w.runID)
			if err != nil {
				return errors.Wrap(err, "get outputs")
			}

			state.Outputs = resp.Outputs
		}
		return nil
	})

	eg.Go(func() error {
		resp, err := w.client.GetLogs(subctx, w.runID, since)
		if err != nil {
			return errors.Wrap(err, "get logs")
		}
		sortLogs(resp.Logs)

		state.Logs = resp.Logs
		return nil
	})

	if err := eg.Wait(); err != nil {
		return RunState{}, err
	}

	return *state, nil
}

// SortLogs returns sorted logs.
func sortLogs(logs []LogItem) {
	sort.Slice(logs, func(i, j int) bool {
		a, b := logs[i], logs[j]
		return a.Timestamp.Before(b.Timestamp) && a.InsertID < b.InsertID
	})
}

// LogKey represents a LogItem key.
type logKey struct {
	ts       time.Time
	insertID string
}

// NewLogKey returns a LogKey for LogItem.
func newLogKey(l LogItem) logKey {
	return logKey{
		ts:       l.Timestamp,
		insertID: l.InsertID,
	}
}

// DedupeLogs de-duplicates logs by timestamp and insertId.
func DedupeLogs(a, b []LogItem) []LogItem {
	var dedupe = make(map[logKey]struct{})
	var ret []LogItem

	for _, l := range a {
		dedupe[newLogKey(l)] = struct{}{}
	}

	for _, l := range b {
		if _, ok := dedupe[newLogKey(l)]; !ok {
			ret = append(ret, l)
		}
	}

	return ret
}
