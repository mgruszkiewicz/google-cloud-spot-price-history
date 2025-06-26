package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"
)

func RegionHandler(q *db.Querier) echo.HandlerFunc {
	return func(c echo.Context) error {
		var param Params
		if err := c.Bind(&param); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		if param.RegionName == "" {
			return c.HTML(http.StatusOK, "")
		}

		machines, err := getMachinesByRegion(q, param.RegionName)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to query compute at region: "+err.Error())
		}

		return c.Render(http.StatusOK, "machines.html", map[string]interface{}{
			"ListTitle": fmt.Sprintf("machine types at %s", param.RegionName),
			"Machines":  machines,
		})
	}
}

func getMachinesByRegion(q *db.Querier, regionName string) ([]Machine, error) {
	query := "SELECT DISTINCT(machine_type), MIN(spot_hour_price), MAX(spot_hour_price), spot_hour_price FROM pricing_history WHERE region_name = ? GROUP BY machine_type ORDER BY machine_type DESC"
	var machines []Machine

	err := q.QueryRows(query, func(rows *sql.Rows) error {
		var machine Machine
		machine.RegionName = regionName
		if err := rows.Scan(&machine.MachineType, &machine.MinHourSpotPrice, &machine.MaxHourSpotPrice, &machine.HourSpotPrice); err != nil {
			return fmt.Errorf("failed to scan region: %w", err)
		}
		machines = append(machines, machine)
		return nil
	}, regionName)

	if err != nil {
		return nil, err
	}

	return machines, nil
}
