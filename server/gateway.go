package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

type addGatewayReq struct {
	ID    string
	Model string `default0:"custom"`
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
