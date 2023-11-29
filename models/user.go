package models

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int    `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type UserService struct {
	DB *sqlx.DB
}

func (us UserService) Create(email, password string) (*User, error) {
	email = strings.ToLower(email)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	passwordHash := string(hashedBytes)
	user := User{
		Email:        email,
		PasswordHash: passwordHash,
	}
	query := `INSERT INTO users (email, password_hash) VALUES (:email, :password_hash) RETURNING id`
	rows, err := us.DB.NamedQuery(query, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&user.ID); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	}

	return &user, nil
}

func (us UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	var user User
	query := `SELECT * FROM users WHERE email=$1`
	err := us.DB.Get(&user, query, email)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	return &user, nil
}
