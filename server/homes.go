package server

import (
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

type homeRes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	CreatedAt string `json:"created_at"`
	Creator   User   `json:"creator"`
	Read      int    `json:"read"`
	Write     int    `json:"write"`
	Manage    int    `json:"manage"`
	Admin     int    `json:"admin"`
}

// GetHomes route get list of user homes
func GetHomes(c echo.Context) error {
	user := c.Get("user").(User)

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

	var homes []homeRes
	for rows.Next() {
		var permission permissionHome
		err := rows.StructScan(&permission)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Error 3: Can't retrieve homes",
			})
		}
		homes = append(homes, homeRes{
			ID:        permission.Home.ID,
			Name:      permission.Home.Name,
			Address:   permission.Home.Address,
			CreatedAt: permission.Home.CreatedAt,
			Creator:   user,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		})
	}

	return c.JSON(http.StatusInternalServerError, DataReponse{
		Data: homes,
	})
}

// GetHome route get specific home with id
func GetHome(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT * FROM permissions
		JOIN homes ON permissions.type_id = homes.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "home", c.Param("id"), user.ID)

	if row == nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Home not found",
		})
	}

	var permission permissionHome
	err := row.StructScan(&permission)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 4: Can't retrieve homes",
		})
	}

	return c.JSON(http.StatusInternalServerError, DataReponse{
		Data: homeRes{
			ID:        permission.Home.ID,
			Name:      permission.Home.Name,
			Address:   permission.Home.Address,
			CreatedAt: permission.Home.CreatedAt,
			Creator:   user,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		},
	})
}
