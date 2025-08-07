package sqlite_repo

import (
	"context"
	"errors"

	"github.com/DennisPing/cs6650-twinder-a3/lib/models"
)

// If you want to implement SQLite in the future

type SQLiteClient struct {
	// SQLite connection
}

// Create a Repository for SQLite
func NewSQLiteClient( /** Inject SQLite conn here **/ ) *SQLiteClient {
	// Create table if not exists
	return &SQLiteClient{
		// Set the conn
	}
}

// Implement repository methods here...

func (sq *SQLiteClient) GetUserStats(ctx context.Context, userId int) (bool, models.UserStats, error) {
	return false, models.UserStats{}, errors.New("GetUserStats not implemented")
}

func (d *SQLiteClient) GetMatches(ctx context.Context, userId int) (bool, models.UserMatches, error) {
	return false, models.UserMatches{}, errors.New("GetMatches not implemented")
}

func (d *SQLiteClient) UpdateUserStats(ctx context.Context, userId, swipee int, swipeDir string) error {
	return errors.New("UpdateUserStats not implemented")
}
