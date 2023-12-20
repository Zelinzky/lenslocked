package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"lenslocked/rand"
)

type PasswordReset struct {
	ID     int `db:"id"`
	UserID int `db:"user_id"`
	// Token is only set when a PasswordReset is being created.
	Token     string    `db:"token"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
}

const DefaultResetDuration = 30 * time.Minute

type PasswordResetService struct {
	DB            *sqlx.DB
	BytesPerToken int
	Duration      time.Duration
}

func (p *PasswordResetService) Create(email string) (*PasswordReset, error) {
	// check if we have a valid email address
	email = strings.ToLower(email)
	var userID int
	query := `SELECT id FROM users WHERE email = $1`
	err := p.DB.Get(&userID, query, email)
	if err != nil {
		// TODO: consider returning a specific error when the user does not exist
		return nil, fmt.Errorf("create: %w", err)
	}

	// Build the passwordReset
	bytesPerToken := p.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	duration := p.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}

	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: p.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	// Insert the pwReset to the db
	query = `
		INSERT INTO password_resets (user_id, token_hash, expires_at) 
		VALUES (:user_id, :token_hash, :expires_at) ON CONFLICT (user_id) DO 
		UPDATE SET token_hash = :token_hash, expires_at = :expires_at RETURNING id;
	`
	err = sqlxnDB{p.DB}.namedGet(&pwReset.ID, query, pwReset)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	log.Println(token)

	return &pwReset, nil
}

func (p *PasswordResetService) Consume(token string) (*User, error) {
	tokenHash := p.hash(token)
	var dao struct {
		User    User          `db:"user"`
		PwReset PasswordReset `db:"pw_reset"`
	}
	// TODO: use this as example on retrieval of multiple items from single query
	query := `SELECT password_resets.id "pw_reset.id", password_resets.expires_at "pw_reset.expires_at",
		users.id "user.id", users.email "user.email", users.password_hash "user.password_hash"
		FROM password_resets JOIN users ON users.id = password_resets.user_id
		WHERE password_resets.token_hash = $1;`
	err := p.DB.Get(&dao, query, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	if time.Now().After(dao.PwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", token)
	}
	err = p.delete(dao.PwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	out := dao.User

	return &out, nil
}

func (p *PasswordResetService) delete(id int) error {
	query := `DELETE FROM password_resets WHERE id = $1`
	_, err := p.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (p *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
