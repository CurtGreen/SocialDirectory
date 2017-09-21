package main

import (
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/engine/memstore"
	"github.com/alexedwards/scs/session"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Init opens a connection to Database and Migrates Models
func Init() {
	Db, err := gorm.Open("postgres", "user=Meliphas password=them47r1x dbname=Meliphas sslmode=disable")
	defer Db.Close()
	if err != nil {
		log.Panic(err)
	}
	Db.AutoMigrate(&User{}, &Profile{}, &Email{}, &Address{}, &Speciality{}, &Organization{}, &Membership{}, &Event{}, &Message{}, &Comment{}, &Contact{}, &SocialUser{})
	Db.Raw("SELECT AddGeometryColumn('addresses', 'geom', 4326, 'POINT', 2);	CREATE INDEX idx_organization_addresses ON addresses USING gist(geom);")
	return
}

// DATABASE is global environment string for type of Database
var DATABASE = "postgres"

// CREDENTIALS is global environment string for connectin to Database
var CREDENTIALS = "user=Meliphas password=them47r1x dbname=Meliphas sslmode=disable"

// GOOGLEKEY is global environment string containing Google API key
var GOOGLEKEY = "&key=AIzaSyBkzm70gs5HDFmj63c9GXAoPuNDYd-mElg"

// GEOCODEURL api url for Google Geocode requests
var GEOCODEURL = "https://maps.googleapis.com/maps/api/geocode/json?"

func main() {
	Init()
	// Session storage engine Initialization
	engine := memstore.New(0)
	sessionManager := session.Manage(engine)
	multiplex := mux.NewRouter()
	multiplex.StrictSlash(true)
	// Setup path to Public folder to serve files from file system
	multiplex.PathPrefix("/Public/").Handler(http.StripPrefix("/Public/", http.FileServer(http.Dir("Public"))))
	multiplex.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./Public/favicon.ico") })
	// Set Standard Routes
	multiplex.HandleFunc("/", WelcomeHandler)
	multiplex.HandleFunc("/directory/", DirectoryHandler)
	multiplex.HandleFunc("/dashboard/my/", AuthServe(DashboardHandler))
	multiplex.HandleFunc("/dashboard/profile/", AuthServe(ProfileHandler))
	multiplex.HandleFunc("/dashboard/organizations/", AuthServe(OrganizationHandler))
	multiplex.HandleFunc("/register/", RegisterHandler)
	multiplex.HandleFunc("/login/", LoginHandler)
	multiplex.HandleFunc("/googlelogin/", LoginWithGoogle)
	multiplex.HandleFunc("/GoogleCallback/", GoogleCallback)
	multiplex.HandleFunc("/facebooklogin/", LoginWithFacebook)
	multiplex.HandleFunc("/FacebookCallback/", FacebookCallback)
	multiplex.HandleFunc("/logout/", AuthServe(LogoutHandler))

	http.ListenAndServe("127.0.0.1:8000", handlers.LoggingHandler(os.Stdout, sessionManager(multiplex)))
}
