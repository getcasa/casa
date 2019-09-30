package server

import (
	"fmt"

	"github.com/labstack/echo"
)

// hasAdminPermission
func hasAdminPermission(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("test")
		return next(c)
	}
}
