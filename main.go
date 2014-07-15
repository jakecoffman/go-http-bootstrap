package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	http.Handle("/static/", static)
	http.HandleFunc("/", filter(handle))

	host := "localhost:5001"
	log.Println("Serving on " + host)
	log.Fatal(http.ListenAndServe(host, nil))
}

var static http.Handler = http.StripPrefix("/static", http.FileServer(http.Dir("static")))

// a place to put filters/middleware
func filter(f func(http.ResponseWriter, *http.Request) int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := f(w, r)
		log.Println(r.RemoteAddr, r.URL, status)
	}
}

// handle responses by rendering a template of the same name as the path
func handle(w http.ResponseWriter, r *http.Request) int {
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
		return 404
	}
	t.Execute(w, nil)
	log.Println(r.RemoteAddr, r.URL, 200)
	return 200
}
