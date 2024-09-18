package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-toast/toast"
)

// Constants
// Location for API
const pfLat = "37.9278"
const pfLong = "23.7036"

const korLat = "37.9011"
const korLong = "23.8727"

// Vars for toast notification
const appID = "Rain Detector API"
const title = "Rain Tomorrow!"
const message = "Get your Jacket!"
const iconPath = `C:\Users\giann\Downloads\cloud-rain-solid.svg`

type WeatherResponse struct {
	Daily struct {
		PrecipitationSum []float64 `json:"precipitation_sum"`
		Time             []string  `json:"time"`
	} `json:"daily"`
}

var locations = []struct {
	Latitude  string
	Longitude string
	Name      string
}{
	{"37.9278", "23.7036", "Palaio Faliro"},
	{"37.9011", "23.8727", "Koropi"},
}

func main() {

	pfUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", pfLat, pfLong)
	//korUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", korLat, korLong)

	actions := []toast.Action{
		{Type: "protocol", Label: "More Info", Arguments: "https://www.example.com"},
		{Type: "protocol", Label: "Close", Arguments: "close-app"},
	}

	resp, err := http.Get(pfUrl)
	if err != nil {
		fmt.Println("Error getting weather data:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var weatherResp WeatherResponse

	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		fmt.Println("Error parsing weather data:", err)
		return
	}

	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	for i, date := range weatherResp.Daily.Time {
		if date == tomorrow {
			precipitation := weatherResp.Daily.PrecipitationSum[i]
			if precipitation > 0 {
				toastNotification(actions)
				fmt.Println("Rain expected tomorrow.")
				fmt.Println("Toast notification sent.")
			} else {
				fmt.Println("No rain expected tomorrow")
			}
			break
		}
	}
}

func toastNotification(actions []toast.Action) {
	notification := toast.Notification{
		AppID:   appID,
		Title:   title,
		Message: message,
		Icon:    iconPath,
		Actions: actions,
	}
	err := notification.Push()
	if err != nil {
		log.Println("Error", err)
		log.Printf("**********")
		log.Fatalln(err)
	}
}
