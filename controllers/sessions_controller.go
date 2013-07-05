package controllers

import (
	"appengine"
	"appengine/urlfetch"
	"net/http"
	"fmt"
	"code.google.com/p/goauth2/oauth"
	
	"helpers"
	"models"
)

const (
        // Created at http://code.google.com/apis/console, these identify
        // our app for the OAuth protocol.
        CLIENT_ID     = "36368233290.apps.googleusercontent.com"
        CLIENT_SECRET = "ZSU7Soyt3N2ipRddpghScXlx"
)

// Set up a configuration.
func config(host string) *oauth.Config{
		return &oauth.Config{
			ClientId:     CLIENT_ID,
			ClientSecret: CLIENT_SECRET,
			Scope:        "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email",
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://accounts.google.com/o/oauth2/token",
			RedirectURL:  fmt.Sprintf("http://%s/oauth2callback", host),
		}
}

func Auth(w http.ResponseWriter, r *http.Request){
	url := config(r.Host).AuthCodeURL(r.URL.RawQuery)
    http.Redirect(w, r, url, http.StatusFound)
}

func AuthCallback(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	// Exchange code for an access token at OAuth provider.
	code := r.FormValue("code")
	t := &oauth.Transport{
			Config: config(r.Host),
			Transport: &urlfetch.Transport{
					Context: appengine.NewContext(r),
			},
	}
	
	if _, err := t.Exchange(code); err != nil {
		c.Debugf("Exchange: %q", err)
	}
	user, _ := models.FetchUserInfo(r, t.Client())
	
	if helpers.UserAuthorized(user) {
		models.CurrentUser = user
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func Logout(w http.ResponseWriter, r *http.Request){
	helpers.Logout()
	
	http.Redirect(w, r, "/", http.StatusFound)
}