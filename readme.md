I commute by motorcycle, and there have been far too many times when I didn't know (or even expected that) rain was forecasted. This small project is designed to help prevent that. Personally, I’ve set up a task in Windows Task Scheduler to run the program at startup and every three hours.

# 🌧️ Will It Rain? - Weather Forecast Notification
This Go project fetches weather data from the Open-Meteo API and sends a toast notification on Windows if rain is forecasted for specific locations. It’s designed to help you prepare for rainy weather by sending a quick desktop alert to get your jacket or umbrella!

## Setup

1. Clone this repository to your local machine.
2. Install Go on your machine.
3. Run the following command to get the dependencies:
``` go get github.com/go-toast/toast ```
4. Modify the `locations.json` file to set your preferred locations.

### Example `locations.json`:

```json
[
 {
     "latitude": "37.9278",
     "longitude": "23.7036",
     "name": "Palaio Faliro",
     "cityID": "2281820"
 },
 {
     "latitude": "37.9011",
     "longitude": "23.8727",
     "name": "Koropi",
     "cityID": "4-182368_1_al"
 }
]
```

### Running the code
``` go run main.go ```

### Alternatively, you can compile the program with:
``` go build -o weather-notifier ```

```./weather-notifier ```

### Disclaimer:
The main purpose of this project -- apart from saving me from getting wet -- is to help me learn Go. This repo is my first time using Go, so mistakes are expected.
