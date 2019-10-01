package server

import (
	"net/http"

	"github.com/labstack/echo"
)

// GetUser route get user by id
func GetUser(c echo.Context) error {
	reqUser := c.Get("user").(User)

	if c.Param("userId") == "me" || c.Param("userId") == reqUser.ID {
		return c.JSON(http.StatusOK, DataReponse{
			Data: reqUser,
		})
	}

	return c.JSON(http.StatusBadRequest, MessageResponse{
		Message: "Wrong parameters",
	})
}
