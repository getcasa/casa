package server

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

// GetUser route get user by id
func GetUser(c echo.Context) error {
	reqUser := c.Get("user").(User)

	if c.Param("userId") == "me" || c.Param("userId") == reqUser.ID {
		return c.JSON(http.StatusOK, reqUser)
	}

	err := errors.New("Wrong parameters")
	logger.WithFields(logger.Fields{"code": "CSUGU001", "userId": c.Param("userId")}).Warnf("%s", err.Error())

	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    "CSUGU001",
		Message: err.Error(),
	})
}

// TODO: Add Birthdate field
type updateUserProfilReq struct {
	Firstname string
	Lastname  string
}

// UpdateUserProfil route update user profil
func UpdateUserProfil(c echo.Context) error {
	req := new(updateUserProfilReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUP001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUP001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Firstname", "Lastname"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUP002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUP002",
			Message: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		logger.WithFields(logger.Fields{"code": "CSUUUP003"}).Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSUUUP003",
			Message: "Unauthorized",
		})
	}

	_, err := DB.Exec(`
		UPDATE users
		SET firstname=$1, lastname=$2
		WHERE id=$3
	`, req.Firstname, req.Lastname, c.Param("userId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUP004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSUUUP004",
			Message: "User profil can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User profil has been updated",
	})
}

type updateUserEmailReq struct {
	Email string
}

// UpdateUserEmail route update user email
func UpdateUserEmail(c echo.Context) error {
	req := new(updateUserEmailReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUE001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUE001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUE002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUE002",
			Message: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		logger.WithFields(logger.Fields{"code": "CSUUUE003"}).Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSUUUE003",
			Message: "Unauthorized",
		})
	}

	_, err := DB.Exec(`
		UPDATE users
		SET email=$1
		WHERE id=$2
	`, req.Email, c.Param("userId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUE004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSUUUE004",
			Message: "User email can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User email has been updated",
	})
}

type updateUserPasswordReq struct {
	Password                string
	NewPassword             string
	NewPasswordConfirmation string
}

// UpdateUserPassword route update user password
func UpdateUserPassword(c echo.Context) error {
	req := new(updateUserPasswordReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUPA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUPA001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Password", "NewPassword", "NewPasswordConfirmation"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUPA002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUPA002",
			Message: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		logger.WithFields(logger.Fields{"code": "CSUUUPA003"}).Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSUUUPA003",
			Message: "Unauthorized",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(reqUser.Password), []byte(req.Password)); err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUPA004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUPA004",
			Message: "Password doesn't match",
		})
	}

	if req.NewPassword != req.NewPasswordConfirmation {
		logger.WithFields(logger.Fields{"code": "CSUUUPA005"}).Warnf("New passwords mismatch")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSUUUPA005",
			Message: "New passwords mismatch",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUPA006"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSUUUPA006",
			Message: "New password can't be encrypted",
		})
	}

	_, err = DB.Exec(`
		UPDATE users
		SET password=$1
		WHERE id=$2
	`, hashedPassword, c.Param("userId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSUUUPA007"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSUUUPA007",
			Message: "User password can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User password has been updated",
	})
}
