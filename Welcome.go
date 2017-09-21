package main

import (
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/session"
)

// WelcomeHandler Handles requests for welcome Page
func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := session.GetInt(r, "userID")
	if userID > 0 {
		t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/welcome.html")
		t.Execute(w, r)
	} else {
		t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/publicnav.html", "Templates/welcome.html")
		t.Execute(w, r)
	}

}
