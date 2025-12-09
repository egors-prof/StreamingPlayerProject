package Server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/egors-prof/searchService/internal/domain"
	"github.com/gin-gonic/gin"
)

type Song struct {
	Title      string `json:"title"`
	Duration   string `json:"duration"`
	PhotoPath  string `json:"photo_path"`
	AlbumTitle string `json:"album_title"`
	Pseudonym  string `json:"pseudonym"`
	CreatedAt  string `json:"created_at"`
}

func FromDomainToHttp(ds domain.Song) *Song {
	duration := ds.Duration.Format(time.TimeOnly)
	createdAt := ds.CreatedAt.Format(time.DateOnly)
	return &Song{
		Title:      ds.Title,
		Duration:   duration,
		PhotoPath:  ds.PhotoPath,
		AlbumTitle: ds.AlbumTitle,
		Pseudonym:  ds.Pseudonym,
		CreatedAt:  createdAt,
	}
}

func (s *Server) GetFirstSongs(c *gin.Context) {
	quantityStr := c.Query("quant")
	offsetStr := c.Query("off")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		// error handling
		fmt.Printf("error occurred %v\n", err)
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		// error handling
		fmt.Printf("error occurred %v\n", err)
		return
	}

	dSongs, err := s.Uc.InfoGetter.GetFirstSongs(c, quantity, offset)
	if err != nil {
		//error handling
		fmt.Printf("error occured %v\n", err)
		return
	}
	hSongs := make([]Song, 0)
	for _, dSong := range dSongs {
		hSong := FromDomainToHttp(dSong)
		hSongs = append(hSongs, *hSong)
	}

	c.JSON(http.StatusOK, hSongs)
}

func (s *Server) GetSearch(c *gin.Context) {
	search := c.Query("q")
	fmt.Printf("search query: %s\n", search)
	dSongs, err := s.Uc.Searcher.GetSearch(c, search)
	if err != nil {
		return
	}
	hSongs := make([]Song, 0)
	for _, dSong := range dSongs {
		hSong := FromDomainToHttp(dSong)
		hSongs = append(hSongs, *hSong)
	}

	c.JSON(http.StatusOK, hSongs)

}

func (s *Server) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"ping": "pong",
	})
}
