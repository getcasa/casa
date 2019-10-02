package server

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

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
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email", "Password", "PasswordConfirmation", "Firstname"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	if req.Password != req.PasswordConfirmation {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Passwords mismatch",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error password encryption",
		})
	}

	newUser := User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Birthdate: req.Birthdate,
	}
	_, err = DB.NamedExec("INSERT INTO users (id, email, password, firstname, lastname, birthdate) VALUES (generate_ulid(), :email, :password, :firstname, :lastname, :birthdate)", newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Can't add new user: " + err.Error(),
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
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email", "Password"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	var user User
	err := DB.Get(&user, "SELECT * FROM users WHERE email=$1", req.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Email and password doesn't match",
		})
	}

	row, err := DB.Query("INSERT INTO tokens (id, user_id, type, ip, user_agent) VALUES (generate_ulid(), $1, $2, $3, $4) RETURNING id;",
		user.ID, "signin", c.RealIP(), c.Request().UserAgent())
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 1: Token can't be create",
		})
	}
	var id string
	row.Next()
	err = row.Scan(&id)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 2: Token can't be create",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: id,
	})
}

type tokenUser struct {
	Token
	User
}

// IsAuthenticated verify validity of token
func IsAuthenticated(key string, c echo.Context) (bool, error) {
	var token tokenUser
	err := DB.Get(&token, "SELECT * FROM tokens JOIN users ON tokens.user_id = users.id WHERE tokens.id=$1", key)
	if err != nil {
		return false, nil
	}

	expireAt, err := time.Parse(time.RFC3339, token.Token.ExpireAt)
	if err != nil {
		return false, nil
	}
	if expireAt.Sub(time.Now()) <= 0 {
		return false, nil
	}

	token.User.Password = ""
	c.Set("user", token.User)

	return true, nil
}
