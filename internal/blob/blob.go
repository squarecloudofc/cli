// Package blob is the CLI's client for Square Cloud Blob. It follows
// @squarecloud/blob: files above 90 MiB go up in parts sized from the
// server's limits, and only what is safe to repeat is retried.
package blob

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/squarecloudofc/sdk-api-go/v4"
)

// BaseURL is the Blob API root; tests point it at a fake.
var BaseURL = "https://blob.squarecloud.app/v1"

const (
	// ChunkedThreshold: a larger file goes up in parts (a single request
	// stops at 100 MB at the edge).
	ChunkedThreshold = 90 << 20
	minPartSize      = 16 << 20
	// A 7th part in flight is refused with TOO_MANY_CONCURRENT_CHUNKS.
	partConcurrency = 6
	maxRetries      = 2
	codeBusyChunks  = "TOO_MANY_CONCURRENT_CHUNKS"
)

// Client calls the Blob API with an API key. Only Key is required.
type Client struct {
	Key        string
	HTTPClient *http.Client
	UserAgent  string
}

// Object is one stored file, as listed.
type Object struct {
	ID        string  `json:"id"`
	Size      int64   `json:"size"`
	CreatedAt string  `json:"created_at"`
	ExpiresAt string  `json:"expires_at,omitempty"`
	Private   bool    `json:"private"`
	URL       *string `json:"url"` // nil for private objects
}

// UploadResult is a stored upload. Parts is set for chunked uploads only.
type UploadResult struct {
	ID        string  `json:"id"`
	Private   bool    `json:"private"`
	URL       *string `json:"url"`
	Size      int64   `json:"size"`
	ExpiresAt string  `json:"expires_at,omitempty"`
	Parts     int     `json:"parts,omitempty"`
}

// Link is a download link: the permanent URL of a public object, or a
// temporary one (ExpiresAt set) for a private object.
type Link struct {
	URL         string  `json:"url"`
	ExpiresAt   *string `json:"expires_at"`
	Private     bool    `json:"private"`
	Size        int64   `json:"size"`
	ContentType string  `json:"content_type"`
}

// Stats is the account's usage, in bytes. Billing is an estimate only.
type Stats struct {
	Usage struct {
		Objects int64 `json:"objects"`
		Storage int64 `json:"storage"`
	} `json:"usage"`
	Plan struct {
		Included int64 `json:"included"`
	} `json:"plan"`
	Billing struct {
		ExtraStorage  float64 `json:"extraStorage"`
		StoragePrice  float64 `json:"storagePrice"`
		ObjectsPrice  float64 `json:"objectsPrice"`
		TotalEstimate float64 `json:"totalEstimate"`
	} `json:"billing"`
	Month struct {
		Days           int     `json:"days"`
		AverageStorage float64 `json:"average_storage"`
	} `json:"month"`
}

// List returns every object whose ID starts with prefix ("" for all),
// following the cursor.
func (c *Client) List(ctx context.Context, prefix string) ([]Object, error) {
	var all []Object
	query := url.Values{}
	if prefix != "" {
		query.Set("prefix", prefix)
	}
	for {
		var page struct {
			Objects []Object `json:"objects"`
			Cursor  string   `json:"continuationToken"`
		}
		if err := c.call(ctx, http.MethodGet, "objects", query, nil, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Objects...)
		if page.Cursor == "" {
			return all, nil
		}
		query.Set("cursor", page.Cursor)
	}
}

// Delete removes an object for good (there is no trash).
func (c *Client) Delete(ctx context.Context, id string) error {
	return c.call(ctx, http.MethodDelete, "objects", nil, jsonBody(map[string]string{"object": id}), nil)
}

// DownloadURL returns a link to the object.
func (c *Client) DownloadURL(ctx context.Context, id string) (*Link, error) {
	var link Link
	query := url.Values{"object": {id}, "redirect": {"false"}}
	return &link, c.call(ctx, http.MethodGet, "objects/download", query, nil, &link)
}

// Stats returns the account's usage.
func (c *Client) Stats(ctx context.Context) (*Stats, error) {
	var stats Stats
	return &stats, c.call(ctx, http.MethodGet, "account/stats", nil, nil, &stats)
}

// compressed are the suffixes the service folds into a compound extension
// (backup.tar.gz has "tar.gz"), so both parts come off the name.
var compressed = map[string]bool{".gz": true, ".bz2": true, ".xz": true, ".zst": true, ".lz4": true, ".br": true, ".z": true}

