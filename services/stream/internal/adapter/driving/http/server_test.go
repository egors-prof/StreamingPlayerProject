package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/egors-prof/StreamingPlayerProject/pkg/authjwt"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/domain"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/limiter"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/usecase/streaming"
)

type fakeTracks map[int]domain.Track

func (f fakeTracks) GetTrack(_ context.Context, id int) (domain.Track, error) {
	t, ok := f[id]
	if !ok {
		return domain.Track{}, errs.ErrNotFound
	}
	return t, nil
}

type testEnv struct {
	handler http.Handler
	jwt     *authjwt.Manager
	lim     *limiter.Limiter
	content []byte
}

func newTestEnv(t *testing.T, maxStreams int) *testEnv {
	t.Helper()
	dir := t.TempDir()
	content := []byte(strings.Repeat("0123456789", 1000)) // 10 000 байт
	if err := os.WriteFile(filepath.Join(dir, "song.mp3"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	m, _ := authjwt.NewManager("test-secret-that-is-long-enough-for-hs256")
	tracks := fakeTracks{1: {SongID: 1, FileName: "song.mp3", ContentType: "audio/mpeg"}}
	uc := streaming.New(tracks, m, dir, time.Hour, "/api")
	lim := limiter.New(maxStreams, 10*time.Millisecond)
	h, err := New(nil, uc, lim, m, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return &testEnv{handler: h, jwt: m, lim: lim, content: content}
}

func (e *testEnv) accessToken(t *testing.T) string {
	tok, _, err := e.jwt.Issue(authjwt.Claims{UserID: 1, Type: authjwt.TypeAccess}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// ticketURL проходит весь путь фронта: access-токен → билет → URL трека.
func (e *testEnv) ticketURL(t *testing.T, songID int) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/stream/tickets", strings.NewReader(`{"song_id":`+strconv.Itoa(songID)+`}`))
	req.Header.Set("Authorization", "Bearer "+e.accessToken(t))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("issue ticket: status %d, body %s", rec.Code, rec.Body)
	}
	var resp TicketResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return strings.TrimPrefix(resp.URL, "/api") // в тесте нет шлюза с префиксом /api
}

func TestStreamSupportsRangeRequests(t *testing.T) {
	env := newTestEnv(t, 5)
	url := env.ticketURL(t, 1)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Range", "bytes=100-199")
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", rec.Code)
	}
	if got := rec.Header().Get("Content-Range"); got != "bytes 100-199/10000" {
		t.Fatalf("Content-Range = %q", got)
	}
	if got := rec.Body.String(); got != string(env.content[100:200]) {
		t.Fatalf("wrong byte range returned")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "audio/mpeg" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestStreamRejectsBadTickets(t *testing.T) {
	env := newTestEnv(t, 5)
	access := env.accessToken(t)
	refresh, _, _ := env.jwt.Issue(authjwt.Claims{UserID: 1, Type: authjwt.TypeRefresh}, time.Minute)
	otherSong, _, _ := env.jwt.Issue(authjwt.Claims{UserID: 1, Type: authjwt.TypeStream, SongID: 2}, time.Minute)

	for name, ticket := range map[string]string{
		"no ticket":             "",
		"access token":          access,
		"refresh token":         refresh,
		"ticket for other song": otherSong,
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			env.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stream/songs/1?ticket="+ticket, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestTicketRequiresAccessToken(t *testing.T) {
	env := newTestEnv(t, 5)
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/stream/tickets", strings.NewReader(`{"song_id":1}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestTicketForUnknownSong(t *testing.T) {
	env := newTestEnv(t, 5)
	req := httptest.NewRequest(http.MethodPost, "/stream/tickets", strings.NewReader(`{"song_id":42}`))
	req.Header.Set("Authorization", "Bearer "+env.accessToken(t))
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestStreamReturns503WhenLimitReached(t *testing.T) {
	env := newTestEnv(t, 1)
	url := env.ticketURL(t, 1)

	// Занимаем единственный слот, как будто кто-то уже слушает.
	release, err := env.lim.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("status = %d, Retry-After = %q; want 503 with Retry-After", rec.Code, rec.Header().Get("Retry-After"))
	}

	release()
	rec = httptest.NewRecorder()
	env.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("after release: status = %d, want 200", rec.Code)
	}
	if env.lim.Stats().Active != 0 {
		t.Fatal("slot must be released after the response")
	}
}
