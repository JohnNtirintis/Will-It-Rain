I drive a motorcycle for commuting and there have been way too many times where I didn't know it was gonna rain and i forgot my jacket or umbrella. This very small project aims to solve that. I personally have created a task using Task Scheduler on Windows that runs on startup and every 3 hours.

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
``` go build -o weather-notifier 
./weather-notifier ```