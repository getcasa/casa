package server

import (
	"net/http"
	"reflect"

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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRAR001"})
		contextLogger.Errorf("%s", err.Error())
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRAR002"})
		contextLogger.Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSRAR002",
			Error: err.Error(),
		})
	}

	user := c.Get("user").(User)

	row, err := DB.Query("INSERT INTO rooms (id, name, home_id, creator_id) VALUES (generate_ulid(), $1, $2, $3) RETURNING id;", req.Name, c.Param("homeId"), user.ID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRAR003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSRAR003",
			Error: "Room can't be created",
		})
	}
	var roomID string
	row.Next()
	err = row.Scan(&roomID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRAR004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSRAR004",
			Error: "Room can't be created",
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
	_, err = DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin) VALUES (generate_ulid(), :user_id, :type, :type_id, :read, :write, :manage, :admin)", newPermission)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRAR005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSRAR005",
			Error: "Room can't be created",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRUR001"})
		contextLogger.Errorf("%s", err.Error())
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRUR002"})
		contextLogger.Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSRUR002",
			Error: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "room", c.Param("roomId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRUR003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSRUR003",
			Error: "Room not found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRUR004"})
		contextLogger.Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSRUR004",
			Error: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE rooms SET Name=$1 WHERE id=$2", req.Name, c.Param("roomId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRUR005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSRUR005",
			Error: "Room can't be updated",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRDR001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSRDR001",
			Error: "Room not found",
		})
	}

	if permission.Admin == 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRDR002"})
		contextLogger.Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSRDR002",
			Error: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM rooms WHERE id=$1", c.Param("roomId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRDR003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSRDR003",
			Error: "Room can't be deleted",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "room", c.Param("roomId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRDR004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSRDR004",
			Error: "Room can't be deleted",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRGRS001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSRGRS001",
			Error: "Rooms can't be retrieved",
		})
	}

	var rooms []roomRes
	for rows.Next() {
		var permission permissionRoom
		err := rows.StructScan(&permission)
		if err != nil {
			contextLogger := logger.WithFields(logger.Fields{"code": "CSRGRS002"})
			contextLogger.Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:  "CSRGRS002",
				Error: "Rooms can't be retrieved",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRGR001"})
		contextLogger.Errorf("QueryRowx: Select error")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSRGR001",
			Error: "Room can't be found",
		})
	}

	var permission permissionRoom
	err := row.StructScan(&permission)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSRGR002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSRGR002",
			Error: "Room can't be found",
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
