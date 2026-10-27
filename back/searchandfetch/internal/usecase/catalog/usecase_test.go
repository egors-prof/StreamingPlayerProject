package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/egors-prof/searchService/internal/domain"
	"github.com/egors-prof/searchService/internal/errs"
)

type fakeRepo struct{ songs []domain.Song }

func (f *fakeRepo) List(_ context.Context, afterID, limit int) ([]domain.Song, error) {
	var out []domain.Song
	for _, s := range f.songs {
		if s.ID > afterID && len(out) < limit {
			out = append(out, s)
		}
	}
	return out, nil
}
func (f *fakeRepo) Search(context.Context, string, int) ([]domain.Song, error) { return f.songs, nil }
func (f *fakeRepo) GetByIDs(_ context.Context, ids []int) ([]domain.Song, error) {
	var out []domain.Song
	for _, s := range f.songs { // порядок как в «БД», не как в запросе
		for _, id := range ids {
			if s.ID == id {
				out = append(out, s)
				break
			}
		}
	}
	return out, nil
}

func songs(n int) []domain.Song {
	out := make([]domain.Song, n)
	for i := range out {
		out[i] = domain.Song{ID: i + 1}
	}
	return out
}

func TestListKeysetPagination(t *testing.T) {
	uc := New(&fakeRepo{songs: songs(5)})
	ctx := context.Background()

	first, _ := uc.List(ctx, 0, 2)
	if len(first.Songs) != 2 || first.NextAfterID != 2 {
		t.Fatalf("page 1: %d songs, next %d", len(first.Songs), first.NextAfterID)
	}
	second, _ := uc.List(ctx, first.NextAfterID, 2)
	if second.Songs[0].ID != 3 || second.NextAfterID != 4 {
		t.Fatalf("page 2 starts at %d, next %d", second.Songs[0].ID, second.NextAfterID)
	}
	last, _ := uc.List(ctx, second.NextAfterID, 2)
	if len(last.Songs) != 1 || last.NextAfterID != 0 {
		t.Fatalf("last page: %d songs, next %d (want 1 song, no next)", len(last.Songs), last.NextAfterID)
	}
}

func TestGetByIDsKeepsRequestOrder(t *testing.T) {
	uc := New(&fakeRepo{songs: songs(5)})
	got, err := uc.GetByIDs(context.Background(), []int{4, 1, 99, 4, 2})
	if err != nil {
		t.Fatal(err)
	}
	want := []int{4, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("got %d songs, want %d", len(got), len(want))
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("position %d: id %d, want %d", i, got[i].ID, id)
		}
	}
}

func TestSearchValidation(t *testing.T) {
	uc := New(&fakeRepo{songs: songs(3)})
	if res, _ := uc.Search(context.Background(), "   ", 10); len(res) != 0 {
		t.Fatal("blank query must return nothing")
	}
	long := make([]rune, MaxQueryLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := uc.Search(context.Background(), string(long), 10); !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}
