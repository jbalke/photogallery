package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// Gallery is our image container resource
type Gallery struct {
	gorm.Model
	UserID uint     `gorm:"not null;index"`
	Title  string   `gorm:"not null"`
	Images []string `gorm:"-"`
}

type GalleryService interface {
	GalleryDB
}

type GalleryDB interface {
	ByID(id uint) (*Gallery, error)
	ByUserID(id uint) ([]Gallery, error)
	Create(gallery *Gallery) error
	Delete(id uint) error
	Update(gallery *Gallery) error
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{&galleryGorm{db}},
	}
}

type galleryService struct {
	GalleryDB
}

type galleryValidator struct {
	GalleryDB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFuncs(gallery,
		gv.titleRequired,
		gv.userIDRequired)
	if err != nil {
		return err
	}
	return gv.GalleryDB.Create(gallery)

}

func (gv *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	err := runGalleryValFuncs(&gallery, gv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return gv.GalleryDB.Delete(id)
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFuncs(gallery,
		gv.titleRequired,
		gv.userIDRequired)
	if err != nil {
		return err
	}
	return gv.GalleryDB.Update(gallery)

}

func (gv *galleryValidator) titleRequired(gallery *Gallery) error {
	if strings.TrimSpace(gallery.Title) == "" {
		return ErrTitleRequired
	}
	return nil
}

func (gv *galleryValidator) userIDRequired(gallery *Gallery) error {
	if gallery.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) idGreaterThan(n uint) galleryValFunc {
	return galleryValFunc(func(gallery *Gallery) error {
		if gallery.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	if err != nil {
		return nil, err
	}
	return &gallery, err
}

func (gg *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery
	db := gg.db.Where("user_id = ?", userID).Find(&galleries)
	err := db.Error
	if err != nil {
		return nil, err
	}
	return galleries, err
}

type galleryValFunc func(*Gallery) error

func runGalleryValFuncs(gallery *Gallery, fns ...galleryValFunc) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}
