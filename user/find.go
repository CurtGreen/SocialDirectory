package user

import (
	"github.com/jinzhu/gorm"
)

// EmailNotInDatabaseError provides specific error information
type EmailNotInDatabaseError struct{}

func (*EmailNotInDatabaseError) Error() string {
	return "Email does not exist in database!"
}

// FindByEmail returns user for given email string
func FindByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	res := db.Find(&user, &User{Email: email})
	if res.RecordNotFound() {
		return nil, &EmailNotInDatabaseError{}
	}
	return &user, nil
}
