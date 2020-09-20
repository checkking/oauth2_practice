package handlers

import (
	"encoding/json"
	"github.com/hashicorp/go-uuid"
	"github.com/golang/glog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

)

var (
	codeMap map[string]string
	appMap map[string]*AppRegistry
	tokenMap map[string]string
)

type AppRegistry struct {
	AppId       string
	AppSecret   string
	RedirectUrl string
	Scope       string
}

func init() {
	codeMap = make(map[string]string, 0)
	appMap = make(map[string]*AppRegistry, 0)
	appMap["test_appid"] = &AppRegistry{
		AppId:       "test_appid",
		AppSecret:   "test_secret",
		RedirectUrl: "http://localhost:8000/oauth2/callback",
		Scope:       "name,email",
	}
	tokenMap = make(map[string]string, 0)
}

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/oauth2/auth", oauthCode)
	mux.HandleFunc("/oauth2/token", oauthToken)

	mux.HandleFunc("/user", getUserInfo)

	return mux
}

func generateCode(appId, userId string) string {
	code := rand.Intn(999999999)
	codeStr := strconv.Itoa(code)
	codeMap[codeStr] = appId + "|" + userId + "|" + strconv.Itoa(int(time.Now().Unix()))
	return codeStr
}

func oauthCode(w http.ResponseWriter, r *http.Request) {
	redirectUrl := r.FormValue("redirect_url")
	redirectUrl,_ = url.QueryUnescape(redirectUrl)
	clientId := r.FormValue("client_id")
	state := r.FormValue("state")
	responseType := r.FormValue("response_type")
	registry, ok := appMap[clientId]
	if !ok {
		glog.Errorf("clinet_id not registered:%s", clientId)
		http.Redirect(w, r, "/",  http.StatusTemporaryRedirect)
	}
	if registry.RedirectUrl != redirectUrl {
		glog.Errorf("redirect_url ivalid:%s", redirectUrl)
		http.Redirect(w, r, "/",  http.StatusTemporaryRedirect)
	}
	code := ""
	if responseType == "code" {
		code = generateCode(clientId, "user_id")
	}
	// 构造跳转url
	resUrl := redirectUrl + "?code=" + code + "&state=" + state
	http.Redirect(w, r, resUrl,  http.StatusTemporaryRedirect)
	return
}

func oauthToken(w http.ResponseWriter, r *http.Request) {
	redirectUrl := r.FormValue("redirect_uri")
	redirectUrl,_ = url.QueryUnescape(redirectUrl)
	clientId := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	code := r.FormValue("code")
	grantType := r.FormValue("grant_type")
	if grantType != "authorization_code" {
		glog.Errorf("invalid type")
		w.Write([]byte("failed"))
		return
	}
	registry, ok := appMap[clientId]
	if !ok {
		glog.Errorf("clinet_id not registered:%s", clientId)
		w.Write([]byte("failed"))
		return
	}
	if registry.RedirectUrl != redirectUrl {
		glog.Errorf("redirect_url ivalid:%s", redirectUrl)
		w.Write([]byte("failed"))
		return
	}
	if registry.AppSecret != clientSecret {
		glog.Errorf("invalid secret")
		w.Write([]byte("failed"))
		return
	}
	if _, ok = codeMap[code]; !ok {
		glog.Errorf("invalid code")
		w.Write([]byte("failed"))
		return
	}
	// 删除code缓存
	delete(codeMap, code)

	// 生成token
	token, _ := uuid.GenerateUUID()
	tokenMap[token] = "user_id"

	type tokenJSON struct {
		AccessToken  string         `json:"access_token"`
		TokenType    string         `json:"token_type"`
		RefreshToken string         `json:"refresh_token"`
		ExpiresIn    int32          `json:"expires_in"` // at least PayPal returns string, while most return number
	}
	var tj tokenJSON
	tj.AccessToken = token
	tj.TokenType = "Bearer"
	tj.ExpiresIn = int32(time.Now().Add(time.Hour * 24).Unix())
	tj.RefreshToken = ""
	rawData, _ := json.Marshal(&tj)
	w.Header().Set("Content-Type", "application/json")
	w.Write(rawData)
	glog.Infof("oauthToken resp:%s", string(rawData))
	return
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("access_token")
	userId, ok := tokenMap[token]
	if !ok {
		glog.Errorf("invalid token:%s", token)
		w.Write([]byte("invalid token"))
		return
	}
	w.Write([]byte(userId + "@xx.com"))
	return
}
