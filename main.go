package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"code.google.com/p/goauth2/oauth"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// variables used during oauth protocol flow of authentication
var (
	code  = ""
	token = ""
)

var oauthCfg oauth.Config

//This is the URL that Google has defined so that an authenticated application may obtain the user's info in json format
const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func init() {
	file, _ := os.Open("conf.json")
	conf := map[string]string{}
	err := json.NewDecoder(file).Decode(&conf)
	check(err)
	oauthCfg = oauth.Config{
		//TODO: put your project's Client Id here.  To be got from https://code.google.com/apis/console
		ClientId: conf["client"],
		//TODO: put your project's Client Secret value here https://code.google.com/apis/console
		ClientSecret: conf["secret"],
		//For Google's oauth2 authentication, use this defined URL
		AuthURL: "https://accounts.google.com/o/oauth2/auth",
		//For Google's oauth2 authentication, use this defined URL
		TokenURL: "https://accounts.google.com/o/oauth2/token",
		//To return your oauth2 code, Google will redirect the browser to this page that you have defined
		//TODO: This exact URL should also be added in your Google API console for this project within "API Access"->"Redirect URIs"
		RedirectURL: "http://localhost/oauth2callback",
		//This is the 'scope' of the data that you are asking the user's permission to access. For getting user's info, this is the url that Google has defined.
		Scope: "https://www.googleapis.com/auth/userinfo.profile",
	}
}

func main() {
	db, err := sql.Open("sqlite3", "db.sqlite")
	check(err)
	defer db.Close()

	initDb(db)

	http.Handle("/static/", static)
	http.HandleFunc("/", filter(handle))
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)
	http.HandleFunc("/logout", logoutHandler)

	host := "localhost:80"
	log.Println("Serving on " + host)
	log.Fatal(http.ListenAndServe(host, context.ClearHandler(http.DefaultServeMux)))
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

	session, err := store.Get(r, "loginSession")
	if err != nil {
		log.Println("Unable to get session", err)
		return 500
	}
	if val, ok := session.Values["user"].(string); ok {
		t.Execute(w, map[string]interface{}{"user": val})
	} else {
		log.Println("No user info found in session")
		t.Execute(w, nil)
	}
	log.Println(r.RemoteAddr, r.URL, 200)
	return 200
}

// Start the authorization process
func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

// Function that handles the callback from the Google server
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{Config: &oauthCfg}
	// Exchange the received code for a token
	t.Exchange(code)
	//now get user data based on the Transport which has the token
	resp, err := t.Client().Get(profileInfoURL)
	if err != nil {
		log.Println("Unable to get client info", err)
		return
	}

	userData := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&userData)
	if err != nil {
		log.Println("Unable to decode JSON", err)
		return
	}

	session, err := store.Get(r, "loginSession")
	if err != nil {
		log.Println("Unable to get session", err)
		return
	}
	session.Values["user"] = userData["name"]
	err = session.Save(r, w)
	if err != nil {
		log.Println("Unable to save session", err)
		return
	}
	log.Println("Session values:", session.Values)
	http.Redirect(w, r, "/", 302)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "loginSession")
	if err != nil {
		log.Println("Unable to get session", err)
		return
	}
	delete(session.Values, "user")
	session.Save(r, w)
	http.Redirect(w, r, "/", 302)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
