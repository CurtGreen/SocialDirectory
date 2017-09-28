package user

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// User type defines model for database records that represent our user's properties
type User struct {
	gorm.Model
	Email        string `sql:"unique"`
	PasswordHash string
}

// UniqueConstraintEmail constant for meaningful error checking
const (
	UniqueConstraintEmail = "user_email_key"
)

// EmailDuplicateError matachers Error type interface to provide meaningful information
type EmailDuplicateError struct {
	Email string
}

func (e *EmailDuplicateError) Error() string {
	return fmt.Sprintf("Email '%s' already exists", e.Email)
}
