package login

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/CurtGreen/SocialDirectory/user"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// TestSignup tests package signup
func TestSignup(t *testing.T) {
	db, err := gorm.Open("postgres", "user=Meliphas password=them47r1x dbname=test sslmode=disable")
	if err != nil {
		t.Error("No database connection")
	}
	defer db.Close()

	if db.HasTable(user.User{}) == true {
		db.DropTable(user.User{})
	}
	db.CreateTable(user.User{})

	var currentUser user.User
	var hashedPass []byte
	password := "test"

	hashedPass, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Error("Test Data population bcrypting failed")
	}
	currentUser = user.User{Email: "test@test.com", PasswordHash: string(hashedPass)}
	_, err = user.Create(db, &currentUser)
	if err != nil {
		t.Error("Test Data population user create failed")
	}

	req := &Request{Email: currentUser.Email, Password: password}
	_, err = Login(db, req)
	if err != nil {
		t.Error("Login matching pass operation failed!")
	}

	req.Password = "wrong"
	_, err = Login(db, req)
	if err == nil {
		t.Error("Somehow logged in with wrong pass")
	}

	req.Email = "illusory@reality.net"
	_, err = Login(db, req)
	if err == nil {
		t.Error("Logged in without registered email")
	}

}
