package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/oklog/ulid/v2"
)

type addGatewayReq struct {
	ID    string
	Model string `default0:"custom"`
}

type linkGatewayReq struct {
	ID     string
	User   string
	HomeID string
}

// CheckUlid check if id is a real ulid
func checkUlid(id string) error {
	_, err := ulid.ParseStrict(id)

	return err
}

// AddGateway o
func AddGateway(c echo.Context) error {
	req := new(addGatewayReq)
	err := c.Bind(req)
	if err != nil {
		return err
	} else if req.ID == "" {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Empty ID",
		})
	} else if checkUlid(req.ID) != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Not an ULID",
		})
	}

	var gateway Gateway
	err = DB.Get(&gateway, "SELECT id FROM gateways WHERE id=$1", req.ID)
	fmt.Println(err)
	if err == nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "ID already exists",
		})
	}

	newGateway := Gateway{
		ID:        req.ID,
		Model:     req.Model,
		CreatedAt: time.Now().Format(time.RFC1123),
	}
	DB.NamedExec("INSERT INTO gateways (id, model, created_at) VALUES (:id, :model, :created_at)", newGateway)

	fmt.Println(req.ID)

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Gateway created",
	})
}

// LinkGateway o
func LinkGateway(c echo.Context) error {
	req := new(linkGatewayReq)
	err := c.Bind(req)
	if err != nil {
		return err
	} else if req.ID == "" || req.User == "" || req.HomeID == "" {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Empty values",
		})
	}

	var gateway Gateway
	err = DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1 AND creator_id IS NULL AND home_id IS NULL", req.ID)
	if err == nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Gateway already linked",
		})
	}

	DB.NamedExec("UPDATE gateways SET creator_id=:User, home_id=:HomeID WHERE id=:ID", req)

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Gateway linked",
	})
}
