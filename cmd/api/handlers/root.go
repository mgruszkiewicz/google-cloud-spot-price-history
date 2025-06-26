package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"
)

func RootHandler(q *db.Querier) echo.HandlerFunc {
	return func(c echo.Context) error {
		var regions []string
		err := q.QueryRows("SELECT DISTINCT(region_name) FROM pricing_history", func(rows *sql.Rows) error {
			var region string
			if err := rows.Scan(&region); err != nil {
				return fmt.Errorf("failed to scan region: %w", err)
			}
			regions = append(regions, region)
			return nil
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to query regions: "+err.Error())
		}

		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Regions": regions,
		})
	}
}
