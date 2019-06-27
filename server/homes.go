package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

type addHomeReq struct {
	Name    string
	Address string
}

// AddHome route create and add user to an home
func AddHome(c echo.Context) error {
	req := new(addHomeReq)
	if err := c.Bind(req); err != nil {
		return err
	}
	var missingFields []string
	if req.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if req.Address == "" {
		missingFields = append(missingFields, "address")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	user := c.Get("user").(User)

	homeID := NewULID().String()
	newHome := Home{
		ID:        homeID,
		Name:      req.Name,
		Address:   req.Address,
		CreatedAt: time.Now().Format(time.RFC1123),
		CreatorID: user.ID,
	}
	DB.NamedExec("INSERT INTO homes (id, name, address, created_at, creator_id) VALUES (:id, :name, :address, :created_at, :creator_id)", newHome)

	permissionID := NewULID().String()
	newPermission := Permission{
		ID:        permissionID,
		UserID:    user.ID,
		Type:      "home",
		TypeID:    homeID,
		Read:      1,
		Write:     1,
		Manage:    1,
		Admin:     1,
		UpdatedAt: time.Now().Format(time.RFC1123),
	}
	DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin, updated_at) VALUES (:id, :user_id, :type, :type_id, :read, :write, :manage, :admin, :updated_at)", newPermission)

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: homeID,
	})
}

type permissionHome struct {
	Permission
	Home
}

// GetHomes route get list of user homes
func GetHomes(c echo.Context) error {
	user := c.Get("user").(User)
	var permissions []permissionHome

	rows, err := DB.Queryx(`
		SELECT * FROM permissions
		JOIN homes ON permissions.type_id = homes.id
		WHERE type=$1 AND user_id=$2
	`, "home", user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve homes",
		})
	}
	for rows.Next() {
		var permission permissionHome
		err = rows.StructScan(&permission)
		permissions = append(permissions, permission)
	}

	fmt.Println(permissions[0])

	return c.JSON(http.StatusInternalServerError, DataReponse{
		Data: "homes",
	})
}
