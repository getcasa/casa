package server

import (
	"net/http"

	"github.com/labstack/echo"
)

// AddHome route create and add user to an home
func AddHome(c echo.Context) error {
	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Home created",
	})
}
