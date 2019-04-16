package models

import (
	"fmt"
	"testing"
	"time"
)

func testingUserService() (*Services, error) {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "sParhwk72"
		dbname   = "lenslocked_test"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	services, err := NewServices(psqlInfo, true)
	if err != nil {
		return nil, err
	}
	// Clear users table between tests
	services.DestructiveReset()

	return services, nil
}

func TestCreateUser(t *testing.T) {
	services, err := testingUserService()
	if err != nil {
		t.Fatal(err)
	}

	defer services.db.Close()

	user := User{
		Name:     "Ted",
		Email:    "ted@home.net",
		Password: "Pas5word!",
	}

	err = services.User.Create(&user)
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
