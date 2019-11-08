package server

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// MessageResponse define json response for API
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse define json reponse error for API
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DataReponse define json response for API
type DataReponse struct {
	Data interface{} `json:"data"`
}

// Start start echo server
func Start(port string) {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, MessageResponse{
			Message: "Casa Server API",
		})
	})

	e.GET("/casa", func(c echo.Context) error {
		return c.JSON(http.StatusOK, MessageResponse{
			Message: "Hi",
		})
	})

	// V1
	v1 := e.Group("/v1")

	// Signup
	v1.POST("/signup", SignUp)

	// Signin
	v1.POST("/signin", SignIn)

	// Link Gateway
	v1.POST("/gateway", AddGateway)
	v1.POST("/gateway/:gatewayId/plugins", AddPlugin)
	v1.GET("/gateway/:gatewayId/plugins/:pluginName", GetPlugin)

	// WS
	v1.GET("/ws", InitConnection)

	// Check authorization
	v1.Use(middleware.KeyAuth(IsAuthenticated))

	// Signout
	v1.POST("/signout", SignOut)

	// Gateway
	v1.POST("/gateways/link", LinkGateway)
	v1.PUT("/gateways/:id", UpdateGateway)
	v1.DELETE("/gateways/:id", DeleteGateway)
	v1.GET("/gateways/:id", GetGateway)
	v1.GET("/gateways/:gatewayId/discover", GetDiscoveredDevices)

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

	// Homes Members
	v1.GET("/homes/:homeId/members", GetMembers, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 1, 0, 0, 0)
	})
	v1.POST("/homes/:homeId/members", AddMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 1, 0)
	})
	v1.DELETE("/homes/:homeId/members/:userId", RemoveMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 1, 0)
	})
	v1.PUT("/homes/:homeId/members/:userId", EditMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", 0, 0, 0, 1)
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

	// Rooms Members
	v1.GET("/homes/:homeId/rooms/:roomId/members", GetRoomMembers, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 1, 0, 0, 0)
	})
	v1.PUT("/homes/:homeId/rooms/:roomId/members/:userId", EditRoomMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", 0, 0, 0, 1)
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

	// Users
	v1.GET("/users/:userId", GetUser)
	v1.PUT("/users/:userId", UpdateUserProfil)
	v1.PUT("/users/:userId/email", UpdateUserEmail)
	v1.PUT("/users/:userId/password", UpdateUserPassword)

	e.Logger.Fatal(e.Start(":" + port))
}
