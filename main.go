package main

import (
	"log"
	"net/http"
	"time"

	"github.com/CurtGreen/SocialDirectory/controllers"
	"github.com/alexedwards/scs"
	"github.com/alexedwards/scs/stores/memstore"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var sessionManager = scs.NewManager(memstore.New(time.Hour * 24))

func main() {
	// Configure ServMux
	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/", controllers.Welcome)

	log.Fatal(http.ListenAndServe("localhost:5000", serveMux))
}
