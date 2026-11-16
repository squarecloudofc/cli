package cmdutil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

// fakeSleep stands in for the poll wait, recording each one; cancel, when
// set, runs on the n-th wait to play a Ctrl+C.
func fakeSleep(t *testing.T, waits *[]time.Duration, cancelAt int, cancel context.CancelFunc) {
	t.Helper()
	real := sleep
	sleep = func(ctx context.Context, d time.Duration) error {
		*waits = append(*waits, d)
		if cancel != nil && len(*waits) == cancelAt {
			cancel()
		}
		return ctx.Err()
	}
	t.Cleanup(func() { sleep = real })
}

func TestFollowRestore(t *testing.T) {
	en := i18n.NewLocalizer("en")
	st := func(state, step, code string) squarecloud.SnapshotRestoreState {
		return squarecloud.SnapshotRestoreState{State: state, Step: step, Code: code}
	}

	cases := []struct {
		name     string
		states   []squarecloud.SnapshotRestoreState
		stateErr error
		cancelAt int // the wait that plays Ctrl+C; 0 = never
		wantErr  string
		wantOut  []string
		notOut   []string
	}{
		{
			name:    "runs, then succeeds",
			states:  []squarecloud.SnapshotRestoreState{st("queued", "", ""), st("running", "download", ""), st("running", "download", ""), st("running", "extract", ""), st("success", "done", "")},
			wantOut: []string{"Step: download", "Step: extract", en.T("commands.snapshot.restore.success")},
		},
		{
			name:    "fails with the API's code",
			states:  []squarecloud.SnapshotRestoreState{st("running", "extract", ""), st("error", "done", "SNAPSHOT_CORRUPT")},
			wantErr: en.T("commands.snapshot.restore.failed", map[string]any{"Code": "SNAPSHOT_CORRUPT"}),
			notOut:  []string{en.T("commands.snapshot.restore.success")},
		},
		{
			name:    "fails without a code",
			states:  []squarecloud.SnapshotRestoreState{st("error", "", "")},
			wantErr: en.T("commands.snapshot.restore.failed", map[string]any{"Code": squarecloud.CodeSnapshotRestoreFailed}),
		},
		{
			name:    "the restore is no longer reported",
			states:  []squarecloud.SnapshotRestoreState{st("running", "x", ""), st("none", "", "")},
			wantErr: en.T("commands.snapshot.restore.lost"),
		},
		{
			// Only queued and running go on: polling a state this version
			// doesn't know might never end.
			name:    "an unknown state ends it, result unknown",
			states:  []squarecloud.SnapshotRestoreState{st("queued", "", ""), st("paused", "x", "")},
			wantErr: en.T("commands.snapshot.restore.lost"),
		},
		{
			name:    "an empty state ends it too",
			states:  []squarecloud.SnapshotRestoreState{st("", "", "")},
			wantErr: en.T("commands.snapshot.restore.lost"),
		},
		{
			name:     "Ctrl+C stops watching, not the restore",
			states:   []squarecloud.SnapshotRestoreState{st("running", "download", ""), st("running", "download", "")},
			cancelAt: 2,
			wantOut:  []string{"Step: download", en.T("commands.snapshot.restore.background")},
			notOut:   []string{en.T("commands.snapshot.restore.success")},
		},
		{
			name:     "a failed state read says the restore goes on",
			stateErr: &squarecloud.APIError{Status: 503, Code: "UNKNOWN_ERROR"},
			wantErr:  "UNKNOWN_ERROR",
			wantOut:  []string{en.T("commands.snapshot.restore.background")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var waits []time.Duration
			fakeSleep(t, &waits, c.cancelAt, cancel)

			reads := 0
			state := func(context.Context) (squarecloud.SnapshotRestoreState, error) {
				if c.stateErr != nil {
					return squarecloud.SnapshotRestoreState{}, c.stateErr
				}
				if reads >= len(c.states) {
					t.Fatalf("state read %d after the restore ended", reads+1)
				}
				reads++
				return c.states[reads-1], nil
			}

			f := &fakeCLI{}
			err := followRestore(ctx, f, state)
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("followRestore = %v, want success", err)
			case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
				t.Fatalf("followRestore = %v, want %q", err, c.wantErr)
			}

			out := f.out.String()
			for _, want := range c.wantOut {
				if strings.Count(out, want) != 1 {
					t.Errorf("output has %q %d times, want once:\n%s", want, strings.Count(out, want), out)
				}
			}
			for _, unwanted := range c.notOut {
				if strings.Contains(out, unwanted) {
					t.Errorf("output has %q:\n%s", unwanted, out)
				}
			}
			for _, w := range waits {
				if w != restorePoll {
					t.Fatalf("waited %v between reads, want %v (the API allows 30 a minute)", w, restorePoll)
				}
			}
		})
	}
}

