package server

import (
	"net/http"
	"time"

	cryptorand "crypto/rand"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/oklog/ulid/v2"
)

// MessageResponse define json response for API
type MessageResponse struct {
	Message string `json:"message"`
}

// DataReponse define json response for API
type DataReponse struct {
	Data interface{} `json:"data"`
}

// NewULID create an ulid
func NewULID() ulid.ULID {
	id, _ := ulid.New(ulid.Timestamp(time.Now()), cryptorand.Reader)
	return id
}

// Start start echo server
func Start(port string) {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// V1
	v1 := e.Group("/v1")

	// Signup
	v1.POST("/signup", SignUp)

	// Signin
	v1.POST("/signin", SignIn)

	// Check authorization
	v1.Use(middleware.KeyAuth(IsAuthenticated))

	// Homes
	v1.POST("/homes", AddHome)
	v1.PUT("/homes/:id", UpdateHome)
	v1.DELETE("/homes/:id", DeleteHome)
	v1.GET("/homes", GetHomes)
	v1.GET("/homes/:id", GetHome)

	e.Logger.Fatal(e.Start(":" + port))
}
