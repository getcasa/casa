package server

import (
	"net/http"

	"github.com/ItsJimi/casa/logger"
	"github.com/labstack/echo"
)

// hasPermission
func hasPermission(next echo.HandlerFunc, permissionType string, read, write, manage, admin bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqUser := c.Get("user").(User)

		row := DB.QueryRowx(`
			SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3
		`, reqUser.ID, permissionType, c.Param(permissionType+"Id"))

		if row == nil {
			logger.WithFields(logger.Fields{"code": "CSPHP001", "userId": reqUser.ID, "type": permissionType, "typeId": c.Param(permissionType + "Id")}).Warnf("Unauthorized")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "CSPHP001",
				Message: "Unauthorized",
			})
		}

		var permission Permission
		err := row.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSPHP002", "userId": reqUser.ID, "type": permissionType, "typeId": c.Param(permissionType + "Id")}).Errorf("%s", err.Error())
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "CSPHP002",
				Message: "Unauthorized",
			})
		}

		if read == true && permission.Read == false && permission.Admin != true {
			logger.WithFields(logger.Fields{"code": "CSPHP003", "userId": reqUser.ID, "type": permissionType, "typeId": c.Param(permissionType + "Id")}).Warnf("Unauthorized")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "CSPHP003",
				Message: "Unauthorized",
			})
		}
		if write == true && permission.Write == false && permission.Admin != true {
			logger.WithFields(logger.Fields{"code": "CSPHP004", "userId": reqUser.ID, "type": permissionType, "typeId": c.Param(permissionType + "Id")}).Warnf("Unauthorized")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "CSPHP004",
				Message: "Unauthorized",
			})
		}
		if manage == true && permission.Manage == false && permission.Admin != true {
			logger.WithFields(logger.Fields{"code": "CSPHP005", "userId": reqUser.ID, "type": permissionType, "typeId": c.Param(permissionType + "Id")}).Warnf("Unauthorized")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "CSPHP005",
				Message: "Unauthorized",
			})
		}
		if admin == true && permission.Admin == false {
			logger.WithFields(logger.Fields{"code": "CSPHP006", "userId": reqUser.ID, "type": permissionType, "typeId": c.Param(permissionType + "Id")}).Warnf("Unauthorized")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "CSPHP006",
				Message: "Unauthorized",
			})
		}

		return next(c)
	}
}
