package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// first will query using the provided gorm.DB and will
// get the first item returned and place in the provided dst.
// If nothing is found it will return ErrNotFound.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error

	fmt.Println("first err = ", err)

	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}
