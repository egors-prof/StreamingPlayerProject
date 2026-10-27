// Package catalog — просмотр и поиск песен.
package catalog

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/egors-prof/searchService/internal/domain"
	"github.com/egors-prof/searchService/internal/errs"
	"github.com/egors-prof/searchService/internal/port/driven"
)

const (
	DefaultLimit = 20
	MaxLimit     = 50
	MaxBatchIDs  = 100
	MaxQueryLen  = 100
)

type UseCase struct {
	songs driven.SongRepository
}

func New(songs driven.SongRepository) *UseCase { return &UseCase{songs: songs} }

func normalizeLimit(limit int) int {
	switch {
	case limit <= 0:
		return DefaultLimit
	case limit > MaxLimit:
		return MaxLimit
	}
	return limit
}

// List — keyset-пагинация: клиент передаёт ID последней полученной песни.
// В отличие от OFFSET, база не читает и не выбрасывает пропущенные строки,
// а вставки между запросами не дают дублей и пропусков.
func (u *UseCase) List(ctx context.Context, afterID, limit int) (domain.Page, error) {
	if afterID < 0 {
		return domain.Page{}, fmt.Errorf("%w: after_id must be >= 0", errs.ErrInvalidInput)
	}
	limit = normalizeLimit(limit)
	// Запрашиваем на одну больше, чтобы узнать, есть ли следующая страница.
	songs, err := u.songs.List(ctx, afterID, limit+1)
	if err != nil {
		return domain.Page{}, err
	}
	page := domain.Page{Songs: songs}
	if len(songs) > limit {
		page.Songs = songs[:limit]
		page.NextAfterID = page.Songs[limit-1].ID
	}
	return page, nil
}

func (u *UseCase) Search(ctx context.Context, query string, limit int) ([]domain.Song, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []domain.Song{}, nil
	}
	if utf8.RuneCountInString(query) > MaxQueryLen {
		return nil, fmt.Errorf("%w: query is too long", errs.ErrInvalidInput)
	}
	return u.songs.Search(ctx, query, normalizeLimit(limit))
}

// GetByIDs возвращает песни в порядке переданных ID (нужно странице лайков:
// новые лайки первыми). Несуществующие ID пропускаются.
func (u *UseCase) GetByIDs(ctx context.Context, ids []int) ([]domain.Song, error) {
	if len(ids) > MaxBatchIDs {
		return nil, fmt.Errorf("%w: at most %d ids", errs.ErrInvalidInput, MaxBatchIDs)
	}
	if len(ids) == 0 {
		return []domain.Song{}, nil
	}
	found, err := u.songs.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]domain.Song, len(found))
	for _, s := range found {
		byID[s.ID] = s
	}
	ordered := make([]domain.Song, 0, len(found))
	for _, id := range ids {
		if s, ok := byID[id]; ok {
			ordered = append(ordered, s)
			delete(byID, id) // повторный ID в запросе не даёт дубль в ответе
		}
	}
	return ordered, nil
}

func (u *UseCase) Get(ctx context.Context, id int) (domain.Song, error) {
	songs, err := u.songs.GetByIDs(ctx, []int{id})
	if err != nil {
		return domain.Song{}, err
	}
	if len(songs) == 0 {
		return domain.Song{}, errs.ErrNotFound
	}
	return songs[0], nil
}
