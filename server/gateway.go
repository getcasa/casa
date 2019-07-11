package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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

type updateGatewayReq struct {
	Name string
}

// CheckUlid check if id is a real ulid
func checkUlid(id string) error {
	_, err := ulid.ParseStrict(id)

	return err
}

// AddGateway route add new gateway in system
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

// UpdateGateway route update gateway
func UpdateGateway(c echo.Context) error {
	id := c.Param("id")
	req := new(updateGatewayReq)
	if err := c.Bind(req); err != nil {
		return err
	}
	var missingFields []string
	if req.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET Name=$1 WHERE id=$3", req.Name, gateway.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't update Gateway",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Gateway updated",
	})
}

// DeleteGateway route delete gateway
func DeleteGateway(c echo.Context) error {
	id := c.Param("id")
	user := c.Get("user").(User)

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	if permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM gateway WHERE id=$1", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 6: Can't delete gateway",
		})
	}

	//TODO: delete device sync with gateway

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Home deleted",
	})
}

// LinkGateway route link gateway with user & home
func LinkGateway(c echo.Context) error {
	req := new(linkGatewayReq)
	err := c.Bind(req)
	if err != nil {
		return err
	}

	var missingFields []string
	if req.ID == "" {
		missingFields = append(missingFields, "id")
	}
	if req.User == "" {
		missingFields = append(missingFields, "user")
	}
	if req.HomeID == "" {
		missingFields = append(missingFields, "home_id")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	var gateway Gateway
	err = DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1 AND creator_id IS NULL AND home_id IS NULL", req.ID)
	if err == nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Gateway already linked",
		})
	}

	var user User
	err = DB.Get(&user, "SELECT * FROM users WHERE ID=$1", req.User)
	if err != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "User " + req.User + " not found",
		})
	}

	var home Home
	err = DB.Get(&home, "SELECT * FROM homes WHERE ID=$1", req.HomeID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Home " + req.HomeID + " not found",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET creator_id=$1, home_id=$2 WHERE id=$3", req.User, req.HomeID, req.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't link Gateway",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Gateway linked",
	})
}

// GetGateway route get specific gateway with id
func GetGateway(c echo.Context) error {
	// user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT * FROM gateways
		WHERE id=$1
	 `, c.Param("id"))

	if row == nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	var gateway Gateway
	err := row.StructScan(&gateway)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 4: Can't retrieve gateway",
		})
	}
	log.Println(gateway)

	if row == nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Home not found",
		})
	}

	// var permission permissionHome
	// err := row.StructScan(&permission)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, MessageResponse{
	// 		Message: "Error 4: Can't retrieve homes",
	// 	})
	// }

	// return c.JSON(http.StatusOK, DataReponse{
	// 	Data: homeRes{
	// 		ID:        permission.Home.ID,
	// 		Name:      permission.Home.Name,
	// 		Address:   permission.Home.Address,
	// 		CreatedAt: permission.Home.CreatedAt,
	// 		Creator:   user,
	// 		Read:      permission.Permission.Read,
	// 		Write:     permission.Permission.Write,
	// 		Manage:    permission.Permission.Manage,
	// 		Admin:     permission.Permission.Admin,
	// 	},
	// })

	return nil
}
