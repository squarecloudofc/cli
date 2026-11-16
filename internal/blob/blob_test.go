package blob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/squarecloudofc/sdk-api-go/v4"
)

// fakeBlob points BaseURL at handler and makes retries instant.
func fakeBlob(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	oldURL, oldSleep := BaseURL, sleep
	BaseURL, sleep = srv.URL+"/v1", func(context.Context, time.Duration) error { return nil }
	t.Cleanup(func() { BaseURL, sleep = oldURL, oldSleep })
	return &Client{Key: "test-key"}
}

func reply(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

func tempFile(t *testing.T, name string, size int64) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPartSize(t *testing.T) {
	const MiB = 1 << 20
	for _, c := range []struct{ size, maxSize, maxParts, want int64 }{
		{ChunkedThreshold + 1, 32 * MiB, 2048, 16 * MiB}, // the 16 MiB floor
		{17 * MiB * 2048, 32 * MiB, 2048, 17 * MiB},      // more to stay within max_parts
		{1 << 40, 32 * MiB, 2048, 32 * MiB},              // never above max_size
		{200 * MiB, 8 * MiB, 2048, 8 * MiB},              // even below the floor
		{91 * MiB, 0, 0, 91 * MiB},                       // no limits sent: one part
	} {
		if got := PartSize(c.size, c.maxSize, c.maxParts); got != c.want {
			t.Errorf("PartSize(%d, %d, %d) = %d, want %d", c.size, c.maxSize, c.maxParts, got, c.want)
		}
	}
}

func TestDefaultName(t *testing.T) {
	for in, want := range map[string]string{
		"photo.png":                       "photo",
		"/tmp/backup.tar.gz":              "backup", // the service's extension is tar.gz
		"archive.gz":                      "archive",
		"my.photo.png":                    "my.photo",
		"my file (1).zip":                 "my_file__1_",
		".env":                            "_env",
		"-x.txt":                          "_x",
		"relatório.pdf":                   "relat_rio",
		strings.Repeat("a", 200) + ".bin": strings.Repeat("a", 128),
		// The service refuses ".." and a trailing -ex<digits> (an old expiry
		// suffix).
		"a..b.txt":        "a_b",
		"a....b":          "a_.", // .b is the extension
		"...txt":          "_",
		"backup-ex30.zip": "backup_ex30",
		"x-ex1.tar.gz":    "x_ex1",
		"photo-ex.png":    "photo-ex",
		strings.Repeat("a", 124) + "-ex12345.bin": strings.Repeat("a", 124) + "_ex1", // cut to 128 first
	} {
		got := DefaultName(in)
		if got != want {
			t.Errorf("DefaultName(%q) = %q, want %q", in, got, want)
		}
		// validName in the Blob service.
		if !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`).MatchString(got) || strings.Contains(got, "..") || regexp.MustCompile(`-ex\d+$`).MatchString(got) {
			t.Errorf("DefaultName(%q) = %q, which the service refuses", in, got)
		}
	}
}

func TestUploadSimple(t *testing.T) {
	path := tempFile(t, "photo.png", 2048)
	c := fakeBlob(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/objects" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query(); got.Get("name") != "avatar" || got.Get("private") != "true" {
			t.Errorf("query = %v", got)
		}
		if r.Header.Get("Authorization") != "test-key" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(file)
		if header.Filename != "photo.png" || len(data) != 2048 {
			t.Errorf("file %q of %d bytes", header.Filename, len(data))
		}
		if r.ContentLength <= 2048 {
			t.Errorf("Content-Length = %d, want the whole form", r.ContentLength)
		}
		reply(w, 200, `{"status":"success","response":{"id":"u/avatar_h.png","private":true,"url":null,"size":2048}}`)
	})

	res, err := c.Upload(context.Background(), path, "avatar", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "u/avatar_h.png" || !res.Private || res.URL != nil {
		t.Fatalf("result = %+v", res)
	}
}

// Above 90 MiB: open, parts in parallel (a busy slot is retried), complete.
func TestUploadChunked(t *testing.T) {
	size := int64(ChunkedThreshold + 1)
	path := tempFile(t, "big.iso", size)

	var mu sync.Mutex
	received := map[string]int64{}
	refused := false
	c := fakeBlob(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/objects/chunked":
			if q := r.URL.Query(); q.Get("filename") != "big.iso" || q.Get("name") != "big" {
				t.Errorf("open query = %v", q)
			}
			reply(w, 200, `{"status":"success","response":{"upload":"handle","id":"u/big.iso","chunk":{"min_size":5242880,"max_size":33554432,"max_parts":2048}}}`)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/objects/chunked":
			part := r.URL.Query().Get("part")
			if r.URL.Query().Get("upload") != "handle" {
				t.Errorf("part without the handle")
			}
			if part == "2" && !refused {
				refused = true
				reply(w, 429, `{"status":"error","code":"TOO_MANY_CONCURRENT_CHUNKS"}`)
				return
			}
			n, _ := io.Copy(io.Discard, r.Body)
			received[part] += n
			reply(w, 200, `{"status":"success","response":{"part":`+part+`}}`)
		case r.Method == http.MethodPatch && r.URL.Path == "/v1/objects/chunked":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["upload"] != "handle" {
				t.Errorf("complete body = %v", body)
			}
			reply(w, 200, `{"status":"success","response":{"id":"u/big.iso","private":false,"url":"https://cdn.example/big.iso","size":94371841,"parts":6}}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			reply(w, 500, `{}`)
		}
	})

	var calls, lastDone, lastTotal int
	res, err := c.Upload(context.Background(), path, "", false, func(done, total int) { calls, lastDone, lastTotal = calls+1, done, total })
	if err != nil {
		t.Fatal(err)
	}
	if res.Parts != 6 || res.URL == nil {
		t.Fatalf("result = %+v", res)
	}
	var total int64
	for part, n := range received {
		if part != "6" && n != 16<<20 {
			t.Errorf("part %s carried %d bytes, want 16 MiB", part, n)
		}
		total += n
	}
	if len(received) != 6 || total != size || !refused {
		t.Fatalf("received %d parts, %d bytes (want 6, %d); busy part retried: %v", len(received), total, size, refused)
	}
	if calls != 6 || lastDone != 6 || lastTotal != 6 {
		t.Fatalf("progress: %d calls, last %d/%d", calls, lastDone, lastTotal)
	}
}

