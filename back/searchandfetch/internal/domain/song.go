package domain

import "time"

type Song struct{
	Title string
	Duration time.Time
	PhotoPath string
	AlbumTitle string
	Pseudonym string 
	CreatedAt time.Time
}



