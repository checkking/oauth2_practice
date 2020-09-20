package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var myEndPoint = oauth2.Endpoint{
	AuthURL:   "http://localhost:8001/oauth2/auth",
	TokenURL:  "http://localhost:8001/oauth2/token",
	AuthStyle: oauth2.AuthStyleInParams,
}


func New() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/",  http.FileServer(http.Dir("templates/")))

	// OauthGoogle
	mux.HandleFunc("/oauth2/login", oauthLogin)
	mux.HandleFunc("/oauth2/callback", oauthCallback)

	return mux
}

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var oauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/oauth2/callback",
	ClientID:     "test_appid",
	ClientSecret: "test_secret",
	Scopes:       []string{"name", "email"},
	Endpoint:     myEndPoint,
}

const oauthUrlAPI = "http://localhost:8001/user?access_token="

func oauthLogin(w http.ResponseWriter, r *http.Request) {

	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	u := myEndPoint.AuthURL + "?client_id=" + oauthConfig.ClientID + "&response_type=code&state=" + oauthState + "&redirect_url=" + url.QueryEscape(oauthConfig.RedirectURL) +
		"&state=" + oauthState
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func oauthCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth  state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromEndpoint(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// GetOrCreate User in your db.
	// Redirect or response with a token.
	// More code .....
	fmt.Fprintf(w, "UserInfo: %s\n", data)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func getUserDataFromEndpoint(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	glog.Infof("get token succ:%s", token)
	response, err := http.Get(oauthUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}
