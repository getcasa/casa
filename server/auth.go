package server

import (
	"net/http"
	"regexp"
	"strings"
	"time"

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
		return err
	}
	var missingFields []string
	if req.Email == "" {
		missingFields = append(missingFields, "email")
	}
	if req.Password == "" {
		missingFields = append(missingFields, "password")
	}
	if req.PasswordConfirmation == "" {
		missingFields = append(missingFields, "passwordConfirmation")
	}
	if req.Firstname == "" {
		missingFields = append(missingFields, "firstname")
	}
	if req.Lastname == "" {
		missingFields = append(missingFields, "lastname")
	}
	if req.Birthdate == "" {
		missingFields = append(missingFields, "birthdate")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}
	isValid, _ := regexp.MatchString(emailRegExp, req.Email)
	if isValid == false {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Invalid email",
		})
	}
	if req.Password != req.PasswordConfirmation {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Passwords mismatch",
		})
	}
	var user User
	err := DB.Get(&user, "SELECT * FROM users WHERE email=$1", req.Email)
	if err == nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Email already used",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 1",
		})
	}

	newUser := User{
		ID:        NewULID().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Birthdate: req.Birthdate,
		CreatedAt: time.Now().Format(time.RFC1123),
	}
	DB.NamedExec("INSERT INTO users (id, email, password, firstname, lastname, birthdate, created_at) VALUES (:id, :email, :password, :firstname, :lastname, :birthdate, :created_at)", newUser)

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
	var missingFields []string
	if req.Email == "" {
		missingFields = append(missingFields, "email")
	}
	if req.Password == "" {
		missingFields = append(missingFields, "password")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	var user User
	err := DB.Get(&user, "SELECT * FROM users WHERE email=$1", req.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Email and password doesn't match",
		})
	}

	id := NewULID().String()
	createdAt := time.Now()

	newToken := Token{
		ID:        id,
		UserID:    user.ID,
		Type:      "signin",
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
		Read:      1,
		Write:     1,
		Manage:    1,
		Admin:     1,
		CreatedAt: createdAt.Format(time.RFC1123),
		ExpireAt:  createdAt.AddDate(0, 1, 0).Format(time.RFC1123),
	}
	DB.NamedExec("INSERT INTO tokens (id, user_id, type, ip, user_agent, read, write, manage, admin, created_at, expire_at) VALUES (:id, :user_id, :type, :ip, :user_agent, :read, :write, :manage, :admin, :created_at, :expire_at)", newToken)

	return c.JSON(http.StatusOK, MessageResponse{
		Message: id,
	})
}

// IsAuthenticated verify validity of token
func IsAuthenticated(key string, c echo.Context) (bool, error) {
	var token Token
	err := DB.Get(&token, "SELECT * FROM tokens WHERE id=$1", key)
	if err != nil {
		return false, nil
	}

	expireAt, err := time.Parse(time.RFC1123, token.ExpireAt)
	if err != nil {
		return false, nil
	}
	if expireAt.Sub(time.Now()) <= 0 {
		return false, nil
	}

	var user User
	err = DB.Get(&user, "SELECT * FROM users WHERE id=$1", token.UserID)
	if err != nil {
		return false, nil
	}
	c.Set("user", user)

	return true, nil
}
