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

	// Link Gateway
	v1.POST("/gateway/link", LinkGateway)
	v1.GET("/gateway/sync/:id", SyncGateway)

	// Check authorization
	v1.Use(middleware.KeyAuth(IsAuthenticated))

	// Gateway
	v1.POST("/gateway", AddGateway)
	v1.PUT("/gateway/:id", UpdateGateway)
	v1.DELETE("/gateway/:id", DeleteGateway)
	v1.GET("/gateway/:id", GetGateway)

	// Homes
	v1.POST("/homes", AddHome)
	v1.PUT("/homes/:homeId", UpdateHome, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 1, 0)
	})
	v1.DELETE("/homes/:homeId", DeleteHome, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 0, 1)
	})
	v1.GET("/homes", GetHomes)
	v1.GET("/homes/:homeId", GetHome, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 1, 0, 0, 0)
	})

	// Members
	v1.GET("/homes/:homeId/members", GetMembers, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 1, 0, 0, 0)
	})
	v1.POST("/homes/:homeId/members", AddMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 1, 0)
	})

	// Rooms
	v1.POST("/homes/:homeId/rooms", AddRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 1, 0)
	})
	v1.PUT("/homes/:homeId/rooms/:roomId", UpdateRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 0, 0, 1, 0)
	})
	v1.DELETE("/homes/:homeId/rooms/:roomId", DeleteRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 0, 0, 0, 1)
	})
	v1.GET("/homes/:homeId/rooms", GetRooms, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 1, 0, 0, 0)
	})
	v1.GET("/homes/:homeId/rooms/:roomId", GetRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 1, 0, 0, 0)
	})

	// Devices
	v1.POST("/homes/:homeId/rooms/:roomId/devices", AddDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 0, 0, 1, 0)
	})
	v1.PUT("/homes/:homeId/rooms/:roomId/devices/:deviceId", UpdateDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "device", 0, 0, 1, 0)
	})
	v1.DELETE("/homes/:homeId/rooms/:roomId/devices/:deviceId", DeleteDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "device", 0, 0, 0, 1)
	})
	v1.GET("/homes/:homeId/rooms/:roomId/devices", GetDevices, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 1, 0, 0, 0)
	})
	v1.GET("/homes/:homeId/rooms/:roomId/devices/:deviceId", GetDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "device", 1, 0, 0, 0)
	})

	// Automations
	v1.POST("/homes/:homeId/automations", AddAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 1, 0, 0)
	})
	// TODO: Do Update
	// v1.PUT("/homes/:homeId/automations/:automationId", UpdateAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
	// 	return hasPermission(next, "home", 0, 1, 0, 0)
	// })
	v1.DELETE("/homes/:homeId/automations/:automationId", DeleteAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 1, 0)
	})
	v1.GET("/homes/:homeId/automations", GetAutomations, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 1, 0, 0, 0)
	})
	v1.GET("/homes/:homeId/automations/:automationId", GetAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 1, 0, 0, 0)
	})

	e.Logger.Fatal(e.Start(":" + port))
}
