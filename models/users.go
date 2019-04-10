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

// UserDB is used to interact with the users model.
//
// If the user is found, we return a niil error.
// If the user is not found, we return ErrNotFound.
// If there is another error, we return an error with
// more information. This error may originate from the DB layer.
//
// For sinlge user queries, any error but ErrNot found should probably
// result in a 500 error.
type UserDB interface {
	// Methods for querying for single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	UpdateRememberHash(user *User) error
	Delete(id uint) error

	// Used to close the DB connection
	Close() error

	// Migration helpers
	AutoMigrate() error
	DestructiveReset() error
}

var (
	// ErrNotFound is returned when a resource can not be found in the DB.
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID is provided to a method like Delete.
	ErrInvalidID = errors.New("models: ID provided is invalid")

	// ErrInvalidPassword is returned when the credentials provided to Authenticate() are incorrect.
	ErrInvalidPassword = errors.New("models: Incorrect Password")
)

// UserService is a set of methods used to work with the user model
type UserService interface {
	// Authenticate will verify the provided email and password. If correct, the matching
	// user will be returned. Otherwise an error will be returned: ErrNotFound, ErrInvalidPassword,
	// or another if something goes wrong.
	Authenticate(email, password string) (*User, error)
	UserDB
}

// NewUserService takes a connection string for the DB and returns a *UserService.
// If the returned error is not nil, there was a problem opening the database.
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	return &userService{
		UserDB: &userValildator{
			UserDB: ug,
		},
	}, nil
}

type userService struct {
	UserDB
}

type userValildator struct {
	UserDB
}

func (uv *userValildator) ByID(id uint) (*User, error) {
	// validate the ID
	if id <= 0 {
		return nil, errors.New("Invalid ID")
	}
	return uv.UserDB.ByID(id)
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// ByEmail looks up the user with the given email address.
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// ByRemember looks up the user with the given remember token.
// This method will handle hashing the token for comparison with stored hashed tokens.
func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User
	hashedToken := ug.hmac.Hash(token)
	db := ug.db.Where("remember_hash = ?", hashedToken)
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
func (ug *userGorm) Create(user *User) error {
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
	user.RememberHash = ug.hmac.Hash(user.Remember)
	return ug.db.Create(user).Error
}

// Authenticate checks for a user with mathcing email and password.
func (us *userService) Authenticate(email, password string) (*User, error) {
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
func (ug *userGorm) Update(user *User) error {
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = ""
		user.PasswordHash = hashedPassword
	}
	if user.Remember != "" {
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}
	return ug.db.Save(user).Error
}

// UpdateRememberHash will update the remember hash stored on the user.
func (ug *userGorm) UpdateRememberHash(user *User) error {
	user.RememberHash = ug.hmac.Hash(user.Remember)
	return ug.db.Model(user).Update("remember_hash", user.RememberHash).Error
}

// Delete will delete the given user from the db.
func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

// Close closes the UserService database connection.
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

// DestructiveReset drops the user table and rebuilds it.
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate the users table.
func (ug *userGorm) AutoMigrate() error {
	return ug.db.AutoMigrate(&User{}).Error
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
