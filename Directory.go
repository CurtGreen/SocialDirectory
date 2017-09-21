package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/alexedwards/scs/session"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Result holds Address string and Geometry struct
type Result struct {
	Address string   `json:"formatted_address"`
	Geom    Geometry `json:"geometry"`
}

// Location holds lat and lng values
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Geometry struct holds a Location struct
type Geometry struct {
	Loc Location `json:"location"`
}

// Geocode holds a slice of Results structs
type Geocode struct {
	Res []Result `json:"results"`
}

// DistanceStruct contains distance search result information for a record
type DistanceStruct struct {
	OrganizationID int
	Name           string
	Street         string
	Lat            float64
	Lng            float64
	Distance       float64
}

// DirectoryHandler Handles requests for directory
func DirectoryHandler(w http.ResponseWriter, r *http.Request) {
	// Serve Appropriate Nav
	userID, _ := session.GetInt(r, "userID")

	switch r.Method {
	case http.MethodGet:
		if userID > 0 {
			t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/directorycontent.html", "Templates/directoryget.html")
			t.Execute(w, r)
		} else {
			t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/publicnav.html", "Templates/directorycontent.html", "Templates/directoryget.html")
			t.Execute(w, r)
		}

	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			log.Panic(err)
		}

		requestData := GeocodeAddress(r)
		type Data struct {
			Lat float64
			Lng float64
		}

		type OrgResults struct {
			Location Data
			Results  []DistanceStruct
		}
		searchData := Data{Lat: requestData.Res[0].Geom.Loc.Lat, Lng: requestData.Res[0].Geom.Loc.Lng}
		templateData := OrgResults{Location: searchData, Results: DistanceSearch(searchData.Lat, searchData.Lng)}
		if userID > 0 {
			t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/privatenav.html", "Templates/directorycontent.html", "Templates/directorypost.html")
			t.Execute(w, templateData)
		} else {
			t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/publicnav.html", "Templates/directorycontent.html", "Templates/directorypost.html")
			t.Execute(w, templateData)
		}

	}

}

// DistanceSearch function to retrieve database records from a given distance
func DistanceSearch(lat float64, lng float64) (selected []DistanceStruct) {
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	if err != nil {
		log.Panic(err)
	}
	defer Db.Close()

	// Db.Raw("SELECT organization_id, street, lat, lng, 2 * 3961 * asin(sqrt((sin(radians((lat - $1) / 2))) ^ 2 + cos(radians(lat)) * cos(radians($1)) * (sin(radians(($2 - lng) / 2))) ^ 2)) as distance FROM addresses GROUP BY organization_id, street, lat, lng HAVING 2 * 3961 * asin(sqrt((sin(radians((lat - $1) / 2))) ^ 2 + cos(radians(lat)) * cos(radians($1)) * (sin(radians(($2 - lng) / 2))) ^ 2)) < 50 ORDER BY distance", lat, lng).Scan(&selected)
	Db.Raw("SELECT organization_id, street, lat, lng, ROUND(CAST(ST_Distance(geom, ST_MakePoint($1, $2)::geography) / 1609.34 AS numeric), 2) AS distance FROM addresses WHERE ST_DWithin(geom, ST_MakePoint($1, $2)::geography, 50000) ORDER BY distance ASC", lng, lat).Scan(&selected)
	for i := range selected {
		Db.Table("organizations").Select("name").Where("id = ?", selected[i].OrganizationID).Scan(&selected[i])
	}
	fmt.Printf("%v \n", selected)
	return
}

// GeocodeAddress Returns Geocoded address from form Request
func GeocodeAddress(r *http.Request) (requestData Geocode) {
	location := "address=" + r.FormValue("location")
	urlRequest := GEOCODEURL + url.PathEscape(location) + GOOGLEKEY
	response, err := http.Get(urlRequest)
	if err != nil {
		log.Panic(err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		log.Panic(err.Error())
	}
	fmt.Printf("Address: %v, Lat: %v, Lng: %v\n", requestData.Res[0].Address, requestData.Res[0].Geom.Loc.Lat, requestData.Res[0].Geom.Loc.Lng)
	return
}
