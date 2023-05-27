package models

// Using the aws-sdk-go-v2 library

// MatchList is stored as a NumberSet to avoid duplicates
type DynamoUserStats struct {
	UserId      int                 `dynamodbav:"userId"`
	NumLikes    int                 `dynamodbav:"numLikes"`
	NumDislikes int                 `dynamodbav:"numDislikes"`
	MatchList   map[string]struct{} `dynamodbav:"matchList"`
}
