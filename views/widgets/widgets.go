package widgets

import (
	"fmt"
	"math"
	"net/url"
)

type AlertLevel string

const (
	AlertLevelPrimary AlertLevel = "alert-primary"
	AlertLevelInfo    AlertLevel = "alert-info"
	AlertLevelSuccess AlertLevel = "alert-success"
	AlertLevelWarning AlertLevel = "alert-warning"
	AlertLevelDanger  AlertLevel = "alert-danger"
)

type Alert struct {
	Message string
	Level   AlertLevel
}

type ItemCard struct {
	ItemName string
	Brand    string
	ImageURL string
	ImageAlt string
	Href     string
}

type MoreLike struct {
	Title string
	Items []ItemCard
}

type Pageination struct {
	CurrentPage int
	TotalPages  int
	BaseURL     url.URL
}

type pageinationData struct {
	Page    int
	PageUrl string
}

func (p Pageination) PagesAvailable() []pageinationData {
	startPage := p.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}

	endPage := startPage + 4
	if endPage > p.TotalPages {
		endPage = p.TotalPages
	}

	pages := []pageinationData{}
	for i := startPage; i <= endPage; i++ {

		u := p.BaseURL
		q := u.Query()
		q.Set("page", fmt.Sprintf("%d", i))
		u.RawQuery = q.Encode()
		page := pageinationData{
			Page:    i,
			PageUrl: u.String(),
		}
		pages = append(pages, page)
	}

	return pages
}

type Rating struct {
	Rating float64
	Max    int
}

func (r Rating) FullStars() []int {
	return make([]int, int(math.Floor(r.Rating)))
}

func (r Rating) HalfStar() bool {
	fraction := r.Rating - math.Floor(r.Rating)
	return fraction >= 0.25 && fraction < 0.75
}

func (r Rating) EmptyStars() []int {
	emptyCount := r.Max - len(r.FullStars())
	if r.HalfStar() {
		emptyCount -= 1
	}
	return make([]int, emptyCount)
}
