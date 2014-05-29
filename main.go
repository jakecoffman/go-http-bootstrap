package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
)

// An example struct used in templates below
type Example struct {
	Title string
	Text  string
}

func main() {
	// handles static files in subdirectory "static"
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	// example of how to use templates
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("templates/example.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, Example{"Hello", "World"})
	})

	// example of returning text some other way
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	// log if listen and serve fails
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
