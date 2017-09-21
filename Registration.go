package main

import (
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// RegisterHandler Handle requests for creating new accounts
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/publicnav.html", "Templates/register.html")
		t.Execute(w, "Registration Page!")
	case "POST":
		//Establish Database Connection
		Db, err := gorm.Open(DATABASE, CREDENTIALS)
		defer Db.Close()
		if err != nil {
			log.Panic(err.Error())
		}

		// Make a query to determine if records already exist for this Email Address
		r.ParseForm()
		user := User{}
		email := r.PostFormValue("email")
		var count int64
		Db.Model(&Email{}).Where("email = ?", email).Count(&count)
		// If there are no records that contain the email provided create a new user
		if count == 0 {

			user.Name = r.PostFormValue("name")
			user.Emails = []Email{Email{Email: email}}
			bcryptedPass, err := bcrypt.GenerateFromPassword([]byte(r.PostFormValue("password")), 10)
			if err != nil {
				log.Panic(err.Error())
			}
			user.Password = string(bcryptedPass)
			user.Create()
			LoginHandler(w, r)
		} else {
			LoginHandler(w, r)
		}
	}
}
