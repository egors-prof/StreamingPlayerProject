package domain

type Song struct {
	ID          int
	Title       string
	Artist      string
	Album       string
	CoverPath   string // пусто, если обложки нет
	DurationSec int
}

// Page — страница списка с курсором для следующей.
type Page struct {
	Songs       []Song
	NextAfterID int // 0 — следующей страницы нет
}
