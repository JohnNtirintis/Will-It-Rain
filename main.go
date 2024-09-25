package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/go-toast/toast"
)

const iconPath = `C:\Users\giann\Downloads\cloud-rain-solid.svg`

type Location struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Name      string `json:"name"`
	CityID    string `json:"cityID"`
}

type WeatherResponse struct {
	Daily struct {
		PrecipitationSum []float64 `json:"precipitation_sum"`
		Time             []string  `json:"time"`
	} `json:"daily"`
}

// TODO: Use goroutines?
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	locations, err := loadLocations("locations.json")
	if err != nil {
		log.Fatalf("Error loading locations: %v", err)
	}

	for _, loc := range locations {
		checkWeather(ctx, loc.Latitude, loc.Longitude, loc.Name, loc.CityID)
	}
}

func makeRequestWithRetry(ctx context.Context, url string, maxRetries int) ([]byte, error) {
	var respBody []byte
	var err error

	delay := 1 * time.Second

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			// context is canceled, so return
			return nil, ctx.Err()
		default:
			respBody, err = makeHttpRequest(url)
			// is successful
			if err == nil {
				return respBody, nil
			}

			log.Printf("Attempt %d: Failed to fetch %s, error: %v", i+1, url, err)

			if i == maxRetries {
				return nil, fmt.Errorf("max retries reached for %s: %v", url, err)
			}

			// Exponential backoff with jitter
			time.Sleep(delay + time.Duration(rand.IntN(1000))*time.Millisecond)

			maxDelay := 10 * time.Second
			if delay > maxDelay {
				delay = maxDelay
			} else {
				delay *= 2
			}

		}
	}
	return nil, errors.New("unreachable code")
}

func makeHttpRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error during HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func loadLocations(filename string) ([]Location, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open locations file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read locations file: %v", err)
	}

	var locations []Location
	if err := json.Unmarshal(bytes, &locations); err != nil {
		return nil, fmt.Errorf("could not parse locations file: %v", err)
	}

	return locations, nil
}

func checkWeather(ctx context.Context, latitude, longtitude, name, cityID string) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", latitude, longtitude)

	moreInfoUrl := fmt.Sprintf("https://www.accuweather.com/en/gr/%s/%s/weather-forecast/%s", name, cityID, cityID)

	actions := []toast.Action{
		{Type: "protocol", Label: "More Info", Arguments: moreInfoUrl},
		{Type: "protocol", Label: "Close", Arguments: "close-app"},
	}

	respBody, err := makeRequestWithRetry(ctx, url, 2)
	if err != nil {
		fmt.Printf("Error fetching weather data for %s: %v\n", name, err)
		return
	}

	var weatherResp WeatherResponse
	err = json.Unmarshal(respBody, &weatherResp)
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
			} else {
				noRainMsg := "No rain expected today."
				if timeNow.Hour() >= 14 {
					noRainMsg = "No rain expected tomorrow."
				}

				fmt.Println(noRainMsg)
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
