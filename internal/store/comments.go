package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type CommentsStore struct {
	db *sql.DB
}
type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

func (s *CommentsStore) GetByPostId(ctx context.Context, postID int64) ([]*Comment, error) {
	query := `
    SELECT 
        c.id, c.post_id, c.user_id, c.content, c.created_at, 
        u.id, u.username
    FROM 
        comments c
    JOIN 
        users u ON u.id = c.user_id
    WHERE 
        c.post_id = $1 
    ORDER BY 
        c.created_at DESC
    `
	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("error querying comments: %w", err)
	}
	defer rows.Close()

	comments := []*Comment{}
	for rows.Next() {
		comment := &Comment{}
		comment.User = User{}
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.User.ID,
			&comment.User.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning comment row: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return comments, nil
}

func (s *CommentsStore) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			return fmt.Errorf("postgres error: %s", pqErr.Message)
		}
		return fmt.Errorf("error inserting comment: %w", err)
	}

	return nil
}
