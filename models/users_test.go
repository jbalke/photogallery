package models

import (
	"fmt"
	"testing"
	"time"
)

func testingUserService() (UserService, error) {
	const (
		host   = "localhost"
		port   = 5432
		user   = "postgres"
		dbname = "lenslocked_test"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	us, err := NewUserService(psqlInfo, false)
	if err != nil {
		return nil, err
	}

	// Clear users table between tests
	us.DestructiveReset()

	return us, nil
}

func TestCreateUser(t *testing.T) {
	us, err := testingUserService()
	if err != nil {
		t.Fatal(err)
	}

	user := User{
		Name:  "Ted",
		Email: "ted@home.net",
	}

	err = us.Create(&user)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Errorf("Expected ID > 0. Received %d", user.ID)
	}
	if time.Since(user.CreatedAt) > time.Duration(3*time.Second) {
		t.Errorf("Expected CreatedAt to be recent. Received %s", user.CreatedAt)
	}
	if time.Since(user.CreatedAt) > time.Duration(5*time.Second) {
		t.Errorf("Expected UpdatedAt to be recent. Received %s", user.CreatedAt)
	}
}
