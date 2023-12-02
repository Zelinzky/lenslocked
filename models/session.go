package models

import "github.com/jmoiron/sqlx"

type Session struct {
	ID     int `db:"id"`
	UserID int `db:"user_id"`
	// Token is only set when createing a news session. When looking upa seddsion
	// this will be left empaty, as we only store the hash of a session token
	// in our db and we are not able to reverse it into a raw token.
	Token     string `db:"token"`
	TokenHash string `db:"token_hash"`
}

type SessionService struct {
	DB *sqlx.DB
}

func (s *SessionService) Create(userID int) (*Session, error) {
	// TODO: Create the session token
	// TODO: Implement session service.create
	return nil, nil
}

func (s *SessionService) User(token string) (*User, error) {
	// TODO: Implement SessionService.User
	return nil, nil
}
