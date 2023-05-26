package models

// Client side swipe
type SwipeRequest struct {
	Swiper  string `json:"swiper"`
	Swipee  string `json:"swipee"`
	Comment string `json:"comment"`
}

// Server side swipe with the direction payload
type SwipePayload struct {
	Swiper    string `json:"swiper"`
	Swipee    string `json:"swipee"`
	Comment   string `json:"comment"`
	Direction string `json:"direction"`
}

// Server side GET /stats/{userId}/
type UserStats struct {
	UserId      string `json:"userId"`
	NumLikes    int    `json:"numLikes"`
	NumDislikes int    `json:"numDislikes"`
}

// Server side GET /matches/{userId}/
type Matches struct {
	UserId    string   `json:"userId"`
	MatchList []string `json:"matchList"`
}

// Server side debugging GET /stats/all
type AllUserStats struct {
	UsersStats []UserStats `json:"allUsersStats"`
}
