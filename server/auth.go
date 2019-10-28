package server

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

var emailRegExp = "(?:[a-z0-9!#$%&'*+=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])"

type signupReq struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
	Firstname            string `json:"firstname"`
	Lastname             string `json:"lastname"`
	Birthdate            string `json:"birthdate"`
}

type signinReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUp route create an user
func SignUp(c echo.Context) error {
	req := new(signupReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASU001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASU001",
			Error: "Wrong parameters",
		})
	}

	if req.Password != req.PasswordConfirmation {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASU002"})
		contextLogger.Warnf("Passwords mismatch")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASU002",
			Error: "Passwords mismatch",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASU003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSASU003",
			Error: "Password can't be encrypted",
		})
	}

	birthdate, err := time.Parse(time.RFC3339, req.Birthdate)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASU004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSASU004",
			Error: "Birthdate can't be parsed",
		})
	}

	newUser := User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Birthdate: birthdate.Format("2006-01-02 00:00:00"),
	}
	_, err = DB.NamedExec("INSERT INTO users (id, email, password, firstname, lastname, birthdate) VALUES (generate_ulid(), :email, :password, :firstname, :lastname, :birthdate)", newUser)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASU005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSASU005",
			Error: "Account can't be created",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Account created",
	})
}

// SignIn route log an user by giving token
func SignIn(c echo.Context) error {
	req := new(signinReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASI001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASI001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email", "Password"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASI002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASI002",
			Error: err.Error(),
		})
	}

	var user User
	err := DB.Get(&user, "SELECT id, password FROM users WHERE email=$1", req.Email)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASI003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASI003",
			Error: "Email and password doesn't match",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASI004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASI003",
			Error: "Email and password doesn't match",
		})
	}

	row, err := DB.Query("INSERT INTO tokens (id, user_id, type, ip, user_agent) VALUES (generate_ulid(), $1, $2, $3, $4) RETURNING id;",
		user.ID, "signin", c.RealIP(), c.Request().UserAgent())
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASI005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASI005",
			Error: "Token can't be created",
		})
	}
	var id string
	row.Next()
	err = row.Scan(&id)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASI006"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSASI006",
			Error: "Token can't be created",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: id,
	})
}

// SignOut route logout user and delete his token
func SignOut(c echo.Context) error {
	token := strings.Split(c.Request().Header.Get("Authorization"), " ")[1]
	_, err := DB.Exec(`
		DELETE FROM tokens
		WHERE id=$1
	`, token)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSASO001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSASO001",
			Error: "Token can't be delete",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "You've been disconnected and your token has been deleted",
	})
}

type tokenUser struct {
	Token
	User
}

// IsAuthenticated verify validity of token
func IsAuthenticated(key string, c echo.Context) (bool, error) {
	var token tokenUser
	err := DB.Get(&token, "SELECT users.*, tokens.expire_at FROM tokens JOIN users ON tokens.user_id = users.id WHERE tokens.id=$1", key)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSAIA001"})
		contextLogger.Errorf("%s", err.Error())
		return false, nil
	}

	expireAt, err := time.Parse(time.RFC3339, token.Token.ExpireAt)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSAIA002"})
		contextLogger.Errorf("%s", err.Error())
		return false, nil
	}
	if expireAt.Sub(time.Now()) <= 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSAIA003"})
		contextLogger.Warnf("Expired tokens")
		return false, nil
	}

	c.Set("user", token.User)

	return true, nil
}
