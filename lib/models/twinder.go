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
	NumLikes    int      `json:"numLikes,omitempty"`
	NumDislikes int      `json:"numDislikes,omitempty"`
	MatchList   []string `json:"matchList,omitempty"`
}
