package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"
)

func HelloHandler(q *db.Querier) echo.HandlerFunc {
	return func(c echo.Context) error {
		var count int
		err := q.QueryRow("SELECT COUNT(*) FROM pricing_history", func(row *sql.Row) error {
			return row.Scan(&count)
		})
		if err != nil {
			slog.Error("failed to get count of rows", "error", err)
			return c.String(http.StatusInternalServerError, "Failed to retrieve data")
		}
		return c.JSON(http.StatusOK, Hello{TotalRecords: count})
	}
}
