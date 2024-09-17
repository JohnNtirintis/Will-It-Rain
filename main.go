package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/go-toast/toast"
)

const pfLat = "37.9278"
const pfLong = "23.7036"

const korLat = "37.9011"
const korLong = "23.8727"

type WeatherResponse struct {
	Daily struct {
		PrecipitationSum []float64 `json:"precipitaion_sum"`
		Time             []string  `json:"time"`
	} `json:"daily"`
}

func main() {
	pfUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", pfLat, pfLong)
	korUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&daily=precipitation_sum&timezone=auto", korLat, korLong)

	resp, err := http.get(url)
	if err != nil {
		fmt.Println("Error getting weather data:", err)
		return
	}
	defer resp.body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var weatherResp WeatherResponse

	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		fmt.Println("Error parsing weather data:", err)
		return
	}

	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	for i, date := range weatherResp.Daily.Time {
		if date == tomorrow {
			precipitation = weatherResp.Daily.PrecipitationSum[i]
			if precipitaion > 0 {
				// TODO: Add params etc.
				notification()
				fmt.PrintLn("No rain expected tomorrow")
			}
			else {
				fmt.PrintLn("No rain expected tomorrow")
			}
			break
		}
	}
}

func notification() {
	notification := toast.Notification{
		AppID:   "Rain Detector API",
		Title:   "Rain Tomorrow!",
		Message: "Get your jacket",
		Icon:    "C:\\Users\\giann\\Downloads\\cloud-rain-solid.svg",
		Actions: []toast.Action{
			{Type: "protocol", Label: "More Info", Arguments: ""},
			{Type: "protocol", Label: "Close", Arguments: ""},
		},
	}
	err := notification.Push()
	if err != nil {
		log.Fatalln(err)
	}
}
