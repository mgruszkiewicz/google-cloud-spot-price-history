// Package models contains data structures used throughout the API.
package models

import "time"

// Region represents a Google Cloud region.
type Region struct {
	Name string `json:"name" example:"us-central1"`
}

// Machine represents a machine type with pricing information.
type Machine struct {
	MachineType      string  `json:"machine_type" example:"n1-standard-1"`
	RegionName       string  `json:"region_name" example:"us-central1"`
	MinHourSpotPrice float64 `json:"min_hour_spot_price" example:"0.01"`
	MaxHourSpotPrice float64 `json:"max_hour_spot_price" example:"0.05"`
	HourSpotPrice    float64 `json:"hour_spot_price" example:"0.02"`
}

// PriceHistory represents a single price data point.
type PriceHistory struct {
	Price     float64   `json:"price" example:"0.02"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// MachineDetail contains full machine information including price history.
type MachineDetail struct {
	MachineType          string         `json:"machine_type"`
	RegionName           string         `json:"region_name"`
	MinHourSpotPrice     float64        `json:"min_hour_spot_price"`
	MaxHourSpotPrice     float64        `json:"max_hour_spot_price"`
	HourSpotPrice        float64        `json:"hour_spot_price"`
	SpotHourPriceHistory []PriceHistory `json:"spot_hour_price_history"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Message string `json:"message" example:"Detailed error message"`
	Code    int    `json:"code" example:"400"`
}

// SuccessResponse represents a generic success response.
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RegionListResponse represents a list of regions response.
type RegionListResponse struct {
	Regions []string `json:"regions"`
	Count   int      `json:"count"`
}

// MachineListResponse represents a list of machines response.
type MachineListResponse struct {
	RegionName string    `json:"region_name"`
	Machines   []Machine `json:"machines"`
	Count      int       `json:"count"`
}
