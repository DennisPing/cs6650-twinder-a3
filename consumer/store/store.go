package store

import (
	"sync"
)

// An in-memory kv store

type UserStats struct {
	Likes    int
	Dislikes int
	Matches  []string // swipees list
}

type SimpleStore struct {
	db    map[string]UserStats
	mutex sync.Mutex
}

func NewStore() *SimpleStore {
	return &SimpleStore{
		db: make(map[string]UserStats),
	}
}

// Get the stats of a userId (swiper)
func (s *SimpleStore) GetUserStats(userId string) (UserStats, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, found := s.db[userId]
	return stats, found
}

// Add a userId (swiper) and update swipe direction count
func (s *SimpleStore) Add(userId, swipee, swipeDir string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, found := s.db[userId]
	if !found {
		stats = UserStats{}
	}

	if swipeDir == "right" {
		stats.Likes++
		stats.Matches = append(stats.Matches, swipee)
	} else if swipeDir == "left" {
		stats.Dislikes++
	}

	s.db[userId] = stats
}

// Get a copy of all the SimpleStore data
func (s *SimpleStore) GetAllUserStats() map[string]UserStats {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result := make(map[string]UserStats)
	for k, v := range s.db {
		result[k] = v
	}
	return result
}
