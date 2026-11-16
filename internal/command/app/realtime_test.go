package app

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/config"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

type fakeCLI struct {
	api      *squarecloud.Client
	out, err bytes.Buffer
}

func (f *fakeCLI) Config() *config.Config   { return &config.Config{} }
func (f *fakeCLI) API() *squarecloud.Client { return f.api }
func (f *fakeCLI) I18n() i18n.Localizer     { return i18n.NewLocalizer("en") }
func (f *fakeCLI) Err() io.Writer           { return &f.err }
func (f *fakeCLI) In() io.ReadCloser        { return io.NopCloser(strings.NewReader("")) }
func (f *fakeCLI) Out() io.Writer           { return &f.out }

// TestFollowRealtimeReopens: a stream the API ends cleanly (its 10-minute
// limit) is opened again, until Ctrl+C or an error that isn't an end.
func TestFollowRealtimeReopens(t *testing.T) {
	old := realtimeWait
	realtimeWait = func(ctx context.Context, _ time.Duration) bool { return ctx.Err() == nil }
	t.Cleanup(func() { realtimeWait = old })

	for _, c := range []struct {
		name    string
		opens   int // the open that answers 404; 0 = never
		wantErr string
	}{
		{name: "until Ctrl+C"},
		{name: "until the app is gone", opens: 3, wantErr: "APP_NOT_FOUND"},
	} {
		t.Run(c.name, func(t *testing.T) {
			opens := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				opens++
				if opens == c.opens {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"status":"error","code":"APP_NOT_FOUND"}`))
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "event: logs\ndata: \x01line %d\n\nevent: logs\ndata: \x02oops %d\n\n", opens, opens)
			}))
			defer srv.Close()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f := &fakeCLI{api: squarecloud.New("key", squarecloud.WithBaseURL(srv.URL))}
			stream, err := f.api.Apps.Realtime(ctx, "a1")
			if err != nil {
				t.Fatal(err)
			}

			err = followRealtime(ctx, f, "a1", stream, func(event squarecloud.RealtimeEvent) {
				printLogLine(f, event)
				if event.Line == "oops 3" {
					cancel() // Ctrl+C
				}
			})
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("followRealtime = %v, want nil after Ctrl+C", err)
			case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
				t.Fatalf("followRealtime = %v, want %s", err, c.wantErr)
			}

			wantOut, wantErr := "line 1\nline 2\n", "oops 1\n"
			if c.wantErr == "" {
				wantOut += "line 3\n"
			}
			if got := f.out.String(); got != wantOut {
				t.Errorf("stdout = %q, want %q", got, wantOut)
			}
			if !strings.HasPrefix(f.err.String(), wantErr) || strings.Count(f.err.String(), f.I18n().T("commands.app.realtime.reconnecting")) != 2 {
				t.Errorf("stderr = %q, want app stderr lines and 2 reconnect notes", f.err.String())
			}
		})
	}
}

// fakeStream replays events, then ends with err (io.EOF when nil).
type fakeStream struct {
	events []squarecloud.RealtimeEvent
	err    error
}

func (s *fakeStream) Next() (squarecloud.RealtimeEvent, error) {
	if len(s.events) == 0 {
		return squarecloud.RealtimeEvent{}, cmp.Or(s.err, io.EOF)
	}
	event := s.events[0]
	s.events = s.events[1:]
	return event, nil
}

func (s *fakeStream) Close() error { return nil }

// TestFollowRealtimePacing: after each end the CLI waits the full pace (the
// SDK may have reopened in between, so its own open time means nothing), and
// gives up after 3 reopens in a row that bring no logs or status event.
func TestFollowRealtimePacing(t *testing.T) {
	logs := squarecloud.RealtimeEvent{Event: "logs", Line: "hi"}
	status := squarecloud.RealtimeEvent{Event: "status", Status: &squarecloud.RealtimeStatus{Status: "running"}}
	system := squarecloud.RealtimeEvent{Event: "system", Data: "REALTIME_CONNECTING | 1"}
	stream := func(events ...squarecloud.RealtimeEvent) *fakeStream { return &fakeStream{events: events} }
	en := i18n.NewLocalizer("en")
	noEvents := en.T("commands.app.realtime.no_events", map[string]any{"Count": realtimeIdleReopens})
	if !strings.Contains(noEvents, "3 times") {
		t.Fatalf("no_events = %q", noEvents)
	}

	for _, c := range []struct {
		name     string
		first    *fakeStream
		reopens  []*fakeStream
		openErr  error // the open after the reopens run out
		cancelAt int   // the wait that plays Ctrl+C; 0 = never
		wantErr  string
	}{
		{name: "three empty reopens end it", first: stream(logs), reopens: []*fakeStream{stream(system), stream(system), stream(system)}, wantErr: noEvents},
		{name: "logs and status reset the count", first: stream(system),
			reopens: []*fakeStream{stream(system), stream(system), stream(system, logs), stream(), stream(), stream(status), stream(), stream(), stream()}, wantErr: noEvents},
		{name: "Ctrl+C while waiting", first: stream(logs), cancelAt: 1},
		{name: "a failed reopen ends it", first: stream(logs), reopens: []*fakeStream{stream(logs)}, openErr: errors.New("APP_NOT_FOUND"), wantErr: "APP_NOT_FOUND"},
		{name: "an error that isn't an end", first: &fakeStream{err: errors.New("REALTIME_ERROR")}, wantErr: "REALTIME_ERROR"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var waits []time.Duration
			oldWait, oldOpen := realtimeWait, openRealtime
			t.Cleanup(func() { realtimeWait, openRealtime = oldWait, oldOpen })
			realtimeWait = func(ctx context.Context, d time.Duration) bool {
				waits = append(waits, d)
				if len(waits) == c.cancelAt {
					cancel()
				}
				return ctx.Err() == nil
			}
			opens := 0
			openRealtime = func(context.Context, cli.SquareCLI, string) (realtimeStream, error) {
				if opens == len(c.reopens) {
					if c.openErr == nil {
						t.Fatalf("open %d: followRealtime should have stopped", opens+1)
					}
					return nil, c.openErr
				}
				opens++
				return c.reopens[opens-1], nil
			}

			f := &fakeCLI{}
			err := followRealtime(ctx, f, "a1", c.first, func(squarecloud.RealtimeEvent) {})
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("followRealtime = %v, want nil", err)
			case c.wantErr != "" && (err == nil || err.Error() != c.wantErr):
				t.Fatalf("followRealtime = %v, want %q", err, c.wantErr)
			}
			if opens != len(c.reopens) {
				t.Errorf("%d reopens, want %d", opens, len(c.reopens))
			}
			for _, w := range waits {
				if w != realtimePace {
					t.Errorf("waited %v before a reopen, want the full %v", w, realtimePace)
				}
			}
		})
	}
}

func TestPrintMetrics(t *testing.T) {
	en := i18n.NewLocalizer("en")
	storage, peak := int64(5<<20), 91.5
	day := []squarecloud.MetricPoint{{CPU: 1.5, RAM: 120, Net: []float64{2048, 1024}, Disk: []int64{4096, 0}, Storage: &storage}}
	week := []squarecloud.MetricPoint{{CPU: 1.5, RAM: 120, CPUMax: &peak, RAMMax: &peak}}

	for _, c := range []struct {
		name     string
		points   []squarecloud.MetricPoint
		rng      string
		want     []string
		notWant  []string
		wantNote bool
	}{
		{name: "24h", points: day, rng: "24h", want: []string{"DISK I/O", "STORAGE", "2 KB/1 KB", "4 KB/0 Bytes", "5.00 MB"}, notWant: []string{"CPU MAX"}},
		{name: "7d", points: week, rng: "7d", want: []string{"CPU MAX %", "RAM MAX MB", "91.50", "91.5"}},
		// A plan without 7d gets the 24h answer, and is told so.
		{name: "7d answered with 24h", points: day, rng: "7d", notWant: []string{"CPU MAX"}, wantNote: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeCLI{}
			cmd := NewMetricsCommand(f)
			if err := PrintMetrics(f, cmd, c.points, c.rng); err != nil {
				t.Fatal(err)
			}
			for _, want := range c.want {
				if !strings.Contains(f.out.String(), want) {
					t.Errorf("missing %q in\n%s", want, f.out.String())
				}
			}
			for _, unwanted := range c.notWant {
				if strings.Contains(f.out.String(), unwanted) {
					t.Errorf("unexpected %q in\n%s", unwanted, f.out.String())
				}
			}
			if got := strings.Contains(f.err.String(), en.T("commands.app.metrics.only_24h")); got != c.wantNote {
				t.Errorf("24h note = %v, want %v (stderr %q)", got, c.wantNote, f.err.String())
			}
		})
	}

	if err := CheckMetricsRange(&fakeCLI{}, "30d"); err == nil || !strings.Contains(err.Error(), "30d") {
		t.Errorf("CheckMetricsRange(30d) = %v", err)
	}
	if err := errors.Join(CheckMetricsRange(&fakeCLI{}, "24h"), CheckMetricsRange(&fakeCLI{}, "7d")); err != nil {
		t.Errorf("valid ranges rejected: %v", err)
	}
}
