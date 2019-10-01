package server

import (
	"github.com/labstack/echo"
)

// hasPermission
func hasPermission(c echo.Context, permissionType string, read, write, manage, admin int) bool {
	reqUser := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT * FROM permissions WHERE user_id=$1 AND type=$2
	`, reqUser.ID, permissionType)

	if row == nil {
		return false
	}

	var permission Permission
	err := row.StructScan(&permission)
	if err != nil {
		return false
	}

	if permission.Read != read && permission.Read < read {
		return false
	}
	if permission.Write != write && permission.Write < write {
		return false
	}
	if permission.Manage != manage && permission.Manage < manage {
		return false
	}
	if permission.Admin != admin && permission.Admin < admin {
		return false
	}
	return true
}
