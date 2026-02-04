// Package handlers contains HTTP request handlers for the API.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgruszkiewicz/google-cloud-spot-price-history/cmd/api/service"
)

// HTMLHandler handles HTML template requests.
type HTMLHandler struct {
	service *service.PricingService
}

// NewHTMLHandler creates a new HTMLHandler instance.
func NewHTMLHandler(service *service.PricingService) *HTMLHandler {
	return &HTMLHandler{service: service}
}

// RootHandler handles the root page request.
// @Summary      Home page
// @Description  Display the home page with region selector
// @Tags         html
// @Produce      html
// @Success      200  {string}  string
// @Router       / [get]
func (h *HTMLHandler) RootHandler(c echo.Context) error {
	regions, err := h.service.GetAllRegions()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to query regions: "+err.Error())
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Regions": regions,
	})
}

// RegionHandler handles the region page request.
// @Summary      Region machines page
// @Description  Display machine types for a specific region
// @Tags         html
// @Produce      html
// @Param        region_name  query     string  true  "Region name"
// @Success      200  {string}  string
// @Router       /region [get]
func (h *HTMLHandler) RegionHandler(c echo.Context) error {
	regionName := c.QueryParam("region_name")

	if regionName == "" {
		return c.HTML(http.StatusOK, "")
	}

	machines, err := h.service.GetMachinesByRegion(regionName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to query machines: "+err.Error())
	}

	return c.Render(http.StatusOK, "machines.html", map[string]interface{}{
		"ListTitle": fmt.Sprintf("Machine types in %s", regionName),
		"Machines":  machines,
	})
}

// ComputeHandler handles the compute details page request.
// @Summary      Machine history page
// @Description  Display price history for a specific machine type
// @Tags         html
// @Produce      html
// @Param        region_name   query     string  true  "Region name"
// @Param        machine_type  query     string  true  "Machine type"
// @Success      200  {string}  string
// @Router       /compute [get]
func (h *HTMLHandler) ComputeHandler(c echo.Context) error {
	regionName := c.QueryParam("region_name")
	machineType := c.QueryParam("machine_type")

	if regionName == "" || machineType == "" {
		return c.HTML(http.StatusOK, "")
	}

	machineData, err := h.service.GetMachineDetail(regionName, machineType)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to query compute history: "+err.Error())
	}

	return c.Render(http.StatusOK, "prices.html", machineData)
}
