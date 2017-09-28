package signup

import (
	"github.com/CurtGreen/SocialDirectory/user"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// Request holds signup values
type Request struct {
	Email    string
	Password string
}

// Response returns user id
type Response struct {
	ID uint
}

// Signup handler for User model
func Signup(db *gorm.DB, req *Request) (*Response, error) {
	bcryptedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:        req.Email,
		PasswordHash: string(bcryptedPass),
	}

	id, err := user.Create(db, newUser)
	if err != nil {
		return nil, err
	}
	return &Response{ID: id}, err
}
