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

const (
	dayStartHour      = 0
	dayEndHour        = 14
	rainMsg           = "Rain Expected."
	noRainTodayMsg    = "No rain expected today."
	noRainTomorrowMsg = "No rain expected tomorrow."
	iconPath          = `C:\Users\giann\Downloads\cloud-rain-solid.svg`
)

//const iconPath = `C:\Users\giann\Downloads\cloud-rain-solid.svg`

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
		Temperature      []float64 `json:"temperature_2m_min"`
		Snowfall         []float64 `json:"snowfall_sum"`
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
		checkWeatherData(ctx, loc.Latitude, loc.Longitude, loc.Name, loc.CityID)
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

func checkWeatherData(ctx context.Context, latitude, longitude, name, cityID string) {
	//url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", latitude, longtitude)
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum,temperature_2m_min,snowfall_sum&timezone=auto", latitude, longitude)

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

	validateWeatherData(weatherResp, name)

	handleNotification(weatherResp, actions, name)
}

func handleNotification(weatherResp WeatherResponse, actions []toast.Action, name string) {
	timeNow := time.Now()

	day, targetDate, temperature := determineDayAndTemperature(weatherResp, timeNow)

	temperatureMessage := checkColdTemperature(temperature)

	for i, data := range weatherResp.Daily.Time {
		if data == targetDate {
			rainExpected := weatherResp.Daily.PrecipitationSum[i] > 0

			notificationMsg := generateNotificationMessage(day, rainExpected, temperatureMessage)

			if notificationMsg != "" {
				toastNotification(actions, name, notificationMsg)
				fmt.Println(notificationMsg)
				fmt.Println("Toast notification sent")
			} else {
				// Display no rain message if no precipitation and no temperature message
				if timeNow.Hour() >= dayEndHour {
					fmt.Println(noRainTomorrowMsg)
				} else {
					fmt.Println(noRainTodayMsg)
				}
			}
			break
		}
	}
}

func checkColdTemperature(temperature float64) (message string) {
	// TODO: Add warm weather checks in the feature
	switch {
	case temperature <= 25:
		return "A bit chilly. <= 25"
	case temperature <= 20:
		return "It's going to be slightly cold. <= 20c"
	case temperature <= 15:
		return "It's going to be cold. <= 15c"
	case temperature <= 10:
		return "Very cold. <= 10c"
	case temperature <= 5:
		return "Warning! Freezing cold. <= 5c"
	case temperature <= 0:
		return "!Extreme cold waring! <= 0c"
	default:
		return "Fine weather tomorrow! > 25c"
	}
}

func validateWeatherData(weatherResp WeatherResponse, name string) error {
	if len(weatherResp.Daily.PrecipitationSum) == 0 || len(weatherResp.Daily.Time) == 0 {
		return fmt.Errorf("no weather data available for %s", name)
	}

	if len(weatherResp.Daily.PrecipitationSum) != len(weatherResp.Daily.Time) {
		return fmt.Errorf("mismatch in data length for %s between precipitation and time", name)
	}

	return nil
}

func determineDayAndTemperature(weatherResp WeatherResponse, timeNow time.Time) (string, string, float64) {
	hourNow := timeNow.Hour()

	var day string
	var targetDate string
	var temperature float64

	if hourNow >= dayStartHour && hourNow < dayEndHour {
		day = "Today: "
		targetDate = timeNow.Format("2006-01-02")
		temperature = weatherResp.Daily.Temperature[0]
	} else {
		day = "Tomorrow: "
		targetDate = timeNow.Add(24 * time.Hour).Format("2006-01-02")
		temperature = weatherResp.Daily.Temperature[1]
	}

	return day, targetDate, temperature
}

func toastNotification(actions []toast.Action, name string, message string) {
	title := fmt.Sprintf("Rain forecasted for %s", name)

	notification := toast.Notification{
		AppID:   "Will it Rain API",
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

func generateNotificationMessage(day string, rain bool, temperatureMessage string) string {
	if rain && temperatureMessage != "" {
		return fmt.Sprintf("%s%s%s", day, rainMsg, temperatureMessage)
	} else if rain {
		return fmt.Sprintf("%s%s", day, rainMsg)
	} else if temperatureMessage != "" {
		return fmt.Sprintf("%s%s", day, temperatureMessage)
	}
	return ""
}
