package server

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/ItsJimi/casa/logger"
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
		logger.WithFields(logger.Fields{"code": "CSRAR001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSRAR001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSRAR002"}).Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSRAR002",
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	row, err := DB.Query("INSERT INTO rooms (id, name, home_id, creator_id) VALUES (generate_ulid(), $1, $2, $3) RETURNING id;", req.Name, c.Param("homeId"), user.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRAR003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSRAR003",
			Message: "Room can't be created",
		})
	}
	var roomID string
	row.Next()
	err = row.Scan(&roomID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRAR004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSRAR004",
			Message: "Room can't be created",
		})
	}

	newPermission := Permission{
		UserID: user.ID,
		Type:   "room",
		TypeID: roomID,
		Read:   true,
		Write:  true,
		Manage: true,
		Admin:  true,
	}
	_, err = DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin) VALUES (generate_ulid(), :user_id, :type, :type_id, :read, :write, :manage, :admin)", newPermission)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRAR005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSRAR005",
			Message: "Room can't be created",
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
		logger.WithFields(logger.Fields{"code": "CSRUR001"}).Errorf("%s", err.Error())
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSRUR002"}).Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSRUR002",
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "room", c.Param("roomId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRUR003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSRUR003",
			Message: "Room not found",
		})
	}

	if permission.Manage == false && permission.Admin == false {
		logger.WithFields(logger.Fields{"code": "CSRUR004"}).Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSRUR004",
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE rooms SET Name=$1 WHERE id=$2", req.Name, c.Param("roomId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRUR005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSRUR005",
			Message: "Room can't be updated",
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
		logger.WithFields(logger.Fields{"code": "CSRDR001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSRDR001",
			Message: "Room not found",
		})
	}

	if permission.Admin == false {
		logger.WithFields(logger.Fields{"code": "CSRDR002"}).Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSRDR002",
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM rooms WHERE id=$1", c.Param("roomId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRDR003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSRDR003",
			Message: "Room can't be deleted",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "room", c.Param("roomId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRDR004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSRDR004",
			Message: "Room can't be deleted",
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
	RoomUpdatedAt string `db:"r_updatedat"`
}

type roomRes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	HomeID    string `json:"homeId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Creator   User   `json:"creator"`
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Manage    bool   `json:"manage"`
	Admin     bool   `json:"admin"`
}

// GetRooms route get list of user rooms
func GetRooms(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT permissions.*, users.*,
		rooms.id as r_id,	rooms.name AS r_name, rooms.home_id AS r_homeid, rooms.created_at AS r_createdat, rooms.updated_at AS r_updatedat FROM permissions
		JOIN rooms ON permissions.type_id = rooms.id
		JOIN users ON rooms.creator_id = users.id
		WHERE permissions.type=$1 AND permissions.user_id=$2 AND rooms.home_id=$3 AND (permissions.read=true OR permissions.admin=true)
	`, "room", user.ID, c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRGRS001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSRGRS001",
			Message: "Rooms can't be retrieved",
		})
	}

	rooms := []roomRes{}
	for rows.Next() {
		var permission permissionRoom
		err := rows.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSRGRS002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSRGRS002",
				Message: "Rooms can't be retrieved",
			})
		}

		rooms = append(rooms, roomRes{
			ID:        permission.RoomID,
			Name:      permission.RoomName,
			HomeID:    permission.RoomHomeID,
			CreatedAt: permission.RoomCreatedAt,
			UpdatedAt: permission.RoomUpdatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		})
	}

	totalRooms := strconv.Itoa(len(rooms))
	c.Response().Header().Set("Content-Range", "0-"+totalRooms+"/"+totalRooms)
	return c.JSON(http.StatusOK, rooms)
}

// GetRoom route get specific room with id
func GetRoom(c echo.Context) error {
	user := c.Get("user").(User)

	var permission permissionRoom
	err := DB.QueryRowx(`
		SELECT permissions.*, users.*,
		rooms.id as r_id,	rooms.name AS r_name, rooms.home_id AS r_homeid, rooms.created_at AS r_createdat, rooms.updated_at AS r_updatedat FROM permissions
		JOIN rooms ON permissions.type_id = rooms.id
		JOIN users ON rooms.creator_id = users.id
		WHERE permissions.type=$1 AND permissions.type_id=$2 AND permissions.user_id=$3
	`, "room", c.Param("roomId"), user.ID).StructScan(&permission)

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSRGR001"}).Errorf("QueryRowx: Select error")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSRGR001",
			Message: "Room can't be found",
		})
	}

	return c.JSON(http.StatusOK, roomRes{
		ID:        permission.RoomID,
		Name:      permission.RoomName,
		HomeID:    permission.RoomHomeID,
		CreatedAt: permission.RoomCreatedAt,
		UpdatedAt: permission.RoomUpdatedAt,
		Creator:   permission.User,
		Read:      permission.Permission.Read,
		Write:     permission.Permission.Write,
		Manage:    permission.Permission.Manage,
		Admin:     permission.Permission.Admin,
	})
}
