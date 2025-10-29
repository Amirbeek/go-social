package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	NotRowEffectedError  = errors.New("error rows affected")
	DeleteError          = errors.New("error deleting record")
	ErrEditConflict      = errors.New("edit conflict")
	QueryTimeOutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int64) (*Post, error)
		DeleteById(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		GetById(context.Context, int64) (*User, error)
		DeleteById(context.Context, int64) error
		CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error
		Create(ctx context.Context, db *sql.Tx, user *User) error
		Activate(ctx context.Context, token string) error
		Delete(context.Context, int64) error
		GetUserByEmail(context.Context, string) (*User, error)
	}
	Comments interface {
		GetByPostId(context.Context, int64) ([]*Comment, error)
		Create(context.Context, *Comment) error
	}
	Follower interface {
		Follow(ctx context.Context, followerId, userId int64) error
		UnFollow(ctx context.Context, followerId, userId int64) error
	}
	Roles interface {
		GetByName(ctx context.Context, roleName string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostsStore{db},
		Users:    &UserStore{db},
		Comments: &CommentsStore{db},
		Follower: &FollowerStore{db},
		Roles:    &RolesStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
