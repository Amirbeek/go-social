package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash
	return nil
}

func (p *password) Compare(p2 string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(p2))
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {
	q := `
        INSERT INTO users (username, password, email, role_id)
        VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4))
        RETURNING id, created_at 
    `
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	role := u.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		q,
		u.Username,
		u.Password.hash,
		u.Email,
		role,
	).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pqe *pq.Error
		if errors.As(err, &pqe) && pqe.Code == "23505" {
			switch pqe.Constraint {
			case "users_email_key":
				return ErrDuplicateEmail
			case "users_username_key":
				return ErrDuplicateUsername
			}
		}
		return err
	}
	return nil
}

// ✅ Fixed: function name changed from GetByID → GetById (to match interface)
func (s *UserStore) GetById(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.password, u.created_at,
			r.id, r.name, r.level, r.description
		FROM users u
		JOIN roles r ON (u.role_id = r.id)
		WHERE u.id = $1 AND COALESCE(u.is_active, TRUE) = TRUE
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) DeleteById(ctx context.Context, id int64) error {
	const q = `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete users id=%d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rowsAffected users id=%d: %w", id, err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, u *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.Create(ctx, tx, u); err != nil {
			return err
		}
		if err := s.createUserInvitation(ctx, tx, token, invitationExp, u.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	const q = `INSERT INTO user_invitations (user_id, token, expiry) VALUES ($1, $2, $3)`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	expiryTime := time.Now().Add(exp)
	_, err := tx.ExecContext(ctx, q, userID, token, expiryTime)
	return err
}

// INFO: ACTIVATE USER FUNCTIONS
func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(ctx context.Context, tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	q := `
		SELECT 
			u.id, 
			u.username, 
			u.email, 
			u.created_at, 
			u.is_active
		FROM users AS u
		JOIN user_invitations AS ui 
			ON u.id = ui.user_id
		WHERE ui.token = $1 
		  AND ui.expiry > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var user User
	err := tx.QueryRowContext(ctx, q, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	return err
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	const q = `DELETE FROM user_invitations WHERE user_id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, q, userID)
	return err
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.delete(ctx, tx, id); err != nil {
			return err
		}
		if err := s.deleteUserInvitations(ctx, tx, id); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	q := `DELETE FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	res, err := tx.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	q := `SELECT id, username, email, password, created_at 
	      FROM users 
	      WHERE email = $1 AND is_active = true`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(ctx, q, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
