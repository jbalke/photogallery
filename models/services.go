package models

import "github.com/jinzhu/gorm"

func NewServices(connectionInfo string, logging bool) (*Services, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(logging)
	return &Services{
		User: NewUserService(db),
	}, nil
}

// Services contains all of our services
type Services struct {
	Gallery GalleryService
	User    UserService
}

func (s *Services) CloseDB() error {

}