// TestRestoreCommand drives `snapshot restore` against a fake API: a 202
// is followed through the state route; a 200 is a finished restore.
func TestRestoreCommand(t *testing.T) {
	en := i18n.NewLocalizer("en")
	var waits []time.Duration
	fakeSleep(t, &waits, 0, nil)

	cases := []struct {
		name       string
		scope      squarecloud.SnapshotScope
		path       string
		status     int // of the POST
		args       []string
		wantReads  int
		wantOutput string
	}{
		{"app, finished at once", squarecloud.SnapshotScopeApplications, "/apps/a1/snapshots/restore", 200, nil, 0, en.T("commands.snapshot.restore.success")},
		{"app, still running", squarecloud.SnapshotScopeApplications, "/apps/a1/snapshots/restore", 202, nil, 2, en.T("commands.snapshot.restore.success")},
		{"database, still running", squarecloud.SnapshotScopeDatabases, "/databases/a1/snapshots/restore", 202, nil, 2, en.T("commands.snapshot.restore.success")},
		{"--no-wait", squarecloud.SnapshotScopeApplications, "/apps/a1/snapshots/restore", 202, []string{"--no-wait"}, 0, en.T("commands.snapshot.restore.background")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reads := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != c.path {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost {
					w.WriteHeader(c.status)
					if c.status == http.StatusAccepted {
						_, _ = w.Write([]byte(`{"status":"error","code":"SNAPSHOT_RESTORE_PROCESSING"}`))
					} else {
						_, _ = w.Write([]byte(`{"status":"success"}`))
					}
					return
				}
				reads++
				state := map[string]any{"state": "running", "step": "extract"}
				if reads == 2 {
					state = map[string]any{"state": "success", "step": "done"}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "response": state})
			}))
			defer srv.Close()

			f := &fakeCLI{api: squarecloud.New("key", squarecloud.WithBaseURL(srv.URL))}
			cmd := NewSnapshotRestoreCommand(f, c.scope)
			cmd.SetArgs(append([]string{"a1", "a1_runtime", "version-id", "--yes"}, c.args...))
			if err := cmd.ExecuteContext(context.Background()); err != nil {
				t.Fatalf("restore = %v", err)
			}
			if reads != c.wantReads {
				t.Errorf("state reads = %d, want %d", reads, c.wantReads)
			}
			if !strings.Contains(f.out.String(), c.wantOutput) {
				t.Errorf("output %q, want %q", f.out.String(), c.wantOutput)
			}
		})
	}

	// Refused restores keep the API's error for the hint.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"status":"error","code":"RESTORE_IN_PROGRESS"}`))
	}))
	defer srv.Close()
	f := &fakeCLI{api: squarecloud.New("key", squarecloud.WithBaseURL(srv.URL))}
	cmd := NewSnapshotRestoreCommand(f, squarecloud.SnapshotScopeApplications)
	cmd.SetArgs([]string{"a1", "a1_runtime", "version-id", "--yes"})
	var apiErr *squarecloud.APIError
	if err := cmd.ExecuteContext(context.Background()); !errors.As(err, &apiErr) || apiErr.Code != squarecloud.CodeRestoreInProgress {
		t.Fatalf("restore = %v, want RESTORE_IN_PROGRESS", err)
	}
}
