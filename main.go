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
	CityID    string
}{
	{"37.9278", "23.7036", "Palaio Faliro", "2281820"},
	{"37.9011", "23.8727", "Koropi", "4-182368_1_al"},
}

func main() {
	for _, loc := range locations {
		checkWeather(loc.Latitude, loc.Longitude, loc.Name, loc.CityID)
	}
}

func checkWeather(latitude, longtitude, name, cityID string) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", latitude, longtitude)

	moreInfoUrl := fmt.Sprintf("https://www.accuweather.com/en/gr/%s/%s/weather-forecast/%s", name, cityID, cityID)

	actions := []toast.Action{
		{Type: "protocol", Label: "More Info", Arguments: moreInfoUrl},
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

	timeNow := time.Now()
	var targetDate string
	var notificationMsg string

	// If time is between midnight and noon
	// Get todays weather
	// Else get tomorrow's weather
	// (It's just my personal preference)
	if timeNow.Hour() > 00 && timeNow.Hour() < 14 {
		targetDate = timeNow.Format("2006-01-02")
		notificationMsg = "Rain expected today."
	} else {
		targetDate = timeNow.Add(24 * time.Hour).Format("2006-01-02")
		notificationMsg = "Rain expected tomorrow."
	}

	for i, date := range weatherResp.Daily.Time {
		if date == targetDate {
			precipitation := weatherResp.Daily.PrecipitationSum[i]
			if precipitation > 0 {
				toastNotification(actions, name)
				fmt.Println(notificationMsg)
				fmt.Println("Toast notification sent")
				return
			} else {
				noRainMsg := "No rain expected today."
				if timeNow.Hour() >= 14 {
					noRainMsg = "No rain expected tomorrow."
				}

				fmt.Println(noRainMsg)
				return
			}
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
