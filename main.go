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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	// example of custom middleware/filter
	http.Handle("/baz", filter(func(w http.ResponseWriter, r *http.Request) {
		log.Println("In baz handler")
		w.Write([]byte("Hello!"))
	}))

	// anoter custom middleware filter, breaking from the typical handler interface
	http.Handle("/bap", myFilter(func(w http.ResponseWriter, r *http.Request) string {
		log.Printf("In bap handler")
		return "Greetings!"
	}))

	// log if listen and serve fails
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// create a new type, which is just a normal function handler
type filter func(w http.ResponseWriter, r *http.Request)

// attach ServeHTTP which implements http.Handler
func (f filter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// perform actions before and after calling the handler
	log.Printf("Started request for %s", r.URL)

	f(w, r)

	log.Printf("Finished request for %s", r.URL)
}

// example of custom middleware/filters via function literals and closures
func myFilter(f func(http.ResponseWriter, *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Started request for %s", r.URL)

		// you can create custom handlers that don't match the usual ResponseWriter/Request format
		toWrite := f(w, r)
		w.Write([]byte(toWrite))

		log.Printf("Finished request for %s", r.URL)
	}
}