// trailingEx is the old expiry suffix the service refuses at the end of a
// name.
var trailingEx = regexp.MustCompile(`-(ex[0-9]+)$`)

// DefaultName turns a file name into an object name: the name without its
// extension (the server adds the extension back), with what the API refuses
// (other characters, "..", a leading "." or "-", a trailing -ex<digits>)
// replaced by "_".
func DefaultName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if inner := filepath.Ext(name); compressed[strings.ToLower(ext)] && inner != "" && !compressed[strings.ToLower(inner)] {
		name = strings.TrimSuffix(name, inner)
	}
	if name == "" {
		name = base
	}
	name = strings.Map(func(r rune) rune {
		if r < 128 && (r == '_' || r == '-' || r == '.' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			return r
		}
		return '_'
	}, name)
	name = strings.ReplaceAll(name, "..", "_")
	if name[0] == '.' || name[0] == '-' {
		name = "_" + name[1:]
	}
	return trailingEx.ReplaceAllString(name[:min(len(name), 128)], "_$1")
}

// Upload stores the file at path as name (DefaultName when ""). Above
// ChunkedThreshold it goes up in parts, 6 at a time; progress (optional) is
// called after each part. A failed chunked upload is aborted.
func (c *Client) Upload(ctx context.Context, path, name string, private bool, progress func(done, total int)) (*UploadResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = DefaultName(path)
	}
	query := url.Values{"name": {name}}
	if private {
		query.Set("private", "true")
	}

	if info.Size() <= ChunkedThreshold {
		return c.uploadSimple(ctx, f, info.Size(), query)
	}
	return c.uploadChunked(ctx, f, info.Size(), query, progress)
}

// uploadSimple sends one multipart request with a known length: the server
// reserves quota by Content-Length.
func (c *Client) uploadSimple(ctx context.Context, f *os.File, size int64, query url.Values) (*UploadResult, error) {
	var envelope bytes.Buffer
	form := multipart.NewWriter(&envelope)
	if _, err := form.CreateFormFile("file", filepath.Base(f.Name())); err != nil {
		return nil, err
	}
	headLen := envelope.Len()
	if err := form.Close(); err != nil {
		return nil, err
	}
	head, tail := envelope.Bytes()[:headLen], envelope.Bytes()[headLen:]

	b := &body{
		size:        int64(len(head)) + size + int64(len(tail)),
		contentType: form.FormDataContentType(),
		open: func() io.Reader {
			return io.MultiReader(bytes.NewReader(head), io.NewSectionReader(f, 0, size), bytes.NewReader(tail))
		},
	}
	var res UploadResult
	return &res, c.call(ctx, http.MethodPost, "objects", query, b, &res)
}

