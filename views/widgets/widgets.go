package widgets

import "math"

type Card struct {
	Title    string
	Content  string
	ImageURL string
	ImageAlt string
	Href     string
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
