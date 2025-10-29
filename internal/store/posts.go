package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64      `json:"id"`
	Content   string     `json:"content"`
	Title     string     `json:"title"`
	UserId    int64      `json:"userId"`
	Tags      []string   `json:"tags"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Comments  []*Comment `json:"comments"`
	Version   int        `json:"version"`
	User      *User      `json:"user"`
}

type PostsStore struct {
	db *sql.DB
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

func (s *PostsStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	// oq ro‘yxat: faqat ASC yoki DESC
	sortDir := "DESC"
	if strings.EqualFold(fq.Sort, "ASC") {
		sortDir = "ASC"
	}

	query := `
		SELECT
			p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
			u.username,
			COALESCE(COUNT(c.id), 0) AS comments_count
		FROM posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN comments c ON c.post_id = p.id
		WHERE
			(
				p.user_id = $1
				OR EXISTS (
					SELECT 1
					FROM followers f
					WHERE f.user_id = p.user_id AND f.follower_id = $1
				)
			)
			AND (
				p.title   ILIKE '%' || $4 || '%'
				OR p.content ILIKE '%' || $4 || '%'
			)
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + sortDir + `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var p PostWithMetadata
		p.User = &User{}
		if err := rows.Scan(
			&p.ID,
			&p.UserId,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentsCount,
		); err != nil {
			return nil, err
		}
		feed = append(feed, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feed, nil
}

func (s *PostsStore) Create(ctx context.Context, p *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	return s.db.QueryRowContext(
		ctx,
		query,
		p.Content,
		p.Title,
		p.UserId,
		pq.Array(p.Tags),
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (s *PostsStore) GetById(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, content, title, user_id, tags, created_at, updated_at, version
		FROM posts
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserId,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostsStore) DeleteById(ctx context.Context, id int64) error {
	query := `
    DELETE FROM posts WHERE id = $1;
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return DeleteError
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NotRowEffectedError
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostsStore) Update(ctx context.Context, p *Post) error {
	query := `
		UPDATE posts 
		SET title = $1, content = $2, version = version + 1, updated_at = NOW()
		WHERE id = $3 AND version = $4
		RETURNING version;
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, p.Title, p.Content, p.ID, p.Version).Scan(&p.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
