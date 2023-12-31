package models

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/Zelinzky/go-sqlf"
	"github.com/jmoiron/sqlx"

	"lenslocked/rand"
)

const MinBytesPerToken = 32

type Session struct {
	ID     int `db:"id"`
	UserID int `db:"user_id"`
	// Token is only set when createing a news session. When looking upa seddsion
	// this will be left empaty, as we only store the hash of a session token
	// in our db, and we are not able to reverse it into a raw token.
	Token     string `db:"token"`
	TokenHash string `db:"token_hash"`
}

type SessionService struct {
	DB            *sqlx.DB
	BytesPerToken int
}

//go:embed session.sql
var sessionQueriesFile string

var sessionQueries map[string]string

func init() {
	sessionQueries = sqlf.Load(sessionQueriesFile)
}

func (s *SessionService) Create(userID int) (*Session, error) {
	bytesPerToken := s.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}

	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: s.hash(token),
	}

	err = sqlf.NamedDB{DB: s.DB}.NamedGet(&session.ID, sessionQueries["create"], session)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &session, nil
}

func (s *SessionService) User(token string) (*User, error) {
	tokenHash := s.hash(token)
	var user User
	err := s.DB.Get(&user, sessionQueries["user"], tokenHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	return &user, nil
}

func (s *SessionService) Delete(token string) error {
	tokenHash := s.hash(token)
	_, err := s.DB.Exec(sessionQueries["delete"], tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (s *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
