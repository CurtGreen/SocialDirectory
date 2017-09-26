package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/session"
	"github.com/jinzhu/gorm"
)

// DashboardHandler Handles requests for dashboard
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value(contextKey("currentUser"))
	fmt.Printf("Current User by Context is: %v\n", currentUser)
	if currentUser == "guest" {
		http.Error(w, "Attempted to access restricted content without logging in a user", http.StatusUnauthorized)
	} else {
		t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/dashboardcontent.html")
		t.Execute(w, fmt.Sprintf("Dashboard! %v", currentUser))
	}

}

// ProfileHandler handles requests to edit and publish user profile information
func ProfileHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := session.GetInt(r, "userID")
	if err != nil {
		log.Panic(err.Error())
	}

	user := User{}
	user.GetByID(uint(userID))

	switch r.Method {

	case http.MethodGet:
		t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/dashprofilecontent.html")
		t.Execute(w, user)

	case http.MethodPost:
		r.ParseForm()
		birthday := r.FormValue("dob")
		user.Name = r.FormValue("name")
		user.Emails[0].Email = r.FormValue("email")
		user.About.DOB, err = time.Parse("2 January, 2006", birthday)
		if err != nil {
			log.Panic(err.Error())
		}
		user.Save()
		http.Redirect(w, r, "/dashboard/profile/", 302)
	}
}

// OrganizationHandler manages requests to update User's Organizations
func OrganizationHandler(w http.ResponseWriter, r *http.Request) {

	CurrentUser := User{}
	Org := Organization{}

	userID, _ := session.GetInt(r, "userID")
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	if err != nil {
		log.Panic(err)
	}
	defer Db.Close()
	switch r.Method {
	case http.MethodGet:
		CurrentUser.GetByID(uint(userID))
		// Db.Preload("Location").Where("user_id = ?", uint(userID)).Find(&Organizations)
		if len(CurrentUser.Organizations) > 0 {
			t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/dashboardorgcontent.html")
			t.Execute(w, CurrentUser)
		} else {
			t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/dashboardorgcontent.html")
			t.Execute(w, CurrentUser)
		}

	case http.MethodPost:
		r.ParseForm()
		if r.FormValue("id") != "" {

			Db.Preload("Location").Where("user_id = ?", uint(userID)).Where("id = ?", r.FormValue("id")).First(&Org)
		}

		if r.FormValue("location") != "" {
			jsonAddress := GeocodeAddress(r)
			Org.Location.Street = jsonAddress.Res[0].Address
			Org.Location.Lat = jsonAddress.Res[0].Geom.Loc.Lat
			Org.Location.Lng = jsonAddress.Res[0].Geom.Loc.Lng
		}
		Org.Location.Ext = r.FormValue("ext")
		Org.Name = r.FormValue("name")
		if Org.Model.ID > 0 {
			fmt.Printf("Executing Model Save %v\n", Org)
			Db.Save(&Org)
			DB, err := gorm.Open(DATABASE, CREDENTIALS)
			if err != nil {
				log.Panic(err.Error())
			}
			defer DB.Close()
			DB.Exec("UPDATE addresses SET geom = ST_SetSRID(ST_MakePoint(lng, lat), 4326)")
			http.Redirect(w, r, "/dashboard/organizations/", 302)
		} else {
			Org.UserID = userID
			fmt.Printf("Creating new Record %v", Org)
			Db.Create(&Org)
			DB, err := gorm.Open(DATABASE, CREDENTIALS)
			if err != nil {
				log.Panic(err.Error())
			}
			defer DB.Close()
			DB.Exec("UPDATE addresses SET geom = ST_SetSRID(ST_MakePoint(lng, lat), 4326)")
			http.Redirect(w, r, "/dashboard/organizations/", 302)
		}
	}
}
