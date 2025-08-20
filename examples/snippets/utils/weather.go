// Package tools provides various utility tools for agent workflows, including weather-related operations.
package utils

import (
	"context"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// NewWeatherTool creates a new weather tool.
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

var _ tools.Tool = &WeatherTool{}

// WeatherTool is a tool that provides current weather information for a given city.
type WeatherTool struct{}

// Call implements Tool.
func (t *WeatherTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	// Extract city from parameters
	city, ok := params.Arguments["city"]
	if !ok {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "city parameter is required",
			},
		}, nil
	}

	cityStr, ok := city.(string)
	if !ok {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "city parameter must be a string",
			},
		}, nil
	}

	// Extract units from parameters (default to metric)
	units := "metric"
	if unitsArg, ok := params.Arguments["units"]; ok {
		if unitsStr, ok := unitsArg.(string); ok {
			units = unitsStr
		}
	}

	// Create weather data response
	weatherData := map[string]any{
		"city": cityStr,
		"temperature": map[string]any{
			"current":    22.5,
			"feels_like": 24.0,
			"min":        18.0,
			"max":        26.0,
		},
		"conditions": map[string]any{
			"main":        "Clear",
			"description": "clear sky",
		},
		"humidity": 65,
		"pressure": 1013,
		"wind": map[string]any{
			"speed": 5.2,
			"deg":   180,
		},
		"units":   units,
		"success": true,
	}

	// If units is imperial, convert temperatures
	if units == "imperial" {
		weatherData["temperature"] = map[string]any{
			"current":    72.5,
			"feels_like": 75.2,
			"min":        64.4,
			"max":        78.8,
		}
		weatherData["wind"].(map[string]any)["speed"] = 11.6
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result:     weatherData,
	}, nil
}

// Descriptor implements Tool.
func (t *WeatherTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "get_weather",
		Description: "Get current weather information for a specified city. Provides temperature, conditions, humidity, wind speed, and pressure.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"city": {
					Type:        llms.TypeString,
					Description: "The name of the city to get weather information for (e.g., 'New York', 'London', 'Tokyo').",
				},
				"units": {
					Type:        llms.TypeString,
					Description: "Units for temperature and wind speed. Use 'metric' for Celsius and m/s, 'imperial' for Fahrenheit and mph. Defaults to 'metric'.",
				},
			},
			Required: []string{"city"},
		},
	}
}
