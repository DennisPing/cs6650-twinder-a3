package models

// Client side swipe
type SwipeRequest struct {
	Swiper    string `json:"swiper"`
	Swipee    string `json:"swipee"`
	Comment   string `json:"comment"`
	Direction string `json:"direction,omitempty"`
}

// Server side user stats
type UserStats struct {
	NumLikes    int `json:"numLikes"`
	NumDislikes int `json:"numDislikes"`
}

// Server side user matches
type UserMatches struct {
	MatchList []int `json:"matchList"`
}

// Server side error
type ErrorResponse struct {
	Message string `json:"message"`
}