func (c *Client) uploadChunked(ctx context.Context, f *os.File, size int64, query url.Values, progress func(done, total int)) (*UploadResult, error) {
	query.Set("filename", filepath.Base(f.Name()))
	var start struct {
		Upload string `json:"upload"` // a credential: never logged
		Chunk  struct {
			MaxSize  int64 `json:"max_size"`
			MaxParts int64 `json:"max_parts"`
		} `json:"chunk"`
	}
	if err := c.call(ctx, http.MethodPost, "objects/chunked", query, nil, &start); err != nil {
		return nil, err
	}
	handle := jsonBody(map[string]string{"upload": start.Upload})

	partSize := PartSize(size, start.Chunk.MaxSize, start.Chunk.MaxParts)
	parts := (size + partSize - 1) / partSize

	var (
		next, done atomic.Int64
		failed     atomic.Bool
		failure    error
		failOnce   sync.Once
		progressMu sync.Mutex
		wg         sync.WaitGroup
	)
	for range min(partConcurrency, parts) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// A failure stops new parts; the ones in flight finish first, so
			// none lands after the abort.
			for i := next.Add(1) - 1; i < parts && !failed.Load(); i = next.Add(1) - 1 {
				offset := i * partSize
				n := min(partSize, size-offset)
				b := &body{size: n, contentType: "application/octet-stream", open: func() io.Reader { return io.NewSectionReader(f, offset, n) }}
				q := url.Values{"upload": {start.Upload}, "part": {strconv.FormatInt(i+1, 10)}}
				if err := c.call(ctx, http.MethodPut, "objects/chunked", q, b, nil); err != nil {
					failOnce.Do(func() { failure = err; failed.Store(true) })
					return
				}
				if progress != nil {
					progressMu.Lock()
					progress(int(done.Add(1)), int(parts))
					progressMu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	var res UploadResult
	err := failure
	if err == nil {
		err = c.call(ctx, http.MethodPatch, "objects/chunked", nil, handle, &res)
	}
	if err != nil {
		// Frees one of the account's open-upload slots, Ctrl+C included.
		abortCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		_ = c.call(abortCtx, http.MethodDelete, "objects/chunked", nil, handle, nil)
		return nil, err
	}
	return &res, nil
}

// PartSize is the chunk size for a file of size bytes: at least 16 MiB,
// enough to stay within maxParts, never above the server's maxSize.
func PartSize(size, maxSize, maxParts int64) int64 {
	n := max(int64(minPartSize), (size+max(maxParts, 1)-1)/max(maxParts, 1))
	if maxSize > 0 {
		n = min(n, maxSize)
	}
	return n
}

// body is a request body that can be sent again: open returns a fresh
// reader per attempt.
type body struct {
	open        func() io.Reader
	size        int64
	contentType string
}

func jsonBody(v any) *body {
	data, _ := json.Marshal(v)
	return &body{size: int64(len(data)), contentType: "application/json", open: func() io.Reader { return bytes.NewReader(data) }}
}

// sleep waits between attempts; tests replace it.
var sleep = func(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// call sends one request and decodes the response of the envelope into out
// (nil to skip). Retried, as in the SDK standard (§4.1): network errors and
// 5xx on GET and on part PUTs (idempotent per part number), and 429
// TOO_MANY_CONCURRENT_CHUNKS on part PUTs. Never any other 429: RATE_LIMITED
// can be a 30-minute block. No Retry-After exists.
func (c *Client) call(ctx context.Context, method, path string, query url.Values, b *body, out any) error {
	idempotent := method == http.MethodGet || (method == http.MethodPut && path == "objects/chunked")
	for attempt := 0; ; attempt++ {
		err := c.do(ctx, method, path, query, b, out)
		var apiErr *squarecloud.APIError
		if err == nil || !idempotent || attempt >= maxRetries || ctx.Err() != nil || !errors.As(err, &apiErr) {
			return err
		}
		if apiErr.Status != 0 && apiErr.Status < 500 && apiErr.Code != codeBusyChunks {
			return err
		}
		backoff := min(8*time.Second, 500*time.Millisecond<<attempt)
		if err := sleep(ctx, time.Duration(float64(backoff)*(0.5+rand.Float64()/2))); err != nil {
			return err
		}
	}
}

// do sends one attempt. Every failure is a *squarecloud.APIError, Status 0
// for network errors, so the CLI's hints apply.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, b *body, out any) error {
	endpoint := strings.TrimRight(BaseURL, "/") + "/" + path
	apiPath := endpoint
	if u, err := url.Parse(endpoint); err == nil {
		apiPath = u.Path
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var reader io.Reader
	if b != nil {
		reader = b.open()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.Key)
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if b != nil {
		req.ContentLength = b.size
		req.Header.Set("Content-Type", b.contentType)
	}

	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	netErr := func(err error) error {
		// Keep only the cause: *url.Error prints the URL, whose ?upload= is a
		// credential.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return &squarecloud.APIError{Code: squarecloud.CodeNetworkError, Message: err.Error(), Method: method, Path: apiPath}
	}
	res, err := hc.Do(req)
	if err != nil {
		return netErr(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 16<<20))
	if err != nil {
		return netErr(err) // cut off mid-read: a network error, not a bad body
	}

	var envelope struct {
		Status   string          `json:"status"`
		Code     string          `json:"code"`
		Message  string          `json:"message"`
		Response json.RawMessage `json:"response"`
	}
	decodeErr := json.Unmarshal(data, &envelope)
	if res.StatusCode >= 300 || decodeErr != nil || envelope.Status != "success" {
		code := envelope.Code
		if code == "" {
			code = squarecloud.CodeUnknown
		}
		return &squarecloud.APIError{Status: res.StatusCode, Code: code, Message: envelope.Message, Method: method, Path: apiPath}
	}
	if out == nil || len(envelope.Response) == 0 {
		return nil
	}
	if err := json.Unmarshal(envelope.Response, out); err != nil {
		return &squarecloud.APIError{Status: res.StatusCode, Code: squarecloud.CodeUnknown, Message: "invalid response: " + err.Error(), Method: method, Path: apiPath}
	}
	return nil
}
