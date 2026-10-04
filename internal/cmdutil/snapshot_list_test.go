package cmdutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/squarecloudofc/sdk-api-go/v3"
)

// runSnapshotList runs `snapshot list args...` against a listing with three
// snapshots of abc123 (one legacy) and one of def456, and returns stdout.
func runSnapshotList(t *testing.T, args ...string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "response": []map[string]any{
			{"name": "abc123_javascript_manual", "version_id": "v-1"},
			{"name": "def456_mongo_manual", "version_id": "v-2"},
			{"name": "abc123_javascript_automatic", "version_id": "v-3"},
			{"name": "abc123", "version_id": "v-4"},
		}})
	}))
	defer srv.Close()

	f := &fakeCLI{api: squarecloud.New("test-key", squarecloud.WithBaseURL(srv.URL))}
	cmd := NewSnapshotListCommand(f, squarecloud.SnapshotScopeApplications)
	cmd.Flags().Bool("json", false, "") // global on the real root
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	return f.out.String()
}

func TestSnapshotListJSONHonoursID(t *testing.T) {
	for id, want := range map[string][]string{
		"abc123": {"v-1", "v-3", "v-4"},
		"abc":    nil, // an exact ID, not a prefix
	} {
		out := runSnapshotList(t, id, "--json")
		var got []squarecloud.Snapshot
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("stdout is not JSON: %v\n%s", err, out)
		}
		var versions []string
		for _, s := range got {
			versions = append(versions, s.VersionID)
		}
		if !slices.Equal(versions, want) {
			t.Fatalf("list %s: versions = %v, want %v", id, versions, want)
		}
	}
}

func TestSnapshotListSummaryGroupsByResource(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(runSnapshotList(t)), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want a header and one row per resource:\n%s", len(lines), strings.Join(lines, "\n"))
	}
	counts := map[string]string{}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		counts[fields[0]] = fields[1]
	}
	if counts["abc123"] != "3" || counts["def456"] != "1" {
		t.Fatalf("counts = %v, want abc123:3 def456:1", counts)
	}
}

func TestVersionRowsUsesVersionID(t *testing.T) {
	snaps := []squarecloud.Snapshot{
		{Name: "app1_javascript_manual", VersionID: "v-123", Key: "Expires=1&versionId=wrong", Size: 2048, Modified: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		{Name: "app1_javascript_automatic", VersionID: "v-456"},
		{Name: "app1", VersionID: "v-789"},
		{Name: "other_mongo_manual", VersionID: "v-999"},
	}

	rows := versionRows(snaps, "app1")
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	// The full name and the version ID are what `snapshot restore` takes.
	if rows[0][0] != "app1_javascript_manual" || rows[0][1] != "v-123" {
		t.Fatalf("row = %q, want app1_javascript_manual / v-123", rows[0][:2])
	}
	if rows[0][2] != "2 KB" {
		t.Fatalf("size = %q, want 2 KB", rows[0][2])
	}
	if rows := versionRows(snaps, "app"); len(rows) != 0 {
		t.Fatalf("versionRows(app) = %v, want none: an ID is not a prefix", rows)
	}
}
