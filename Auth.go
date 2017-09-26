package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"

	"github.com/alexedwards/scs/session"
	"github.com/jinzhu/gorm"
)

// Context key type for AuthServe
type contextKey string

// SocialInfo Struct for parsing JSON token response
type SocialInfo struct {
	Provider string
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

var googleOauthConfig = oauth2.Config{
	RedirectURL:  "http://localhost:8000/GoogleCallback",
	ClientID:     "152966871967-159hebcakp2km8ngchbhe6rvd8ikls9j.apps.googleusercontent.com",
	ClientSecret: "4j4bYIw9RdMBJyccAbUcKKB6",
	Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint: google.Endpoint,
}

var facebookOauthConfig = oauth2.Config{
	RedirectURL:  "http://localhost:8000/FacebookCallback",
	ClientID:     "132628397281476",
	ClientSecret: "8aa6aec019015d02ebff3a4485352707",
	Scopes: []string{"email",
		"user_about_me"},
	Endpoint: facebook.Endpoint,
}

// GetNewStateString function to produce random state string for protection against CSRF
func GetNewStateString() string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*"
	byteSlice := make([]byte, 12)
	for i := range byteSlice {
		byteSlice[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(byteSlice)
}

// AuthServe middleware to handle Authentication via Session Cookies
// It attempts to load userID from sessionManager, and serves it's handler with Appropriate
// context
func AuthServe(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser := contextKey("currentUser")
		userID, err := session.GetInt(r, "userID")
		//If user logged in then send authorized status
		if err == nil && userID > 0 {
			ctx := context.WithValue(r.Context(), currentUser, "authorized")
			handler.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// User is not logged in send guest status
			ctx := context.WithValue(r.Context(), currentUser, "guest")
			handler.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

// CSRFServe provides middleware for protecting forms from CSRF attack
// We create a randomized state string for each GET request for a CSRF protected
// handler, storing it in session and passing it along with the context for additional
// processing with the requested content.
// On Post requests we attempt to retrieve this random string from session
// log a Panic if the csrftoken doesn't exist, and pass it in context if it does
func CSRFServe(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			session.PutString(r, "csrftoken", GetNewStateString())
			handler.ServeHTTP(w, r)
		case http.MethodPost:
			csrftoken := contextKey("csrftoken")
			csrfString, err := session.GetString(r, "csrftoken")
			if err != nil {
				log.Panicf("CSRF string not in Session: %v", err.Error())

			}
			ctx := context.WithValue(r.Context(), csrftoken, csrfString)
			handler.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

// LoginHandler Handles requests for logging into Services
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles("Templates/contentlayout.html", "Templates/publicnav.html", "Templates/login.html")
		t.Execute(w, "Login Page!")
	case http.MethodPost:
		r.ParseForm()

		user := User{Emails: []Email{Email{Email: ""}}}
		// Find ID for given email and retrieve User
		user.GetByID(user.Emails[0].UID(r.PostFormValue("email")))
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.PostFormValue("password")))
		if err == nil {
			// Check to see if user has active sessions
			err := session.PutInt(r, "userID", int(user.ID))
			//If we have an error then there is no session in the database, so we should create it
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			http.Redirect(w, r, "/dashboard/my/", 302)
		}
	}
}

// LoginWithGoogle provides login link to Google Oauth2 api
func LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	// Create CSRF protection string
	oauthStateString := GetNewStateString()
	session.PutString(r, "state", oauthStateString)
	// Fire off request for authCode
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

// GoogleCallback provides callback handling from Google Oauth2 api
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Recieve Callback and check CSRF state token
	oauthStateString, err := session.GetString(r, "state")
	if err != nil {
		log.Panic(err.Error())
	}
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Parse AuthCode from respone and exchange for AccessToken
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Code Exchang Failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//Use AccessToken to retrieve user information
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Printf("Error occured reading response %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	// Parse User Info from Response Body
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Final check error!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	// Create SocialInfo struct and read json into it
	info := SocialInfo{Provider: "google"}
	err = json.Unmarshal(contents, &info)
	if err != nil {
		log.Panic(err.Error())
	}
	userID, hasEmail := handleSocial(info, r)
	err = session.PutInt(r, "userID", int(userID))
	if err != nil {
		log.Panic(err.Error())
	}
	if hasEmail == false {
		http.Redirect(w, r, "/dashboard/profile/", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/dashboard/my/", 302)
	}
}

// LoginWithFacebook provides login link to Facebook Oauth2 api
func LoginWithFacebook(w http.ResponseWriter, r *http.Request) {
	// Create CSRF protection string
	oauthStateString := GetNewStateString()
	session.PutString(r, "state", oauthStateString)
	// Fire off request for authCode
	url := facebookOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

// FacebookCallback provides callback handling from Facebook Oauth2 api
func FacebookCallback(w http.ResponseWriter, r *http.Request) {
	// Recieve Callback and check CSRF state token
	oauthStateString, err := session.GetString(r, "state")
	if err != nil {
		log.Panic(err.Error())
	}
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Parse AuthCode from respone and exchange for AccessToken
	code := r.FormValue("code")
	token, err := facebookOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Code Exchang Failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//Use AccessToken to retrieve user information
	response, err := http.Get("https://graph.facebook.com/me?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Printf("Error occured reading response %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	// Parse User Info from Response Body
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Final check error!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	// Create SocialInfo struct and read json into it
	info := SocialInfo{Provider: "facebook"}
	err = json.Unmarshal(contents, &info)
	if err != nil {
		log.Panic(err.Error())
	}
	userID, hasEmail := handleSocial(info, r)
	err = session.PutInt(r, "userID", int(userID))
	if err != nil {
		log.Panic(err.Error())
	}
	if hasEmail == false {
		http.Redirect(w, r, "/dashboard/profile/", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/dashboard/my/", 302)
	}

}

// Function for logging in SocialUsers, and/or creating them returns hasEmail status and UserID.
func handleSocial(info SocialInfo, r *http.Request) (int, bool) {
	// Open database connection
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	defer Db.Close()
	if err != nil {
		log.Panic(err.Error())
	}

	// Find out if user has existing account, and return userID
	result := SocialUser{}
	Db.Table("social_users").Select("user_id").Where("provider = $1 AND provider_user_id = $2", info.Provider, info.ID).Scan(&result)
	if result.UserID != 0 {
		return result.UserID, true
	}

	// Since User has no account create one
	user := User{Name: info.Name,
		Emails:         []Email{Email{Email: info.Email}},
		SocialAccounts: []SocialUser{SocialUser{Provider: info.Provider, ProviderUserID: info.ID}}}
	// If user is logged in, associate social account
	userID, _ := session.GetInt(r, "userID")
	if userID != 0 {
		user.ID = uint(userID)
		user.Save()
		if user.Emails[0].Email == "" {
			return int(user.ID), false
		}
		return int(user.ID), true
	}
	user.Create()
	if user.Emails[0].Email == "" {
		return int(user.ID), false
	}
	return int(user.ID), true

}

// LogoutHandler answers requests to logout, by destroying user's authToken
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session.Destroy(w, r)
	http.Redirect(w, r, "/", 303)

}
