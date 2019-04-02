package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	// ErrNotFound is returned when a resource can not be found in the DB.
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID is provided to a method like Delete.
	ErrInvalidID = errors.New("models: ID provided is invalid")
)

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

type UserService struct {
	db *gorm.DB
}

// ByID will lookup by the id provided.
// 1 - user, nil
// 2 - nil, ErrNotFound
// 3 - nil, otherError
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)

	return &user, err
}

// ByEmail looks up the user with theh given email address.
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)

	return &user, err
}

// Create will create the provided user and backfill data
// like the ID, CreatedAt and UpdatedAt fields.
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Update will update the persisted user with the provided user instance.
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Delete will delete the given user from the db.
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

// Close closes the UserService database connection.
func (us *UserService) Close() error {
	return us.db.Close()
}

// DestructiveReset drops the user table and rebuilds it.
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate the users table.
func (us *UserService) AutoMigrate() error {
	return us.db.AutoMigrate(&User{}).Error
}

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;unique_index"`
}
