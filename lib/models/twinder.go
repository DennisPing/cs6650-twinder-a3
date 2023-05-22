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

// Server side GET /swipes?userId=1234
type TwinderUserStats struct {
	UserId   string `json:"userId"`
	Likes    int    `json:"likes"`
	Dislikes int    `json:"dislikes"`
}

// Server side GET /matches?userId=1234
type TwinderMatches struct {
	UserId  string   `json:"userId"`
	Matches []string `json:"matches"`
}

// Server side debugging GET /matches/all
type AllTwinderUserStats struct {
	UsersStats []TwinderUserStats `json:"users_stats"`
}
