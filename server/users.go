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
	contextLogger := logger.WithFields(logger.Fields{"code": "CSUGU001", "userId": c.Param("userId")})
	contextLogger.Warnf("%s", err.Error())

	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:  "CSUGU001",
		Error: err.Error(),
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUP001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUP001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Firstname", "Lastname"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUP002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUP002",
			Error: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUP003"})
		contextLogger.Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSUUUP003",
			Error: "Unauthorized",
		})
	}

	_, err := DB.Exec(`
		UPDATE users
		SET firstname=$1, lastname=$2
		WHERE id=$3
	`, req.Firstname, req.Lastname, c.Param("userId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUP004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSUUUP004",
			Error: "User profil can't be updated",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUE001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUE001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUE002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUE002",
			Error: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUE003"})
		contextLogger.Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSUUUE003",
			Error: "Unauthorized",
		})
	}

	_, err := DB.Exec(`
		UPDATE users
		SET email=$1
		WHERE id=$2
	`, req.Email, c.Param("userId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUE004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSUUUE004",
			Error: "User email can't be updated",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUPA001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Password", "NewPassword", "NewPasswordConfirmation"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUPA002",
			Error: err.Error(),
		})
	}

	reqUser := c.Get("user").(User)

	if reqUser.ID != c.Param("userId") {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA003"})
		contextLogger.Errorf("%s", "Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSUUUPA003",
			Error: "Unauthorized",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(reqUser.Password), []byte(req.Password)); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUPA004",
			Error: "Password doesn't match",
		})
	}

	if req.NewPassword != req.NewPasswordConfirmation {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA005"})
		contextLogger.Warnf("New passwords mismatch")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSUUUPA005",
			Error: "New passwords mismatch",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA006"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSUUUPA006",
			Error: "New password can't be encrypted",
		})
	}

	_, err = DB.Exec(`
		UPDATE users
		SET password=$1
		WHERE id=$2
	`, hashedPassword, c.Param("userId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSUUUPA007"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSUUUPA007",
			Error: "User password can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User password has been updated",
	})
}
