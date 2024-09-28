package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"os"
	"strconv"
)

type BearerToken struct {
	Token string
	EndTime time.Time
}

var bearerToken BearerToken

type SpotifyData struct {
	Username string
	DisplayName string
	Followers string
	ProfileUrl string
}

type SpotifyImage struct {
	Url string
	Height int
	Width int
}

type ResponseSpotifyData struct {
	DisplayName string `json:"display_name"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href string `json:"href"`
		Total int `json:"total"`
	} `json:"followers"`
	Href string `json:"href"`
	Id string `json:"id"`
	Images []SpotifyImage `json:"images"`
	Type string `json:"type"`
	Uri string `json:"uri"`
}

func getToken() {
	client_id := os.Getenv("CLIENT_ID")
	client_secret := os.Getenv("CLIENT_SECRET")
	
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", client_id)
	formData.Set("client_secret", client_secret)

	url := "https://accounts.spotify.com/api/token"
	contentType := "application/x-www-form-urlencoded"
	data := formData.Encode()

	request, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		log.Fatal("Error creating request: ", err)
		return
	}

	request.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal("Error making request: ", err)
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response: ", err)
		return
	}

	var bodyData map[string]interface{}

	err = json.Unmarshal(body, &bodyData)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
		return
	}

	timeNow := time.Now()
	expiresInFloat, ok := bodyData["expires_in"].(float64)
	if !ok {
		log.Fatal("Error converting seconds from body: ", err)
		return
	}
	endTime := timeNow.Add(time.Second * time.Duration(int(expiresInFloat)))

	bearerToken = BearerToken{Token: bodyData["access_token"].(string), EndTime:endTime}
}

func GetSpotifyData(username string) *SpotifyData {
	if (bearerToken == (BearerToken{})) || (time.Now().Before(bearerToken.EndTime)){
		getToken()
	}

	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s", username)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: ", err)
		return &SpotifyData{}
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken.Token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal("Error making request: ", err)
		return &SpotifyData{}
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response: ", err)
		return &SpotifyData{}
	}

	var spotifyStructData ResponseSpotifyData
	err = json.Unmarshal(body, &spotifyStructData)
	if err != nil {
		log.Fatal("Error unmarshalling body: ", err)
	}

	spotifyData := &SpotifyData{
		Username: username,
		DisplayName: spotifyStructData.DisplayName,
		Followers: strconv.Itoa(spotifyStructData.Followers.Total),
		ProfileUrl: spotifyStructData.ExternalUrls.Spotify,
	}

	return spotifyData
}
