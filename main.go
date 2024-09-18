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
	for _, loc := range locations {
		checkWeather(loc.Latitude, loc.Longitude, loc.Name)
	}
}

func checkWeather(latitude, longtitude, name string) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", latitude, longtitude)

	actions := []toast.Action{
		{Type: "protocol", Label: "More Info", Arguments: "https://www.example.com"},
		{Type: "protocol", Label: "Close", Arguments: "close-app"},
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching weather data for %s: %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading API response for %s: %v\n", name, err)
		return
	}

	var weatherResp WeatherResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		fmt.Printf("Error parsing weather data for %s: %v\n", name, err)
		return
	}

	checkWeatherData(weatherResp, name)

	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	for i, date := range weatherResp.Daily.Time {
		if date == tomorrow {
			precipitation := weatherResp.Daily.PrecipitationSum[i]
			if precipitation > 0 {
				toastNotification(actions, name)
				fmt.Println("Rain expected tomorrow.")
				fmt.Println("Toast notification sent.")
			} else {
				fmt.Println("No rain expected tomorrow")
			}
			break
		}
	}
}

func checkWeatherData(weatherResp WeatherResponse, name string) error {
	if len(weatherResp.Daily.PrecipitationSum) == 0 || len(weatherResp.Daily.Time) == 0 {
		return fmt.Errorf("no weather data available for %s", name)
	}

	if len(weatherResp.Daily.PrecipitationSum) != len(weatherResp.Daily.Time) {
		return fmt.Errorf("mismatch in data length for %s between precipitation and time", name)
	}

	return nil
}

func toastNotification(actions []toast.Action, name string) {
	title := fmt.Sprintf("Rain forecasted for %s", name)
	notification := toast.Notification{
		AppID:   "Rain API",
		Title:   title,
		Message: "Get your jacket!",
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
