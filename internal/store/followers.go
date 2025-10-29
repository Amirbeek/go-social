package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Follower struct {
	UserId     int64     `json:"user_id"`
	FollowerId int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (f *FollowerStore) Follow(ctx context.Context, followerId, userId int64) error {
	if followerId == userId {
		return errors.New("follower can not follow to own account")
	}
	q := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := f.db.ExecContext(ctx, q, userId, followerId)

	return err
}

func (f *FollowerStore) UnFollow(ctx context.Context, followerId, userId int64) error {
	q := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := f.db.ExecContext(ctx, q, userId, followerId)
	return err
}
