package login

import (
	"github.com/CurtGreen/SocialDirectory/user"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// Request values for Login
type Request struct {
	Email    string
	Password string
}

// Response values for Login
type Response struct {
	User *user.User
}

// PasswordMismatchError for detailed error information
type PasswordMismatchError struct{}

func (e *PasswordMismatchError) Error() string {
	return "Password does not match!"
}

// Login returns user with password match
func Login(db *gorm.DB, req *Request) (*Response, error) {
	user, err := user.FindByEmail(db, req.Email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, &PasswordMismatchError{}
	}

	return &Response{User: user}, nil
}
