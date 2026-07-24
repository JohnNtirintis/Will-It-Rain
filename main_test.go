package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateNotificationMessage(t *testing.T) {
	tests := []struct {
		name        string
		day         string
		rain        bool
		tempMsg     string
		expectedMsg string
	}{
		{"rain and cold", "Today: ", true, "Very cold. <= 10c", "Today: Rain Expected.Very cold. <= 10c"},
		{"rain only", "Today: ", true, "", "Today: Rain Expected."},
		{"cold only, no rain", "Tomorrow: ", false, "A bit chilly. <= 25c", "Tomorrow: A bit chilly. <= 25c"},
		{"no rain, no cold warning", "Today: ", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateNotificationMessage(tt.day, tt.rain, tt.tempMsg)
			if got != tt.expectedMsg {
				t.Errorf("generateNotificationMessage(%q, %v, %q) = %q, want %q",
					tt.day, tt.rain, tt.tempMsg, got, tt.expectedMsg)
			}
		})
	}
}

func TestCheckColdTemperature(t *testing.T) {
	wd := "C:\\fake\\dir"

	tests := []struct {
		name        string
		temperature float64
		wantIcon    string
	}{
		{"extreme cold", -5, snowIconPath},
		{"freezing", 3, snowIconPath},
		{"very cold", 8, coldIconPath},
		{"cold", 13, coldIconPath},
		{"slightly cold", 18, coldIconPath},
		{"chilly", 23, sunIconPath},
		{"fine weather", 30, sunIconPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkColdTemperature(tt.temperature, wd)
			wantPath := filepath.Join(wd, tt.wantIcon)
			if got.IconPath != wantPath {
				t.Errorf("checkColdTemperature(%v) icon = %q, want %q", tt.temperature, got.IconPath, wantPath)
			}
			if got.WeatherMessage == "" {
				t.Errorf("checkColdTemperature(%v) returned an empty message", tt.temperature)
			}
		})
	}
}

func TestValidateWeatherData(t *testing.T) {
	tests := []struct {
		name    string
		resp    WeatherResponse
		wantErr bool
	}{
		{
			name: "valid data",
			resp: makeWeatherResponse(
				[]string{"2026-01-01", "2026-01-02"},
				[]float64{0, 1.2},
				[]float64{10, 8},
			),
			wantErr: false,
		},
		{
			name:    "empty data",
			resp:    WeatherResponse{},
			wantErr: true,
		},
		{
			name: "mismatched lengths",
			resp: makeWeatherResponse(
				[]string{"2026-01-01", "2026-01-02"},
				[]float64{0},
				[]float64{10, 8},
			),
			wantErr: true,
		},
		{
			name: "not enough temperature data",
			resp: makeWeatherResponse(
				[]string{"2026-01-01", "2026-01-02"},
				[]float64{0, 1.2},
				[]float64{10},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWeatherData(tt.resp, "TestCity")
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWeatherData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetermineDayAndTemperature(t *testing.T) {
	resp := makeWeatherResponse(
		[]string{"2026-01-01", "2026-01-02"},
		[]float64{0, 1.2},
		[]float64{10, 8},
	)

	tests := []struct {
		name     string
		hour     int
		wantDay  string
		wantTemp float64
	}{
		{"morning uses today's temperature", 9, "Today: ", 10},
		{"late afternoon uses tomorrow's temperature", 15, "Tomorrow: ", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeNow := time.Date(2026, 1, 1, tt.hour, 0, 0, 0, time.UTC)
			day, _, temperature := determineDayAndTemperature(resp, timeNow)
			if day != tt.wantDay {
				t.Errorf("day = %q, want %q", day, tt.wantDay)
			}
			if temperature != tt.wantTemp {
				t.Errorf("temperature = %v, want %v", temperature, tt.wantTemp)
			}
		})
	}
}

func TestLoadLocations(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		content := `[{
        "latitude": "37.9011",
        "longitude": "23.8727",
        "name": "Koropi",
        "cityID": "4-182368_1_al"
    }]`
		path := filepath.Join(t.TempDir(), "locations.json")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		locations, err := loadLocations(path)
		if err != nil {
			t.Fatalf("loadLocations() returned unexpected error: %v", err)
		}
		if len(locations) != 1 {
			t.Fatalf("expected 1 location, got %d", len(locations))
		}
		if locations[0].Name != "Koropi" {
			t.Errorf("expected name %q, got %q", "Koropi", locations[0].Name)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadLocations(filepath.Join(t.TempDir(), "does-not-exist.json"))
		if err == nil {
			t.Error("expected an error for a missing file, got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := loadLocations(path)
		if err == nil {
			t.Error("expected an error for invalid JSON, got nil")
		}
	})
}

// makeWeatherResponse is a small test helper to build a WeatherResponse
// without repeating the nested struct literal in every test case.
func makeWeatherResponse(times []string, precipitation, temperature []float64) WeatherResponse {
	var resp WeatherResponse
	resp.Daily.Time = times
	resp.Daily.PrecipitationSum = precipitation
	resp.Daily.Temperature = temperature
	return resp
}
