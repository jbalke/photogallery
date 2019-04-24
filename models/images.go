package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	imagePath = "images/galleries"
)

type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) Path() string {
	// contruct a URL with our path and output with String() to
	// ensure that paths are url encoded.
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("%s/%v/%v", imagePath, i.GalleryID, i.Filename)
}

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) (int64, error)
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
	DeleteAll(galleryID uint) error
}

type imageValidator struct {
	ImageService
}

var _ ImageService = &imageValidator{&imageService{}}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}

func (is *imageService) DeleteAll(galleryID uint) error {
	return os.RemoveAll(fmt.Sprintf("%s/%d", imagePath, galleryID))
}

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
	galleryPath := is.imagePath(galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imagePath(galleryID)
	imageStrings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	ret := make([]Image, len(imageStrings))
	for i := range imageStrings {
		imageStrings[i] = fwdSlashSeparators(imageStrings[i])
		imageStrings[i] = strings.Replace(imageStrings[i], path, "", 1)
		ret[i] = Image{
			Filename:  imageStrings[i],
			GalleryID: galleryID,
		}
	}
	return ret, nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("%s/%v/", imagePath, galleryID)
}

// On Windows the path separator "\" triggers html escaping,
// so replace with unix path separators
func fwdSlashSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
