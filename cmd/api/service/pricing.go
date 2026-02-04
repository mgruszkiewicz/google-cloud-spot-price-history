// Package service contains business logic for the API.
package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mgruszkiewicz/google-cloud-spot-price-history/cmd/api/models"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"
)

// PricingService provides business logic for pricing data operations.
type PricingService struct {
	querier *db.Querier
}

// NewPricingService creates a new PricingService instance.
func NewPricingService(querier *db.Querier) *PricingService {
	return &PricingService{querier: querier}
}

// GetAllRegions returns a list of all available regions.
func (s *PricingService) GetAllRegions() ([]string, error) {
	var regions []string
	err := s.querier.QueryRows(
		"SELECT DISTINCT(region_name) FROM pricing_history ORDER BY region_name",
		func(rows *sql.Rows) error {
			var region string
			if err := rows.Scan(&region); err != nil {
				return fmt.Errorf("failed to scan region: %w", err)
			}
			regions = append(regions, region)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query regions: %w", err)
	}
	return regions, nil
}

// GetMachinesByRegion returns all machine types for a given region.
func (s *PricingService) GetMachinesByRegion(regionName string) ([]models.Machine, error) {
	query := `
		SELECT 
			DISTINCT(machine_type), 
			MIN(spot_hour_price), 
			MAX(spot_hour_price), 
			spot_hour_price 
		FROM pricing_history 
		WHERE region_name = ? 
		GROUP BY machine_type 
		ORDER BY machine_type DESC`

	var machines []models.Machine
	err := s.querier.QueryRows(query, func(rows *sql.Rows) error {
		var machine models.Machine
		machine.RegionName = regionName
		if err := rows.Scan(&machine.MachineType, &machine.MinHourSpotPrice, &machine.MaxHourSpotPrice, &machine.HourSpotPrice); err != nil {
			return fmt.Errorf("failed to scan machine: %w", err)
		}
		machines = append(machines, machine)
		return nil
	}, regionName)

	if err != nil {
		return nil, fmt.Errorf("failed to query machines: %w", err)
	}

	return machines, nil
}

// GetMachineDetail returns detailed information about a specific machine type in a region.
func (s *PricingService) GetMachineDetail(regionName, machineType string) (*models.MachineDetail, error) {
	result := &models.MachineDetail{
		MachineType: machineType,
		RegionName:  regionName,
	}

	// Get price history
	historyQuery := `
		SELECT spot_hour_price, updated_ts 
		FROM pricing_history 
		WHERE region_name = ? AND machine_type = ? 
		ORDER BY updated_ts ASC`

	err := s.querier.QueryRows(historyQuery, func(rows *sql.Rows) error {
		var price float64
		var timestampUnix int64
		if err := rows.Scan(&price, &timestampUnix); err != nil {
			return fmt.Errorf("failed to scan price history: %w", err)
		}

		result.SpotHourPriceHistory = append(result.SpotHourPriceHistory, models.PriceHistory{
			Price:     price,
			Timestamp: time.Unix(timestampUnix, 0),
		})
		return nil
	}, regionName, machineType)

	if err != nil {
		return nil, fmt.Errorf("failed to query price history: %w", err)
	}

	// Get aggregate statistics
	statsQuery := `
		SELECT MIN(spot_hour_price), MAX(spot_hour_price), spot_hour_price 
		FROM pricing_history 
		WHERE region_name = ? AND machine_type = ? 
		ORDER BY updated_ts DESC 
		LIMIT 1`

	var minPrice, maxPrice, spotHourPrice float64
	err = s.querier.QueryRow(statsQuery, func(row *sql.Row) error {
		return row.Scan(&minPrice, &maxPrice, &spotHourPrice)
	}, regionName, machineType)

	if err != nil {
		return nil, fmt.Errorf("failed to query statistics: %w", err)
	}

	result.MinHourSpotPrice = minPrice
	result.MaxHourSpotPrice = maxPrice
	result.HourSpotPrice = spotHourPrice

	return result, nil
}
