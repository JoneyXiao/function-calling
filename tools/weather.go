package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var weatherURL = "https://api.open-meteo.com/v1/forecast"

// WeatherToolDefine defines the OpenAI tool for getting weather information
var WeatherToolDefine = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name: "GetWeather",
		Description: `
		Use this tool to get current weather and forecast information for a specific location.
		Example:
			"What's the weather in Shenzhen?"
		Then Action Input is: {"latitude": 22.547, "longitude": 114.058}
		
		You can also request specific weather parameters:
		{"latitude": 22.547, "longitude": 114.058, "current": ["temperature_2m", "weather_code"], "daily": ["temperature_2m_max", "temperature_2m_min"]}
		`,
		Parameters: `{
			"type": "object",
			"properties": {
				"latitude": {
					"type": "number",
					"description": "Latitude coordinate of the location"
				},
				"longitude": {
					"type": "number",
					"description": "Longitude coordinate of the location"
				},
				"current": {
					"type": "array",
					"items": {"type": "string"},
					"description": "Current weather parameters to include (e.g., temperature_2m, relative_humidity_2m, wind_speed_10m, weather_code)"
				},
				"hourly": {
					"type": "array",
					"items": {"type": "string"},
					"description": "Hourly forecast parameters to include (e.g., temperature_2m, relative_humidity_2m, wind_speed_10m, weather_code)"
				},
				"daily": {
					"type": "array",
					"items": {"type": "string"},
					"description": "Daily forecast parameters to include (e.g., temperature_2m_max, temperature_2m_min, precipitation_sum)"
				},
				"timezone": {
					"type": "string",
					"description": "Timezone for the data (e.g., 'GMT', 'America/New_York', or 'auto' for automatic detection)"
				}
			},
			"required": ["latitude", "longitude"]
		}`,
	},
}

// WeatherParams contains parameters for weather requests
type WeatherParams struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Current   []string `json:"current,omitempty"`
	Hourly    []string `json:"hourly,omitempty"`
	Daily     []string `json:"daily,omitempty"`
	Timezone  string   `json:"timezone,omitempty"`
}

// OpenMeteoResponse represents the response structure from Open-Meteo API
type OpenMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Current   struct {
		Time               string  `json:"time"`
		Temperature2m      float64 `json:"temperature_2m"`
		RelativeHumidity2m float64 `json:"relative_humidity_2m,omitempty"`
		WindSpeed10m       float64 `json:"wind_speed_10m,omitempty"`
		WeatherCode        int     `json:"weather_code,omitempty"`
	} `json:"current"`
	Hourly *struct {
		Time               []string  `json:"time"`
		Temperature2m      []float64 `json:"temperature_2m,omitempty"`
		RelativeHumidity2m []float64 `json:"relative_humidity_2m,omitempty"`
		WindSpeed10m       []float64 `json:"wind_speed_10m,omitempty"`
		WeatherCode        []int     `json:"weather_code,omitempty"`
	} `json:"hourly,omitempty"`
	Daily *struct {
		Time             []string  `json:"time"`
		Temperature2mMax []float64 `json:"temperature_2m_max,omitempty"`
		Temperature2mMin []float64 `json:"temperature_2m_min,omitempty"`
		PrecipitationSum []float64 `json:"precipitation_sum,omitempty"`
	} `json:"daily,omitempty"`
}

// weatherCodeToDescription maps weather codes to human-readable descriptions
var weatherCodeToDescription = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	61: "Slight rain",
	63: "Moderate rain",
	65: "Heavy rain",
	71: "Slight snow fall",
	73: "Moderate snow fall",
	75: "Heavy snow fall",
	80: "Slight rain showers",
	81: "Moderate rain showers",
	82: "Violent rain showers",
	95: "Thunderstorm",
	96: "Thunderstorm with slight hail",
	99: "Thunderstorm with heavy hail",
}

