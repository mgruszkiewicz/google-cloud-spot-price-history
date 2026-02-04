// Package main is the entry point for the API server.
// Uses Fuego for auto-generating OpenAPI documentation from code.
//
// API docs available at: http://localhost:8080/swagger/index.html
package main

import (
	"database/sql"
	"errors"
	"flag"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"

	"github.com/mgruszkiewicz/google-cloud-spot-price-history/cmd/api/models"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/cmd/api/service"
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
	// Parse command line flags
	dbPath := flag.String("dbpath", "db.sqlite3", "Path to sqlite3 database containing data from dataprocessing")
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Initialize database connection
	sqlDB, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		slog.Error("failed opening connection to sqlite", "error", err)
		return
	}
	defer sqlDB.Close()

	// Test database connection
	if err := sqlDB.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		return
	}

	// Initialize querier and service
	querier := db.NewQuerier(sqlDB)
	pricingService := service.NewPricingService(querier)

	// Create Fuego server with OpenAPI auto-generation
	s := fuego.NewServer(
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				DisableSwaggerUI: false,
				DisableLocalSave: true,
				PrettyFormatJSON: true,
				Info: &openapi3.Info{
					Title:       "Google Cloud Spot Price History API",
					Version:     "1.0.0",
					Description: "API for querying Google Cloud spot instance pricing history",
				},
			}),
		),
	)

	// Add server URL to OpenAPI spec (fixes the "null" base URL issue in Swagger UI)
	serverURL := "http://localhost:" + *port
	s.Engine.OpenAPI.Description().Servers = openapi3.Servers{
		&openapi3.Server{
			URL:         serverURL,
			Description: "Local development server",
		},
	}

	// Setup template renderer for HTML routes
	e := echo.New()
	e.HideBanner = true
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("cmd/api/templates/*.html")),
	}
	e.Renderer = renderer
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())

	// HTML Routes (existing functionality using Echo directly)
	e.GET("/", func(c echo.Context) error {
		regions, err := pricingService.GetAllRegions()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to query regions: "+err.Error())
		}
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{"Regions": regions})
	})

	e.GET("/region", func(c echo.Context) error {
		regionName := c.QueryParam("region_name")
		if regionName == "" {
			return c.HTML(http.StatusOK, "")
		}
		machines, err := pricingService.GetMachinesByRegion(regionName)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to query machines: "+err.Error())
		}
		return c.Render(http.StatusOK, "machines.html", map[string]interface{}{
			"ListTitle": "Machine types in " + regionName,
			"Machines":  machines,
		})
	})

	e.GET("/compute", func(c echo.Context) error {
		regionName := c.QueryParam("region_name")
		machineType := c.QueryParam("machine_type")
		if regionName == "" || machineType == "" {
			return c.HTML(http.StatusOK, "")
		}
		machineData, err := pricingService.GetMachineDetail(regionName, machineType)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to query compute history: "+err.Error())
		}
		return c.Render(http.StatusOK, "prices.html", machineData)
	})

	// Fuego API Routes - Auto-generates OpenAPI docs from function signatures!
	// No annotations needed - types are inferred from return values

	// GET /api/v1/regions - Auto-documented from function signature
	fuego.Get(s, "/api/v1/regions", func(c fuego.ContextNoBody) (models.RegionListResponse, error) {
		regions, err := pricingService.GetAllRegions()
		if err != nil {
			return models.RegionListResponse{}, err
		}
		return models.RegionListResponse{
			Regions: regions,
			Count:   len(regions),
		}, nil
	},
		option.Summary("List all regions"),
		option.Description("Get a list of all available Google Cloud regions"),
		option.Tags("regions"),
	)

	// GET /api/v1/regions/{region}/machines
	fuego.Get(s, "/api/v1/regions/{region}/machines", func(c fuego.ContextNoBody) (models.MachineListResponse, error) {
		region := c.PathParam("region")
		machines, err := pricingService.GetMachinesByRegion(region)
		if err != nil {
			return models.MachineListResponse{}, err
		}
		return models.MachineListResponse{
			RegionName: region,
			Machines:   machines,
			Count:      len(machines),
		}, nil
	},
		option.Summary("List machines in a region"),
		option.Description("Get all machine types available in a specific region with pricing information"),
		option.Tags("machines"),
	)

	// GET /api/v1/regions/{region}/machines/{machine_type}/history
	fuego.Get(s, "/api/v1/regions/{region}/machines/{machine_type}/history", func(c fuego.ContextNoBody) (*models.MachineDetail, error) {
		region := c.PathParam("region")
		machineType := c.PathParam("machine_type")
		return pricingService.GetMachineDetail(region, machineType)
	},
		option.Summary("Get machine price history"),
		option.Description("Get detailed price history for a specific machine type in a region"),
		option.Tags("machines"),
	)

	// GET /api/v1/health
	fuego.Get(s, "/api/v1/health", func(c fuego.ContextNoBody) (map[string]string, error) {
		return map[string]string{
			"status":  "healthy",
			"service": "google-cloud-spot-price-history-api",
			"version": "1.0.0",
		}, nil
	},
		option.Summary("Health check"),
		option.Tags("health"),
	)

	// Mount Echo routes on Fuego server
	s.Mux.Handle("/", e)

	// Update server address
	s.Addr = ":" + *port

	// Start server
	slog.Info("starting server", "port", *port, "docs_url", "http://localhost:"+*port+"/swagger/index.html")
	if err := s.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
