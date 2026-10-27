package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/egors-prof/likes_service/internal/usecase/likes"
	"github.com/egors-prof/streamingplayer/pkg/authjwt"
)

// fakeStorage повторяет поведение БД: UNIQUE (user_id, song_id).
type fakeStorage struct{ liked map[int][]int }

func (f *fakeStorage) AddLike(_ context.Context, userID, songID int) error {
	for _, id := range f.liked[userID] {
		if id == songID {
			return nil
		}
	}
	f.liked[userID] = append([]int{songID}, f.liked[userID]...)
	return nil
}
func (f *fakeStorage) RemoveLike(_ context.Context, userID, songID int) error {
	out := f.liked[userID][:0]
	for _, id := range f.liked[userID] {
		if id != songID {
			out = append(out, id)
		}
	}
	f.liked[userID] = out
	return nil
}
func (f *fakeStorage) LikedSongIDs(_ context.Context, userID int) ([]int, error) {
	return append([]int{}, f.liked[userID]...), nil
}

func setup(t *testing.T) (http.Handler, func(userID int, typ authjwt.TokenType) string) {
	t.Helper()
	m, _ := authjwt.NewManager("test-secret-that-is-long-enough-for-hs256")
	h, err := New(nil, likes.New(&fakeStorage{liked: map[int][]int{}}), m, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	token := func(userID int, typ authjwt.TokenType) string {
		tok, _, _ := m.Issue(authjwt.Claims{UserID: userID, Type: typ}, time.Minute)
		return tok
	}
	return h, token
}

func do(h http.Handler, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLikesRequireAccessToken(t *testing.T) {
	h, token := setup(t)
	if rec := do(h, http.MethodGet, "/likes", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: status %d", rec.Code)
	}
	if rec := do(h, http.MethodPut, "/likes/1", token(1, authjwt.TypeRefresh)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh token: status %d", rec.Code)
	}
}

func TestLikeIsIdempotentAndPerUser(t *testing.T) {
	h, token := setup(t)
	egor, anna := token(1, authjwt.TypeAccess), token(2, authjwt.TypeAccess)

	for i := 0; i < 2; i++ { // двойной PUT не создаёт дубль
		if rec := do(h, http.MethodPut, "/likes/5", egor); rec.Code != http.StatusNoContent {
			t.Fatalf("PUT: status %d", rec.Code)
		}
	}
	do(h, http.MethodPut, "/likes/7", egor)

	var resp LikesResponse
	_ = json.Unmarshal(do(h, http.MethodGet, "/likes", egor).Body.Bytes(), &resp)
	if len(resp.SongIDs) != 2 || resp.SongIDs[0] != 7 || resp.SongIDs[1] != 5 {
		t.Fatalf("egor likes = %v, want [7 5]", resp.SongIDs)
	}

	_ = json.Unmarshal(do(h, http.MethodGet, "/likes", anna).Body.Bytes(), &resp)
	if len(resp.SongIDs) != 0 {
		t.Fatalf("anna must not see egor's likes: %v", resp.SongIDs)
	}

	do(h, http.MethodDelete, "/likes/5", egor)
	_ = json.Unmarshal(do(h, http.MethodGet, "/likes", egor).Body.Bytes(), &resp)
	if len(resp.SongIDs) != 1 || resp.SongIDs[0] != 7 {
		t.Fatalf("after unlike = %v, want [7]", resp.SongIDs)
	}
}

func TestLikeInvalidSongID(t *testing.T) {
	h, token := setup(t)
	for _, path := range []string{"/likes/abc", "/likes/0", "/likes/-3"} {
		if rec := do(h, http.MethodPut, path, token(1, authjwt.TypeAccess)); rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status %d, want 400", path, rec.Code)
		}
	}
}
