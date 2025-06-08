package main

import (
	"database/sql"
	"errors"
	"flag"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"

	"github.com/mgruszkiewicz/google-cloud-spot-price-history/cmd/api/handlers"
)

func main() {
	e := echo.New()
	database_path := flag.String("dbpath", "db.sqlite3", "Path to sqlite3 database containing data from dataprocessing")
	flag.Parse()
	sqlDB, err := sql.Open("sqlite3", *database_path)
	if err != nil {
		slog.Error("failed opening connection to sqlite", "error", err)
		return
	}
	defer sqlDB.Close()

	q := db.NewQuerier(sqlDB)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handlers.HelloHandler(q))
	e.GET("/region", handlers.RegionHandler(q))
	e.GET("/compute", handlers.ComputeHandler(q))

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
