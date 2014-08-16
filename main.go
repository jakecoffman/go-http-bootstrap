package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "db.sqlite")
	check(err)
	defer db.Close()

	initDb(db)

	http.Handle("/static/", static)
	http.HandleFunc("/", filter(handle))

	host := "localhost:5001"
	log.Println("Serving on " + host)
	log.Fatal(http.ListenAndServe(host, nil))
}

// idempotent bootstrap
func initDb(db *sql.DB) {
	s := `create table if not exists users (id integer not null primary key, name text, password text);`
	_, err := db.Exec(s)
	check(err)

	rows, err := db.Query("select id, name from users where name='admin'")
	check(err)
	defer rows.Close()
	if !rows.Next() {
		_, err = db.Exec("insert into users(name, password) values('admin', 'admin')")
		check(err)
	}
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}
