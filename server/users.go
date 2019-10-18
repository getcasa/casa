package server

import (
	"errors"
	"net/http"

	"github.com/ItsJimi/casa/logger"
	"github.com/labstack/echo"
)

// GetUser route get user by id
func GetUser(c echo.Context) error {
	reqUser := c.Get("user").(User)

	if c.Param("userId") == "me" || c.Param("userId") == reqUser.ID {
		return c.JSON(http.StatusOK, reqUser)
	}

	err := errors.New("Wrong parameters")
	contextLogger := logger.WithFields(logger.Fields{"code": "CSUGU001", "userId": c.Param("userId")})
	contextLogger.Warnf("%s", err.Error())

	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:  "CSUGU001",
		Error: err.Error(),
	})
}
