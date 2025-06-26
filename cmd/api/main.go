package main

import (
	"database/sql"
	"errors"
	"flag"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/mgruszkiewicz/google-cloud-spot-price-history/cmd/api/handlers"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/internal/db"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("cmd/api/templates/*.html")),
	}
	e.Renderer = renderer

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handlers.RootHandler(q))
	e.GET("/region", handlers.RegionHandler(q))
	e.GET("/compute", handlers.ComputeHandler(q))

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

