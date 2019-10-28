package server

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
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

// TODO: Add Birthdate field
type updateUserReq struct {
	Firstname string
	Lastname  string
}

// UpdateUser route update user profil
func UpdateUser(c echo.Context) error {
	req := new(updateUserReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUU001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUU001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Firstname", "Lastname"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUU002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUU002",
			Error: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUU003"})
		contextLogger.Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSUUU003",
			Error: "Unauthorized",
		})
	}

	_, err := DB.Exec(`
		UPDATE users
		SET firstname=$1, lastname=$2
		WHERE id=$3
	`, req.Firstname, req.Lastname, c.Param("userId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUU004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSUUU004",
			Error: "User can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User has been updated",
	})
}
