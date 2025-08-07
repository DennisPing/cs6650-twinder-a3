package db

import (
	"context"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
)

//go:generate mockery --name=Repository --filename=mock_repository.go
type Repository interface {
	GetUserStats(ctx context.Context, userId int) (models.UserStats, bool, error)
	GetMatches(ctx context.Context, userId int) (models.UserMatches, bool, error)
	UpdateUserStats(ctx context.Context, userId, swipee int, swipeDir string) error
}
