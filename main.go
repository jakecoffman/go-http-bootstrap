package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
)

func main() {
	// here we set up our routes. see blow main for the interesting parts.
	// read the documentation about http.ServeMux as these routes are slightly
	// different than what you may be used to.
	http.Handle("/static/", static)
	http.HandleFunc("/", templates)
	http.HandleFunc("/bar/", fprintf)
	http.Handle("/baz", withFilter)
	http.Handle("/bap", simpler)

	// log if listen and serve fails
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// static serves static files from the static directory in cwd
var static http.Handler = http.StripPrefix("/static", http.FileServer(http.Dir("static")))

// an example struct used in template below
type Example struct {
	Title string
	Text  string
}

// this is go's server-side templating. you can parse the template(s) once at startup
// for performance reasons, or on demand for development reasons like this does.
func templates(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/example.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, Example{"Hello", "World"})
}

// a simple example of writing a response, including the current path in the response
func fprintf(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

// now for a custom middleware/filter via implementing http.Handler. first we need a type
// to attach a method to. this can be anything, but handy if it's a function.
type filter func(r *http.Request) string

// this is the method that implements http.Handler, it will be called rather than the func
func (f filter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// perform actions before calling the handler
	log.Printf("Started request for %s", r.URL)

	// call the function that got passed in. in this example it returns a string and we
	// write it to the response. in this way we can factor out duplicate code from the
	// handlers.
	str := f(r)
	w.Write([]byte(str))

	// perform actions after calling the handler
	log.Printf("Finished request for %s", r.URL)
}

// now wrap a function with the filter
var withFilter http.Handler = filter(func(r *http.Request) string {
	log.Println("In baz handler")
	return "Hello!"
})

// here's a simpler custom middleware/filters via function literals and closures.
// rather than jumping through the hoops of implementing an interface, we just define
// a function that takes another function and returns a handlerfunc.
func myFilter(f func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Started request for %s", r.URL)

		toWrite := f(r)
		w.Write([]byte(toWrite))

		log.Printf("Finished request for %s", r.URL)
	}
}

// so this is actually the same example as the above, only simpler
var simpler http.Handler = myFilter(func(r *http.Request) string {
	log.Printf("In bap handler")
	return "Greetings!"
})
