package models

// Using the aws-sdk-go-v2 library

// The DynamoDB table schema
type DynamoUserStats struct {
	UserId      int   `dynamodbav:"userId"`
	NumLikes    int   `dynamodbav:"numLikes"`
	NumDislikes int   `dynamodbav:"numDislikes"`
	MatchList   []int `dynamodbav:"matchList"`
}