// GetWeather fetches weather data from Open-Meteo API
func GetWeather(params WeatherParams) (string, error) {
	queryParams := url.Values{}

	// Add required parameters
	queryParams.Add("latitude", fmt.Sprintf("%.6f", params.Latitude))
	queryParams.Add("longitude", fmt.Sprintf("%.6f", params.Longitude))

	// Add optional parameters
	if len(params.Current) > 0 {
		queryParams.Add("current", strings.Join(params.Current, ","))
	} else {
		// Default current parameters if none provided
		queryParams.Add("current", "temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m")
	}

	if len(params.Hourly) > 0 {
		queryParams.Add("hourly", strings.Join(params.Hourly, ","))
	}

	if len(params.Daily) > 0 {
		queryParams.Add("daily", strings.Join(params.Daily, ","))
	}

	if params.Timezone != "" {
		queryParams.Add("timezone", params.Timezone)
	} else {
		queryParams.Add("timezone", "auto")
	}

	// Build the full URL
	fullURL := fmt.Sprintf("%s?%s", weatherURL, queryParams.Encode())

	// Make the HTTP request
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("error making request to Open-Meteo API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Check if response status is not 200 OK
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Parse the JSON response
	var weatherData OpenMeteoResponse
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %w", err)
	}

	// Format the response in a user-friendly way
	var result strings.Builder
	result.WriteString("Current Weather:\n")
	result.WriteString(fmt.Sprintf("Time: %s\n", weatherData.Current.Time))
	result.WriteString(fmt.Sprintf("Temperature: %.1f°C\n", weatherData.Current.Temperature2m))

	if weatherData.Current.RelativeHumidity2m != 0 {
		result.WriteString(fmt.Sprintf("Humidity: %.1f%%\n", weatherData.Current.RelativeHumidity2m))
	}

	if weatherData.Current.WindSpeed10m != 0 {
		result.WriteString(fmt.Sprintf("Wind Speed: %.1f km/h\n", weatherData.Current.WindSpeed10m))
	}

	if weatherData.Current.WeatherCode != 0 || weatherData.Current.WeatherCode == 0 {
		desc, ok := weatherCodeToDescription[weatherData.Current.WeatherCode]
		if ok {
			result.WriteString(fmt.Sprintf("Conditions: %s\n", desc))
		}
	}

	// Include hourly forecast if requested
	if weatherData.Hourly != nil && len(weatherData.Hourly.Time) > 0 {
		result.WriteString("\nHourly Forecast (next 24 hours):\n")
		// Limit to 24 hours for readability
		limit := len(weatherData.Hourly.Time)
		if limit > 24 {
			limit = 24
		}

		for i := 0; i < limit; i++ {
			result.WriteString(fmt.Sprintf("Time: %s\n", weatherData.Hourly.Time[i]))

			if len(weatherData.Hourly.Temperature2m) > i {
				result.WriteString(fmt.Sprintf("  Temperature: %.1f°C\n", weatherData.Hourly.Temperature2m[i]))
			}

			if len(weatherData.Hourly.RelativeHumidity2m) > i {
				result.WriteString(fmt.Sprintf("  Humidity: %.1f%%\n", weatherData.Hourly.RelativeHumidity2m[i]))
			}

			if len(weatherData.Hourly.WindSpeed10m) > i {
				result.WriteString(fmt.Sprintf("  Wind Speed: %.1f km/h\n", weatherData.Hourly.WindSpeed10m[i]))
			}

			if len(weatherData.Hourly.WeatherCode) > i {
				if desc, ok := weatherCodeToDescription[weatherData.Hourly.WeatherCode[i]]; ok {
					result.WriteString(fmt.Sprintf("  Conditions: %s\n", desc))
				}
			}

			result.WriteString("\n")
		}
	}

	// Include daily forecast if requested
	if weatherData.Daily != nil && len(weatherData.Daily.Time) > 0 {
		result.WriteString("\nDaily Forecast:\n")
		for i := 0; i < len(weatherData.Daily.Time); i++ {
			result.WriteString(fmt.Sprintf("Date: %s\n", weatherData.Daily.Time[i]))

			if len(weatherData.Daily.Temperature2mMax) > i {
				result.WriteString(fmt.Sprintf("  Max Temperature: %.1f°C\n", weatherData.Daily.Temperature2mMax[i]))
			}

			if len(weatherData.Daily.Temperature2mMin) > i {
				result.WriteString(fmt.Sprintf("  Min Temperature: %.1f°C\n", weatherData.Daily.Temperature2mMin[i]))
			}

			if len(weatherData.Daily.PrecipitationSum) > i {
				result.WriteString(fmt.Sprintf("  Precipitation: %.1f mm\n", weatherData.Daily.PrecipitationSum[i]))
			}

			result.WriteString("\n")
		}
	}

	return result.String(), nil
}
