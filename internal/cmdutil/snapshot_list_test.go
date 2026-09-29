package cmdutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/squarecloudofc/sdk-api-go/v3"
)

func TestSnapshotListJSONHonoursID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "response": []map[string]any{
			{"name": "app1", "version_id": "v-1"},
			{"name": "other", "version_id": "v-2"},
			{"name": "app1", "version_id": "v-3"},
		}})
	}))
	defer srv.Close()

	f := &fakeCLI{api: squarecloud.New("test-key", squarecloud.WithBaseURL(srv.URL))}
	cmd := NewSnapshotListCommand(f, squarecloud.SnapshotScopeApplications)
	cmd.Flags().Bool("json", false, "") // global on the real root
	cmd.SetArgs([]string{"app1", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var got []squarecloud.Snapshot
	if err := json.Unmarshal(f.out.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, f.out.String())
	}
	if len(got) != 2 || got[0].VersionID != "v-1" || got[1].VersionID != "v-3" {
		t.Fatalf("snapshots = %+v, want only app1's two versions", got)
	}
}

func TestVersionRowsUsesVersionID(t *testing.T) {
	snaps := []squarecloud.Snapshot{
		{Name: "app1", VersionID: "v-123", Key: "Expires=1&versionId=wrong", Size: 2048, Modified: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		{Name: "other", VersionID: "v-999"},
	}

	rows := versionRows(snaps, "app1")
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0][1] != "v-123" {
		t.Fatalf("version id = %q, want v-123", rows[0][1])
	}
	if rows[0][2] != "2 KB" {
		t.Fatalf("size = %q, want 2 KB", rows[0][2])
	}
}
