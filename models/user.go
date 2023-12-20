package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrEmailTaken :
	// A common pattern is to add the package as a prefix to the error for context.
	ErrEmailTaken = errors.New("models: email address is already in use")
)

type User struct {
	ID           int    `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type UserService struct {
	DB *sqlx.DB
}

func (us *UserService) Create(email, password string) (*User, error) {
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
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, ErrEmailTaken
			}
		}
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

func (us *UserService) Authenticate(email, password string) (*User, error) {
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

func (us *UserService) UpdatePassword(userID int, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	passwordHash := string(hashedBytes)
	query := `UPDATE users SET password_hash = $2 WHERE id = $1;`
	_, err = us.DB.Exec(query, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
