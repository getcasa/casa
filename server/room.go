package server

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
)

type addRoomReq struct {
	Name string
}

// AddRoom route create and add user to an room
func AddRoom(c echo.Context) error {
	req := new(addRoomReq)
	if err := c.Bind(req); err != nil {
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	row, err := DB.Query("INSERT INTO rooms (id, name, home_id, created_at, creator_id) VALUES (generate_ulid(), :name, :home_id, :creator_id) RETURNING id;", req.Name, c.Param("homeId"), user.ID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 1: Token can't be create",
		})
	}
	var roomID string
	row.Next()
	err = row.Scan(&roomID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 2: Token can't be create",
		})
	}

	newPermission := Permission{
		UserID: user.ID,
		Type:   "room",
		TypeID: roomID,
		Read:   1,
		Write:  1,
		Manage: 1,
		Admin:  1,
	}
	_, err = DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin, updated_at) VALUES (generate_ulid(), :user_id, :type, :type_id, :read, :write, :manage, :admin)", newPermission)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Can't add new permission: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: roomID,
	})
}

// UpdateRoom route update room
func UpdateRoom(c echo.Context) error {
	req := new(addRoomReq)
	if err := c.Bind(req); err != nil {
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "room", c.Param("roomId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Room not found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE rooms SET Name=$1 WHERE id=$2", req.Name, c.Param("roomId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't update room",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Room updated",
	})
}

// DeleteRoom route delete room
func DeleteRoom(c echo.Context) error {
	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "room", c.Param("roomId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Room not found",
		})
	}

	if permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM rooms WHERE id=$1", c.Param("roomId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 6: Can't delete room",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "room", c.Param("roomId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 7: Can't delete room",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Room deleted",
	})
}

type permissionRoom struct {
	Permission
	User
	RoomID        string `db:"r_id"`
	RoomName      string `db:"r_name"`
	RoomHomeID    string `db:"r_homeid"`
	RoomCreatedAt string `db:"r_createdat"`
}

type roomRes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	HomeID    string `json:"home_id"`
	CreatedAt string `json:"created_at"`
	Creator   User   `json:"creator"`
	Read      int    `json:"read"`
	Write     int    `json:"write"`
	Manage    int    `json:"manage"`
	Admin     int    `json:"admin"`
}

// GetRooms route get list of user rooms
func GetRooms(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT permissions.*, users.*,
		rooms.id as r_id,	rooms.name AS r_name, rooms.home_id AS r_homeid, rooms.created_at AS r_createdat FROM permissions
		JOIN rooms ON permissions.type_id = rooms.id
		JOIN users ON rooms.creator_id = users.id
		WHERE type=$1 AND user_id=$2 AND rooms.home_id=$3 AND (permissions.read=1 OR permissions.admin=1)
	`, "room", user.ID, c.Param("homeId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve rooms",
		})
	}

	var rooms []roomRes
	for rows.Next() {
		var permission permissionRoom
		err := rows.StructScan(&permission)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Error 3: Can't retrieve rooms",
			})
		}
		rooms = append(rooms, roomRes{
			ID:        permission.RoomID,
			Name:      permission.RoomName,
			HomeID:    permission.RoomHomeID,
			CreatedAt: permission.RoomCreatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: rooms,
	})
}

// GetRoom route get specific room with id
func GetRoom(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT permissions.*, users.*,
		rooms.id as r_id,	rooms.name AS r_name, rooms.home_id AS r_homeid, rooms.created_at AS r_createdat FROM permissions
		JOIN rooms ON permissions.type_id = rooms.id
		JOIN users ON rooms.creator_id = users.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "room", c.Param("roomId"), user.ID)

	if row == nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Room not found",
		})
	}

	var permission permissionRoom
	err := row.StructScan(&permission)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 4: Can't retrieve rooms",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: roomRes{
			ID:        permission.RoomID,
			Name:      permission.RoomName,
			HomeID:    permission.RoomHomeID,
			CreatedAt: permission.RoomCreatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		},
	})
}
