package models

import (
	"fmt"
	"io"
	"os"
)

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) (int64, error)
	//ByGalleryID(galleryID uint) []string
}

type imageValidator struct {
	ImageService
}

var _ ImageService = &imageService{}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) (int64, error) {
	defer r.Close()

	path, err := is.mkImagePath(galleryID)
	if err != nil {
		return 0, err
	}

	// create destination file
	dst, err := os.Create(path + filename)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	// copy file to destination
	bytes, err := io.Copy(dst, r)
	if err != nil {
		return 0, err
	}

	return bytes, nil
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := fmt.Sprintf("images/galleries/%v/", galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}
