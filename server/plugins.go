package server

import (
	"net/http"

	"github.com/labstack/echo"
)

// GetPlugins route get list of home plugins
func GetPlugins(c echo.Context) error {
	// TODO:
	return c.JSON(http.StatusOK, MessageResponse{})
}
