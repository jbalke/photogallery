package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"lenslocked.com/hash"
	"lenslocked.com/rand"

	"golang.org/x/crypto/bcrypt"
)

const userPwPepper = "7SZ5t9epC5RFv&*"
const hmacSecretKey = "secret-key"

var (
	// ErrNotFound is returned when a resource can not be found in the DB.
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID is provided to a method like Delete.
	ErrInvalidID = errors.New("models: ID provided is invalid")

	// ErrInvalidPassword is returned when the credentials provided to Authenticate() are incorrect.
	ErrInvalidPassword = errors.New("models: Incorrect Password")
)

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &UserService{
		db:   db,
		hmac: hmac,
	}, nil
}

type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

// ByID will lookup by the id provided.
// 1 - user, nil
// 2 - nil, ErrNotFound
// 3 - nil, otherError
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// ByEmail looks up the user with the given email address.
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// ByRemember looks up the user with the given remember token.
// This method will handle hashing the token for comparison with stored hashed tokens.
func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	hashedToken := us.hmac.Hash(token)
	db := us.db.Where("remember_hash = ?", hashedToken)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// hashPassword is a helper function to return a hash of the user's password.
func hashPassword(password string) (string, error) {
	pwBytes := []byte(password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Create will create the provided user and backfill data
// like the ID, CreatedAt and UpdatedAt fields.
func (us *UserService) Create(user *User) error {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword
	user.Password = ""

	if user.Remember == "" {
		rememberToken, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = rememberToken
	}
	user.RememberHash = us.hmac.Hash(user.Remember)
	return us.db.Create(user).Error
}

// Authenticate checks for a user with mathcing email and password.
func (us *UserService) Authenticate(email string, password string) (*User, error) {
	pwBytes := []byte(password + userPwPepper)
	user, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), pwBytes)
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}
	return user, nil
}

// Update will update the persisted user with the provided user instance.
func (us *UserService) Update(user *User) error {
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = ""
		user.PasswordHash = hashedPassword
	}
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
	return us.db.Save(user).Error
}

// UpdateRememberHash will update the remember hash stored on the user.
func (us *UserService) UpdateRememberHash(user *User) error {
	return us.db.Model(user).Update("remember_hash", user.RememberHash).Error
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
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"` // do not store in the DB
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// first will query using the provided gorm.DB and will
// get the first item returned and place in the provided dst.
// If nothing is found it will return ErrNotFound.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
