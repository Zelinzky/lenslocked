package models

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	// ImagesDir holds the directory where the images are going to be stored
	ImagesDir string
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
	err = os.RemoveAll(g.galleryDir(id))
	if err != nil {
		return fmt.Errorf("delete gallery images: %w", err)
	}
	return nil
}

func (g *GalleryService) galleryDir(id int) string {
	imagesDir := g.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", id))
}

type Image struct {
	GalleryID int
	Path      string
	Filename  string
}

func (g *GalleryService) Images(galleryID int) ([]Image, error) {
	globPattern := filepath.Join(g.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("retrieving gallery images: %w", err)
	}
	var images []Image
	for _, file := range allFiles {
		if hasExtension(file, g.extensions()) {
			images = append(images, Image{
				GalleryID: galleryID,
				Path:      file,
				Filename:  filepath.Base(file),
			})
		}
	}
	return images, nil
}

func (g *GalleryService) Image(galleryID int, filename string) (Image, error) {
	imagePath := filepath.Join(g.galleryDir(galleryID), filename)
	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("querying for image: %w", err)
	}
	return Image{
		GalleryID: galleryID,
		Path:      imagePath,
		Filename:  filename,
	}, nil
}

func (g *GalleryService) DeleteImage(galleryID int, filename string) error {
	image, err := g.Image(galleryID, filename)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	err = os.Remove(image.Path)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}

func (g *GalleryService) CreateImage(galleryID int, filename string, contents io.ReadSeeker) error {
	err := checkContentType(contents, g.imageContentTypes())
	if err != nil {
		return fmt.Errorf("creating image %v: %w", filename, err)
	}
	if !hasExtension(filename, g.extensions()) {
		return fmt.Errorf("creating image %v: %w", filename, err)
	}

	galleryDir := g.galleryDir(galleryID)
	err = os.MkdirAll(galleryDir, 0755)
	if err != nil {
		return fmt.Errorf("creating gallery-%d images directory: %w", galleryID, err)
	}
	imagePath := filepath.Join(galleryDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("creating image file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, contents)
	if err != nil {
		return fmt.Errorf("copying contents to image: %w", err)
	}
	return nil
}

func hasExtension(file string, extensions []string) bool {
	for _, ext := range extensions {
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}

func (g *GalleryService) extensions() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

func (g *GalleryService) imageContentTypes() []string {
	return []string{"image/png", "image/jpeg", "image/gif"}
}
