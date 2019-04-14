package models

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"lenslocked.com/hash"
	"lenslocked.com/rand"

	"golang.org/x/crypto/bcrypt"
)

const (
	userPwPepper      = "7SZ5t9epC5RFv&*"
	hmacSecretKey     = "secret-key"
	minPasswordLength = 6
	maxPasswordLength = 13
)

// User represents the use model in our DB.
// This is used for use accounts, storing both email
// and password to enable access to their content.
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"` // do not store in the DB
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

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

// UserService is a set of methods used to work with the user model
type UserService interface {
	// Authenticate will verify the provided email and password. If correct, the matching
	// user will be returned. Otherwise an error will be returned: ErrNotFound, ErrPasswordIncorrect,
	// or another if something goes wrong.
	Authenticate(email, password string) (*User, error)
	UserDB
}

// NewUserService takes a connection string for the DB and returns a *UserService.
// If the returned error is not nil, there was a problem opening the database.
func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{db}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)

	return &userService{
		UserDB: uv,
	}
}

var _ UserService = &userService{}

type userService struct {
	UserDB
}

// Authenticate checks for a user with mathcing email and password.
func (us *userService) Authenticate(email, password string) (*User, error) {
	user, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrPasswordIncorrect
		default:
			return nil, err
		}
	}
	return user, nil
}

type userValFunc func(*User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

var _ UserDB = &userValidator{}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
}

// ByEmail will normalise the email address before calling ByEmail on UserDB.
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	err := runUserValFuncs(&user, uv.normaliseEmail)
	if err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) ByID(id uint) (*User, error) {
	// validate the ID
	if id <= 0 {
		return nil, errors.New("Invalid ID")
	}
	return uv.UserDB.ByID(id)
}

// ByRemember will hash the remember token and then calls
// ByRemember on the gorm DB layer.
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	err := runUserValFuncs(&user, uv.hmacRemember)
	if err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValFuncs(user,
		uv.requireName,
		uv.requireEmail,
		uv.requirePassword,
		uv.emailFormat,
		uv.normaliseEmail,
		uv.emailIsAvail,
		uv.passwordIsComplex(minPasswordLength, maxPasswordLength),
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setDefaultRemember,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
	)
	if err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

// Delete will delete the given user from the db.
func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValFuncs(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValFuncs(user,
		uv.requireName,
		uv.requireEmail,
		uv.passwordIsComplex(minPasswordLength, maxPasswordLength),
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.emailFormat,
		uv.normaliseEmail,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

// UpdateRememberHash takes a user and persists the hashed remember token if a remember token is set.
func (uv *userValidator) UpdateRememberHash(user *User) error {
	err := runUserValFuncs(user, uv.hmacRemember)
	if err != nil {
		return err
	}
	return uv.UserDB.UpdateRememberHash(user)
}

// bcryptPassword is a helper function to return a hash of the user's password.
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) setDefaultRemember(user *User) error {
	if user.Remember != "" {
		return nil
	}
	rememberToken, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = rememberToken
	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < rand.RememberTokenBytes {
		return ErrRememberTooShort
	}
	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberHashRequired
	}
	return nil
}

func (uv *userValidator) idGreaterThan(n uint) userValFunc {
	return userValFunc(func(user *User) error {
		if user.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

func (uv *userValidator) normaliseEmail(user *User) error {
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		// Email address is not taken
		return nil
	}
	if err != nil {
		return err
	}
	// We found a user with this email address.
	// If the found user ID does not equal the provided user's ID,
	// the email address is not available.
	if existing.ID != user.ID {
		return ErrEmailTaken
	}
	return nil
}

func (uv *userValidator) requireName(user *User) error {
	user.Name = strings.TrimSpace(user.Name)
	if user.Name == "" {
		return ErrNameRequired
	}
	return nil
}

func (uv *userValidator) requirePassword(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordIsComplex(minLength, maxLength int) userValFunc {
	return func(user *User) error {
		var (
			uppercasePresent   bool
			lowercasePresent   bool
			numberPresent      bool
			specialCharPresent bool
			passLen            int
		)

		if strings.TrimSpace(user.Password) == "" {
			return nil
		}

		for _, ch := range user.Password {
			switch {
			case unicode.IsNumber(ch):
				numberPresent = true
				passLen++
			case unicode.IsUpper(ch):
				uppercasePresent = true
				passLen++
			case unicode.IsLower(ch):
				lowercasePresent = true
				passLen++
			case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
				specialCharPresent = true
				passLen++
			case ch == ' ':
				passLen++
			}
		}
		if !(numberPresent && uppercasePresent && lowercasePresent && specialCharPresent && passLen >= minLength && passLen <= maxLength) {
			return ErrPasswordNotComplex
		}
		return nil
	}
}

var _ UserDB = &userGorm{}

// func newUserGorm(connectionInfo string, logging bool) (*userGorm, error) {
// 	db, err := gorm.Open("postgres", connectionInfo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db.LogMode(logging)
// 	return &userGorm{
// 		db: db,
// 	}, nil
// }

type userGorm struct {
	db *gorm.DB
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
// This method expects the rememberToken to be hashed for comparison with stored hashed token.
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

// Create will create the provided user and backfill data
// like the ID, CreatedAt and UpdatedAt fields.
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update will update the persisted user with the provided user instance.
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// UpdateRememberHash will update the remember hash stored on the user.
func (ug *userGorm) UpdateRememberHash(user *User) error {
	return ug.db.Model(user).Update("remember_hash", user.RememberHash).Error
}

// Delete will delete the given user from the db.
func (ug *userGorm) Delete(id uint) error {
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
