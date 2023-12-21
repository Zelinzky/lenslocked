package models

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Zelinzky/go-sqlf"
	"github.com/jmoiron/sqlx"

	"lenslocked/rand"
)

type PasswordReset struct {
	ID     int `db:"id"`
	UserID int `db:"user_id"`
	// The Token is only set when a PasswordReset is being created.
	Token     string    `db:"token"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
}

//go:embed password_reset.sql
var passwordResetQueriesFile string

var passwordResetQueries map[string]string

func init() {
	passwordResetQueries = sqlf.Load(passwordResetQueriesFile)
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
	err := p.DB.Get(&userID, passwordResetQueries["getUserID"], email)
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
	err = sqlf.NamedDB{DB: p.DB}.NamedGet(&pwReset.ID, passwordResetQueries["create"], pwReset)
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
	err := p.DB.Get(&dao, passwordResetQueries["consume"], tokenHash)
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
	_, err := p.DB.Exec(passwordResetQueries["delete"], id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (p *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
