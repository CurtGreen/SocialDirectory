package controllers

import (
	"html/template"
	"net/http"
)

type WelcomePage struct {
	Data string
}

func (p *WelcomePage) Validate(r *http.Request) (bool, error) {
	// Do Validation Work
	return true, nil
}

func (p *WelcomePage) Error(err error, w http.ResponseWriter) {
	t, err1 := template.New("errorPage").Parse(`<html><div><h1>ERROR! : {{.}}</h1></div></html>`)
	if err1 != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, err)
	return
}

func (p *WelcomePage) Serve(w http.ResponseWriter, r *http.Request) error {
	t, err := template.New("welcomePage").Parse(`<html><div><h1>{{.Data}}</h1></div></html>`)
	if err != nil {
		return err
	}
	err = t.Execute(w, p)
	if err != nil {
		return err
	}
	return nil
}

// Welcome handles WelcomePage Processing
func Welcome(w http.ResponseWriter, r *http.Request) {
	page := WelcomePage{Data: "Welcome to the Site!"}
	state, err := page.Validate(r)
	if state != true {
		page.Error(err, w)
		return
	}
	page.Serve(w, r)
	return
}
