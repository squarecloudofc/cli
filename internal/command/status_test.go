package command

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/squarecloudofc/cli/i18n"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/v4"
)

func TestPrintStatus(t *testing.T) {
	en := i18n.NewLocalizer("en")
	start := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	cases := []struct {
		name    string
		status  squarecloud.ServiceStatus
		want    []string
		notWant []string
	}{
		{
			name:    "online",
			status:  squarecloud.ServiceStatus{Status: "online", Message: "All systems operational"},
			want:    []string{ui.CheckMark.String() + " All systems operational"},
			notWant: []string{ui.WarnMark.String(), ui.XMark.String(), en.T("commands.status.incidents")},
		},
		{
			name: "maintenance with an incident",
			status: squarecloud.ServiceStatus{
				Status:  "maintenance",
				Message: "Scheduled maintenance",
				Incidents: []squarecloud.ServiceIncident{
					{Type: "maintenance", Title: "Network upgrade in example-cluster", Status: "maintenance", StartedAt: start, EndsAt: &end},
					{Type: "incident", Title: "Slow deploys", Status: "degraded", StartedAt: start},
				},
				Components: []squarecloud.ServiceComponent{{Name: "Applications", Status: "maintenance"}},
			},
			want: []string{
				ui.WarnMark.String() + " Scheduled maintenance",
				en.T("commands.status.incidents"),
				"Network upgrade in example-cluster (maintenance)",
				en.T("commands.status.started", map[string]any{"Date": start.Local().Format(time.DateTime)}),
				en.T("commands.status.ends", map[string]any{"Date": end.Local().Format(time.DateTime)}),
				"Slow deploys (degraded)",
				en.T("commands.status.components"),
				"Applications (maintenance)",
			},
			notWant: []string{ui.XMark.String(), ui.CheckMark.String()},
		},
		{
			name:    "degraded",
			status:  squarecloud.ServiceStatus{Status: "degraded", Message: "Some services are degraded"},
			want:    []string{ui.WarnMark.String() + " Some services are degraded"},
			notWant: []string{ui.XMark.String()},
		},
		{
			// unknown means the check didn't run: never an X.
			name:    "unknown",
			status:  squarecloud.ServiceStatus{Status: "unknown", Message: "Status unavailable"},
			want:    []string{ui.UnknownMark.String() + " " + en.T("commands.status.unknown")},
			notWant: []string{ui.XMark.String(), ui.WarnMark.String(), ui.CheckMark.String()},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var b bytes.Buffer
			printStatus(&b, en, c.status)
			out := b.String()
			for _, want := range c.want {
				if !strings.Contains(out, want) {
					t.Errorf("missing %q in\n%s", want, out)
				}
			}
			for _, unwanted := range c.notWant {
				if strings.Contains(out, unwanted) {
					t.Errorf("unexpected %q in\n%s", unwanted, out)
				}
			}
		})
	}

	// The second incident has no announced end.
	var b bytes.Buffer
	printStatus(&b, en, cases[1].status)
	if n := strings.Count(b.String(), en.T("commands.status.ends", map[string]any{"Date": ""})); n != 1 {
		t.Errorf("expected end shown %d times, want once:\n%s", n, b.String())
	}
}
