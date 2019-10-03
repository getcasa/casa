package server

import (
	"net/http"

	"github.com/labstack/echo"
)

// hasPermission
func hasPermission(next echo.HandlerFunc, permissionType string, read, write, manage, admin int) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqUser := c.Get("user").(User)

		row := DB.QueryRowx(`
			SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3
		`, reqUser.ID, permissionType, c.Param(permissionType+"Id"))

		if row == nil {
			return c.JSON(http.StatusUnauthorized, MessageResponse{
				Message: "Unauthorized 0",
			})
		}

		var permission Permission
		err := row.StructScan(&permission)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, MessageResponse{
				Message: "Unauthorized 1",
			})
		}

		if permission.Read != read && permission.Read < read && permission.Admin != 1 {
			return c.JSON(http.StatusUnauthorized, MessageResponse{
				Message: "Unauthorized 2",
			})
		}
		if permission.Write != write && permission.Write < write && permission.Admin != 1 {
			return c.JSON(http.StatusUnauthorized, MessageResponse{
				Message: "Unauthorized 3",
			})
		}
		if permission.Manage != manage && permission.Manage < manage && permission.Admin != 1 {
			return c.JSON(http.StatusUnauthorized, MessageResponse{
				Message: "Unauthorized 4",
			})
		}
		if permission.Admin != admin && permission.Admin < admin {
			return c.JSON(http.StatusUnauthorized, MessageResponse{
				Message: "Unauthorized 5",
			})
		}

		return next(c)
	}
}
