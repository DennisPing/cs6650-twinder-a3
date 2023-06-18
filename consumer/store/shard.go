package store

// Get the DynamoDB table shard (5 shards)
func getTableShard(userId int) string {
	switch {
	case userId <= 1000:
		return "SwipeData1"
	case userId <= 2000:
		return "SwipeData2"
	case userId <= 3000:
		return "SwipeData3"
	case userId <= 4000:
		return "SwipeData4"
	default:
		return "SwipeData5"
	}
}
