package models

type UserStats struct {
	ID       int   `bson:"_id"`
	Likes    int   `bson:"likes"`
	Dislikes int   `bson:"dislikes"`
	Matches  []int `bson:"matches"`
}
