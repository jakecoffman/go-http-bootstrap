package main

import (
	"html/template"
	"net/http"

	"code.google.com/p/goauth2/oauth"
)

var notAuthenticatedTemplate = template.Must(template.New("").Parse(`
<html><body>
You have currently not given permissions to access your data. Please authenticate this app with the Google OAuth provider.
<form action="/authorize" method="POST"><input type="submit" value="Ok, authorize this app with my id"/></form>
</body></html>
`))

var userInfoTemplate = template.Must(template.New("").Parse(`
<html><body>
This app is now authenticated to access your Google user info.  Your details are:<br />
{{.}}
</body></html>
`))

// variables used during oauth protocol flow of authentication
var (
	code  = ""
	token = ""
)

var oauthCfg = &oauth.Config{
	//TODO: put your project's Client Id here.  To be got from https://code.google.com/apis/console
	ClientId: "",

	//TODO: put your project's Client Secret value here https://code.google.com/apis/console
	ClientSecret: "",

	//For Google's oauth2 authentication, use this defined URL
	AuthURL: "https://accounts.google.com/o/oauth2/auth",

	//For Google's oauth2 authentication, use this defined URL
	TokenURL: "https://accounts.google.com/o/oauth2/token",

	//To return your oauth2 code, Google will redirect the browser to this page that you have defined
	//TODO: This exact URL should also be added in your Google API console for this project within "API Access"->"Redirect URIs"
	RedirectURL: "http://localhost:8080/oauth2callback",

	//This is the 'scope' of the data that you are asking the user's permission to access. For getting user's info, this is the url that Google has defined.
	Scope: "https://www.googleapis.com/auth/userinfo.profile",
}

//This is the URL that Google has defined so that an authenticated application may obtain the user's info in json format
const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/authorize", handleAuthorize)

	//Google will redirect to this page to return your code, so handle it appropriately
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)

	http.ListenAndServe("localhost:8080", nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	notAuthenticatedTemplate.Execute(w, nil)
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

	t := &oauth.Transport{Config: oauthCfg}

	// Exchange the received code for a token
	t.Exchange(code)

	//now get user data based on the Transport which has the token
	resp, _ := t.Client().Get(profileInfoURL)

	buf := make([]byte, 1024)
	resp.Body.Read(buf)
	userInfoTemplate.Execute(w, string(buf))
}
