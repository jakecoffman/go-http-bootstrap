package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	http.Handle("/static/", static)
	http.HandleFunc("/", handle)

	log.Fatal(http.ListenAndServe("localhost:5001", nil))
}

var static http.Handler = http.StripPrefix("/static", http.FileServer(http.Dir("static")))

func handle(w http.ResponseWriter, r *http.Request) {
	var name string

	if r.URL.String() == "/" {
		name = "templates/index.html"
	} else {
		name = "templates/" + r.URL.String() + ".html"
	}

	t, err := template.ParseFiles(
		"templates/base.html",
		name,
	)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(r.RemoteAddr, r.URL, 404)
		return
	}
	t.Execute(w, nil)
	log.Println(r.RemoteAddr, r.URL, 200)
}
