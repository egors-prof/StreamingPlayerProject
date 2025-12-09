package repository

import (
	"context"
	"fmt"
	"strings"

	"time"

	"github.com/egors-prof/searchService/internal/domain"
)

type Song struct {
	Title      string `db:"title"`
	Duration   string `db:"duration"`
	PhotoPath  string `db:"photo_path"`
	AlbumTitle string `db:"album_title"`
	Pseudonym  string `db:"pseudonym"`
	CreatedAt  string `db:"created_at"`
}

func (rs *Song) FromRepositoryToDomain() (*domain.Song, error) {
	duration, err := time.Parse(time.RFC3339, rs.Duration)
	if err != nil {
		return &domain.Song{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rs.CreatedAt)
	if err != nil {
		return &domain.Song{}, err
	}
	return &domain.Song{
		Title:      rs.Title,
		Duration:   duration,
		PhotoPath:  rs.PhotoPath,
		AlbumTitle: rs.AlbumTitle,
		Pseudonym:  rs.Pseudonym,
		CreatedAt:  createdAt,
	}, nil
}

func (r *Repository) GetSongsInfo(ctx context.Context, quant int, offset int) ([]domain.Song, error) {
	songs := make([]Song, 0)
	fmt.Println(quant, offset)
	err := r.DB.SelectContext(ctx, &songs, `
	select songs.title as title,duration as duration, albums.photo_path as photo_path,albums.title as album_title,pseudonym,songs.created_at as created_at  from artists 
	right join albums on artists.id= albums.artist_id
	right join songs on songs.album_id=albums.id
	offset $1 limit $2
	`, offset, quant)
	fmt.Println(songs, len(songs))
	if err != nil {
		return nil, err
	}

	dSongs := make([]domain.Song, 0)

	for _, song := range songs {
		dSong, err := song.FromRepositoryToDomain()
		if err != nil {
			return nil, err
		}
		dSongs = append(dSongs, *dSong)
	}

	return dSongs, nil

}

func (r *Repository) GetSearch(ctx context.Context, search string) ([]domain.Song, error) {
	//songs := make([]Song, 0)
	//fmt.Println("repository")
	//searchTerm := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
	//err := r.DB.SelectContext(ctx, &songs, `
	//	select songs.title as title,duration as duration, albums.photo_path as photo_path,albums.title as album_title,pseudonym,songs.created_at as created_at  from artists
	//	right join albums on artists.id= albums.artist_id
	//	right join songs on songs.album_id=albums.id
	//	where  songs.title like $1
	//`, searchTerm)
	//if err != nil {
	//	fmt.Println(err)
	//	return nil, err
	//}
	//fmt.Println(songs)
	//dSongs := make([]domain.Song, 0)
	searchTerm := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"

	sqlQuery := `
        SELECT songs.title as title,duration as duration, albums.photo_path as photo_path,albums.title as album_title,pseudonym,songs.created_at as created_at  from artists                                                                                                                                                   
        right join albums on artists.id= albums.artist_id
		right join songs on songs.album_id=albums.id
        WHERE LOWER(songs.title) LIKE $1 
           OR LOWER(pseudonym) LIKE $1
           OR LOWER(albums.title) LIKE $1
           
        ORDER BY 
            CASE 
                WHEN LOWER(songs.title) LIKE $2 THEN 1
                WHEN LOWER(pseudonym) LIKE $2 THEN 2
                WHEN LOWER(albums.title) LIKE $2 THEN 3
                ELSE 4
            END,
            title ASC
        LIMIT 50
    `

	exactTerm := strings.ToLower(strings.TrimSpace(search)) + "%"

	var rSongs []Song
	err := r.DB.SelectContext(ctx, &rSongs, sqlQuery, searchTerm, exactTerm)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	fmt.Println(rSongs)
	dSongs := make([]domain.Song, 0)
	for _, song := range rSongs {
		dSong, err := song.FromRepositoryToDomain()
		if err != nil {
			return nil, err
		}
		dSongs = append(dSongs, *dSong)
	}
	return dSongs, nil
}
