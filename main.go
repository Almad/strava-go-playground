package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/zalando/go-keyring"
)

type ClientConfig struct {
	Client_id string
}

type AccessToken struct {
	Token      string
	Expires_at string
}

type RefreshToken string

type StravaResponse struct {
	Refresh_token string
	Access_token  string
	Expires_in    string
}

const ServiceName = "blog.almad.weeknotes.strava"

func GetClientId() string {

	data, err := ioutil.ReadFile("/Users/almad/Library/Application Support/Weeknotes/config.json")
	if err != nil {
		fmt.Print(err)
	}

	var config ClientConfig

	err = json.Unmarshal([]byte(data), &config)

	if err != nil {
		fmt.Println(err)
	}

	return config.Client_id
}

func GetClientSecret() string {
	clientId := GetClientId()

	// get password
	secret, err := keyring.Get(ServiceName, clientId)
	if err != nil {
		log.Fatal(err)
	}

	return secret
}

func GetAccessToken() AccessToken {
	clientId := GetClientId()

	// get password
	secret, err := keyring.Get(ServiceName, clientId+".access_token")
	if err != nil {
		log.Fatal(err)
	}

	var accessToken AccessToken

	err = json.Unmarshal([]byte(secret), &accessToken)
	if err != nil {
		log.Fatal(err)
	}

	return accessToken
}

func GetRefreshToken() RefreshToken {
	client_id := GetClientId()

	// get password
	secret, err := keyring.Get(ServiceName, client_id+".refresh_token")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(secret)

	var refreshToken RefreshToken

	err = json.Unmarshal([]byte(secret), &refreshToken)
	if err != nil {
		log.Fatal(err)
	}

	return refreshToken
}

func ExchangeCodeForAccessToken(code string) string {
	data := url.Values{
		"client_id":     {GetClientId()},
		"client_secret": {GetClientSecret()},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	}

	fmt.Println("Sending data to strava")
	fmt.Println(data)

	resp, err := http.PostForm("https://www.strava.com/oauth/token", data)

	if err != nil {
		log.Fatal(err)
	}

	var stravaResponse StravaResponse

	json.NewDecoder(resp.Body).Decode(&stravaResponse)

	return stravaResponse.Access_token
}

func main() {
	// accessToken := GetAccessToken()
	// refreshToken := GetRefreshToken()

	// fmt.Println(accessToken.Token)
	// fmt.Println(refreshToken)

	var stravaAccessToken = ""

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET params were:", r.URL.Query())

		code := r.URL.Query().Get("code")

		stravaAccessToken = ExchangeCodeForAccessToken(code)

		fmt.Fprint(w, "Welcome to my website!")
	})

	clientId := GetClientId()
	port := "8080"

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	}()

	cmd := exec.Command("open", fmt.Sprintf("https://www.strava.com/oauth/authorize?client_id=%s&response_type=code&redirect_uri=http://localhost:%s/exchange_token&approval_prompt=force&scope=activity:read_all", clientId, port))

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)

		if stravaAccessToken != "" {
			break
		}
	}

	fmt.Println(stravaAccessToken)

}
