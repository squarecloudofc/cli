package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/cmdutil"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

const (
	// realtimePace is the API's limit of one realtime open per 5 s per app,
	// plus a margin for clock skew.
	realtimePace = 5500 * time.Millisecond
	// realtimeIdleReopens is how many reopens in a row may bring no logs or
	// status event before followRealtime gives up, as in the SDK.
	realtimeIdleReopens = 3
)

// realtimeStream is a *squarecloud.Realtime; tests use a fake.
type realtimeStream interface {
	Next() (squarecloud.RealtimeEvent, error)
	Close() error
}

// openRealtime opens the stream of appID; tests replace it.
var openRealtime = func(ctx context.Context, squareCli cli.SquareCLI, appID string) (realtimeStream, error) {
	stream, err := squareCli.API().Apps.Realtime(ctx, appID)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// realtimeWait waits d, false when ctx ends first; tests replace it.
var realtimeWait = func(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func NewRealtimeCommand(squareCli cli.SquareCLI) *cobra.Command {
	return &cobra.Command{
		Use:   "realtime [app id]",
		Short: squareCli.I18n().T("metadata.commands.app.realtime.short"),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := cmdutil.ResolveAppID(cmd.Context(), squareCli, args)
			if err != nil {
				return err
			}

			stream, err := squareCli.API().Apps.Realtime(cmd.Context(), appID)
			if err != nil {
				return err
			}

			fmt.Fprintln(squareCli.Out(), squareCli.I18n().T("commands.app.realtime.connected", map[string]any{"Appid": appID}))

			var lastStatus string
			return followRealtime(cmd.Context(), squareCli, appID, stream, func(event squarecloud.RealtimeEvent) {
				switch event.Event {
				case "logs":
					printLogLine(squareCli, event)
				case "status":
					// Status frames arrive every few seconds; show only changes.
					if event.Status.Status != lastStatus {
						lastStatus = event.Status.Status
						fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.TextYellow.SetString("[status]"), lastStatus)
					}
				default:
					fmt.Fprintf(squareCli.Out(), "%s %s\n", ui.TextYellow.SetString("["+event.Event+"]"), event.Data)
				}
			})
		},
	}
}

// printLogLine writes a realtime log line to stdout, or stderr for the
// app's stderr.
func printLogLine(squareCli cli.SquareCLI, event squarecloud.RealtimeEvent) {
	out := squareCli.Out()
	if event.Stream == "stderr" {
		out = squareCli.Err()
	}
	fmt.Fprintln(out, event.Line)
}

// followRealtime hands every event of stream to onEvent until Ctrl+C. When
// the API ends the stream (its 10-minute limit, or REALTIME_DISCONNECTED) it
// opens a new one; the SDK already reopens after network drops and
// REALTIME_RECONNECT. It gives up after realtimeIdleReopens reopens in a row
// without a logs or status event, and on any other error.
func followRealtime(ctx context.Context, squareCli cli.SquareCLI, appID string, stream realtimeStream, onEvent func(squarecloud.RealtimeEvent)) error {
	idle := 0 // reopens since the last logs or status event
	for {
		var err error
		for err == nil {
			var event squarecloud.RealtimeEvent
			if event, err = stream.Next(); err == nil {
				if event.Event == "logs" || event.Event == "status" {
					idle = 0
				}
				onEvent(event)
			}
		}
		stream.Close()
		if ctx.Err() != nil {
			return nil
		}
		if !errors.Is(err, io.EOF) {
			return err
		}
		if idle >= realtimeIdleReopens {
			return errors.New(squareCli.I18n().T("commands.app.realtime.no_events", map[string]any{"Count": realtimeIdleReopens}))
		}

		// The full pace from the end, not from the CLI's own open: the SDK
		// may have reopened the stream since.
		if !realtimeWait(ctx, realtimePace) {
			return nil
		}
		fmt.Fprintln(squareCli.Err(), ui.TextYellow.Render(squareCli.I18n().T("commands.app.realtime.reconnecting")))

		idle++
		if stream, err = openRealtime(ctx, squareCli, appID); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}
