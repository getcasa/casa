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
	v1.GET("/gateways/discover/:plugin", GetDiscoveredDevices)
	v1.GET("/gateways/:id", GetGateway)
<<<<<<< HEAD
=======
	v1.GET("/gateways/:gatewayId/discover", GetDiscoveredDevices)
	v1.POST("/gateways/:gatewayId/actions", CallAction) // TODO: Add permissions to CallAction
>>>>>>> add route POST /gateways/:gatewayId/actions to call an action on gateway

	// Homes
	v1.POST("/homes", AddHome)
	v1.PUT("/homes/:homeId", UpdateHome, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, true, false)
	})
	v1.DELETE("/homes/:homeId", DeleteHome, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, false, true)
	})
	v1.GET("/homes", GetHomes)
	v1.GET("/homes/:homeId", GetHome, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", true, false, false, false)
	})

	// Homes Members
	v1.GET("/homes/:homeId/members", GetMembers, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", true, false, false, false)
	})
	v1.POST("/homes/:homeId/members", AddMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, true, false)
	})
	v1.DELETE("/homes/:homeId/members/:userId", RemoveMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, true, false)
	})
	v1.PUT("/homes/:homeId/members/:userId", EditMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, false, true)
	})

	// Rooms
	v1.POST("/homes/:homeId/rooms", AddRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, true, false)
	})
	v1.PUT("/homes/:homeId/rooms/:roomId", UpdateRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", false, false, true, false)
	})
	v1.DELETE("/homes/:homeId/rooms/:roomId", DeleteRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", false, false, false, true)
	})
	v1.GET("/homes/:homeId/rooms", GetRooms, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", true, false, false, false)
	})
	v1.GET("/homes/:homeId/rooms/:roomId", GetRoom, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", true, false, false, false)
	})

	// Rooms Members
	v1.GET("/homes/:homeId/rooms/:roomId/members", GetRoomMembers, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", true, false, false, false)
	})
	v1.PUT("/homes/:homeId/rooms/:roomId/members/:userId", EditRoomMember, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", false, false, false, true)
	})

	// Devices
	v1.POST("/homes/:homeId/rooms/:roomId/devices", AddDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", false, false, true, false)
	})
	v1.PUT("/homes/:homeId/rooms/:roomId/devices/:deviceId", UpdateDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "device", false, false, true, false)
	})
	v1.DELETE("/homes/:homeId/rooms/:roomId/devices/:deviceId", DeleteDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "device", false, false, false, true)
	})
	v1.GET("/homes/:homeId/rooms/:roomId/devices", GetDevices, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "room", true, false, false, false)
	})
	v1.GET("/homes/:homeId/rooms/:roomId/devices/:deviceId", GetDevice, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "device", true, false, false, false)
	})

	// Automations
	v1.POST("/homes/:homeId/automations", AddAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, true, false, false)
	})
	// TODO: Do Update
	// v1.PUT("/homes/:homeId/automations/:automationId", UpdateAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
	// 	return hasPermission(next, "home", false, true, false, false)
	// })
	v1.DELETE("/homes/:homeId/automations/:automationId", DeleteAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", false, false, true, false)
	})
	v1.GET("/homes/:homeId/automations", GetAutomations, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", true, false, false, false)
	})
	v1.GET("/homes/:homeId/automations/:automationId", GetAutomation, func(next echo.HandlerFunc) echo.HandlerFunc {
		return hasPermission(next, "home", true, false, false, false)
	})

	// Users
	v1.GET("/users/:userId", GetUser)
	v1.PUT("/users/:userId", UpdateUserProfil)
	v1.PUT("/users/:userId/email", UpdateUserEmail)
	v1.PUT("/users/:userId/password", UpdateUserPassword)

	e.Logger.Fatal(e.Start(":" + port))
}
