package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetDevices get all devices
func GetDevices(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