// A part that fails for good aborts the upload and never completes it.
func TestUploadChunkedAborts(t *testing.T) {
	path := tempFile(t, "big.bin", ChunkedThreshold+1)
	var mu sync.Mutex
	var aborted, completed bool
	c := fakeBlob(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodPost:
			reply(w, 200, `{"status":"success","response":{"upload":"handle","chunk":{"max_size":33554432,"max_parts":2048}}}`)
		case http.MethodPut:
			_, _ = io.Copy(io.Discard, r.Body)
			reply(w, 400, `{"status":"error","code":"INVALID_CHUNK_PART"}`)
		case http.MethodPatch:
			completed = true
		case http.MethodDelete:
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			aborted = body["upload"] == "handle"
			reply(w, 200, `{"status":"success"}`)
		}
	})

	_, err := c.Upload(context.Background(), path, "big", false, nil)
	var apiErr *squarecloud.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "INVALID_CHUNK_PART" {
		t.Fatalf("Upload = %v, want INVALID_CHUNK_PART", err)
	}
	if !aborted || completed {
		t.Fatalf("aborted %v, completed %v", aborted, completed)
	}
}

// Only what is safe to repeat is retried (SDK standard §4.1).
func TestRetryPolicy(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		code      string
		call      func(*Client) error
		wantCalls int
	}{
		{"GET 503 STORAGE_UNAVAILABLE", 503, "STORAGE_UNAVAILABLE", func(c *Client) error { _, err := c.Stats(context.Background()); return err }, 3},
		{"GET 429 RATE_LIMITED", 429, "RATE_LIMITED", func(c *Client) error { _, err := c.Stats(context.Background()); return err }, 1},
		{"GET 404", 404, "OBJECT_NOT_FOUND", func(c *Client) error { _, err := c.DownloadURL(context.Background(), "x"); return err }, 1},
		{"DELETE 503", 503, "STORAGE_UNAVAILABLE", func(c *Client) error { return c.Delete(context.Background(), "x") }, 1},
		{"POST 500", 500, "UPLOAD_FAILED", func(c *Client) error {
			_, err := c.Upload(context.Background(), tempFile(t, "a.txt", 600), "a", false, nil)
			return err
		}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			c := fakeBlob(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				_, _ = io.Copy(io.Discard, r.Body)
				reply(w, tc.status, `{"status":"error","code":"`+tc.code+`"}`)
			})
			err := tc.call(c)
			var apiErr *squarecloud.APIError
			if !errors.As(err, &apiErr) || apiErr.Code != tc.code || apiErr.Status != tc.status {
				t.Fatalf("error = %v, want %d %s", err, tc.status, tc.code)
			}
			if calls != tc.wantCalls {
				t.Fatalf("%d requests, want %d", calls, tc.wantCalls)
			}
		})
	}
}

func TestListFollowsCursor(t *testing.T) {
	c := fakeBlob(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("prefix") != "avatars/" {
			t.Errorf("prefix = %q", r.URL.Query().Get("prefix"))
		}
		if r.URL.Query().Get("cursor") == "" {
			reply(w, 200, `{"status":"success","response":{"objects":[{"id":"u/avatars/a.png","size":1,"private":false,"url":"https://cdn.example/a.png"}],"continuationToken":"next"}}`)
			return
		}
		reply(w, 200, `{"status":"success","response":{"objects":[{"id":"u/avatars/b.png","size":2,"private":true,"url":null}]}}`)
	})

	objects, err := c.List(context.Background(), "avatars/")
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 || objects[1].ID != "u/avatars/b.png" || objects[1].URL != nil {
		t.Fatalf("objects = %+v", objects)
	}
}

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("connection reset")
}

// A network error never prints the query: ?upload= is the chunked upload's
// credential.
func TestNetworkErrorHidesTheQuery(t *testing.T) {
	old := sleep
	sleep = func(context.Context, time.Duration) error { return nil }
	t.Cleanup(func() { sleep = old })

	c := &Client{Key: "test-key", HTTPClient: &http.Client{Transport: failingTransport{}}}
	err := c.call(context.Background(), http.MethodPut, "objects/chunked", url.Values{"upload": {"secret-handle"}, "part": {"1"}}, nil, nil)
	var apiErr *squarecloud.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != squarecloud.CodeNetworkError || !strings.Contains(err.Error(), "connection reset") {
		t.Fatalf("error = %v, want a network error with its cause", err)
	}
	if strings.Contains(err.Error(), "secret-handle") || strings.Contains(err.Error(), "?") {
		t.Fatalf("error shows the query: %v", err)
	}
}
