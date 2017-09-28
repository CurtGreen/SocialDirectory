package signup

import (
	"testing"

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

	req := &Request{Email: "test@email.com", Password: "test"}

	_, err = Signup(db, req)
	if err != nil {
		t.Error("Signup failed!")
	}

	_, err = Signup(db, req)
	if err == nil {
		t.Error("Unique constraint failed")
	}

}
