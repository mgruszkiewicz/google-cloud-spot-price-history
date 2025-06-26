package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"
)

func ComputeHandler(q *db.Querier) echo.HandlerFunc {
	return func(c echo.Context) error {
		var param Params
		if err := c.Bind(&param); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		if param.RegionName == "" || param.MachineType == "" {
			return c.HTML(http.StatusOK, "")
		}

		machineData, err := getMachineData(q, param.RegionName, param.MachineType)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to query compute history: "+err.Error())
		}

		return c.Render(http.StatusOK, "prices.html", machineData)
	}
}

func getMachineData(q *db.Querier, regionName, machineType string) (Machine, error) {
	var result Machine
	result.MachineType = machineType
	result.RegionName = regionName

	query := "SELECT spot_hour_price, updated_ts FROM pricing_history WHERE region_name = ? AND machine_type = ? ORDER BY updated_ts ASC"
	if err := q.QueryRows(query, func(rows *sql.Rows) error {
		var price float64
		var timestamp string
		if err := rows.Scan(&price, &timestamp); err != nil {
			return fmt.Errorf("failed to scan price history: %w", err)
		}
		result.SpotHourPriceHistory = append(result.SpotHourPriceHistory, PriceData{Price: price, Timestamp: timestamp})
		return nil
	}, regionName, machineType); err != nil {
		return result, err
	}

	var minPrice, maxPrice, spotHourPrice float64
	err := q.QueryRow(
		"SELECT MIN(spot_hour_price), MAX(spot_hour_price), spot_hour_price FROM pricing_history WHERE region_name = ? AND machine_type = ? ORDER BY updated_ts DESC LIMIT 1",
		func(row *sql.Row) error {
			return row.Scan(&minPrice, &maxPrice, &spotHourPrice)
		},
		regionName,
		machineType,
	)
	if err != nil {
		return result, err
	}

	result.MinHourSpotPrice = minPrice
	result.MaxHourSpotPrice = maxPrice
	result.HourSpotPrice = spotHourPrice

	return result, nil
}
