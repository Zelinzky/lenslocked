package models

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Zelinzky/go-sqlf"
	"github.com/jmoiron/sqlx"
)

type Gallery struct {
	ID     int    `db:"id"`
	UserID int    `db:"user_id"`
	Title  string `db:"title"`
}

//go:embed gallery.sql
var galleryQueryFile string

var galleryQueries map[string]string

func init() {
	galleryQueries = sqlf.Load(galleryQueryFile)
}

type GalleryService struct {
	DB *sqlx.DB
}

func (g *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		Title:  title,
		UserID: userID,
	}
	err := sqlf.NamedDB{DB: g.DB}.NamedGet(&gallery.ID, galleryQueries["create"], gallery)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}
	return &gallery, nil
}

func (g *GalleryService) ByID(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}
	err := sqlf.NamedDB{DB: g.DB}.NamedGet(&gallery, galleryQueries["by_id"], gallery)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get gallery by ID: %w", err)
	}
	return &gallery, nil
}

func (g *GalleryService) ByUserID(userID int) ([]Gallery, error) {
	var galleries []Gallery
	err := g.DB.Select(&galleries, galleryQueries["by_user_id"], userID)
	if err != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}
	return galleries, nil
}

func (g *GalleryService) Update(gallery *Gallery) error {
	_, err := g.DB.NamedExec(galleryQueries["update"], *gallery)
	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}
	return nil
}

func (g *GalleryService) Delete(id int) error {
	_, err := g.DB.Exec(galleryQueries["delete"], id)
	if err != nil {
		return fmt.Errorf("delete gallery by id: %w", err)
	}
	return nil
}
