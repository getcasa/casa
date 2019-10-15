package server

import (
	"database/sql"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
)

type addDeviceReq struct {
	GatewayID    string
	Name         string
	PhysicalID   string
	PhysicalName string
	RoomID       string
	Plugin       string
	Icon         string
}

// AddDevice route create a device
func AddDevice(c echo.Context) error {
	req := new(addDeviceReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDAD001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSDAD001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name", "GatewayID", "PhysicalID", "PhysicalName", "Plugin"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDAD002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSDAD002",
			Error: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var device Device
	err := DB.Get(&device, "SELECT * FROM devices WHERE physical_id=$1 AND gateway_id=$2", req.PhysicalID, req.GatewayID)
	if err == nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDAD003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDAD003",
			Error: "Device with the same physical id already exist in this gateway",
		})
	}

	row, err := DB.Query("INSERT INTO devices (id, name, icon, room_id, gateway_id, physical_id, physical_name, plugin, creator_id) VALUES (generate_ulid(), $1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;",
		req.Name, req.Icon, c.Param("roomId"), req.GatewayID, req.PhysicalID, req.PhysicalName, req.Plugin, user.ID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDAD004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSDAD004",
			Error: "Device can't be created",
		})
	}
	var deviceID string
	row.Next()
	err = row.Scan(&deviceID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDAD005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSDAD005",
			Error: "Device can't be created",
		})
	}

	newPermission := Permission{
		UserID: user.ID,
		Type:   "device",
		TypeID: deviceID,
		Read:   1,
		Write:  1,
		Manage: 1,
		Admin:  1,
	}
	_, err = DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin) VALUES (generate_ulid(), :user_id, :type, :type_id, :read, :write, :manage, :admin)", newPermission)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDAD006"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDAD006",
			Error: "Device can't be created",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: deviceID,
	})
}

// UpdateDevice route update device
func UpdateDevice(c echo.Context) error {
	req := new(addDeviceReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDUD001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSDUD001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name", "RoomID"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDUD002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSDUD002",
			Error: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "device", c.Param("deviceId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDUD003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSDUD003",
			Error: "Device can't be found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDUD004"})
		contextLogger.Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSDUD004",
			Error: "Unauthorized modifications",
		})
	}
	request := "UPDATE devices SET "
	if req.Name != "" {
		request += "Name='" + req.Name + "'"
		if req.RoomID != "" {
			request += ", room_id='" + req.RoomID + "'"
		}
	} else if req.RoomID != "" {
		request += "room_id='" + req.RoomID + "'"
	}
	request += " WHERE id=$1"
	_, err = DB.Exec(request, c.Param("deviceId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDUD005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDUD005",
			Error: "Device can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Device updated",
	})
}

// DeleteDevice route delete device
func DeleteDevice(c echo.Context) error {
	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "device", c.Param("deviceId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDDD001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSDDD001",
			Error: "Device can't be found",
		})
	}

	if permission.Admin == 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDDD002"})
		contextLogger.Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSDDD002",
			Error: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM devices WHERE id=$1", c.Param("deviceId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDDD003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDDD003",
			Error: "Device can't be deleted",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "device", c.Param("deviceId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDDD004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDDD004",
			Error: "Device can't be deleted",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Device deleted",
	})
}

type permissionDevice struct {
	Permission
	User
	DeviceID           string         `db:"d_id"`
	DeviceName         string         `db:"d_name"`
	DeviceIcon         sql.NullString `db:"d_icon"`
	DeviceRoomID       string         `db:"d_roomid"`
	DevicePlugin       string         `db:"d_plugin"`
	DeviceGatewayID    string         `db:"d_gatewayid"`
	DevicePhysicalID   string         `db:"d_physicalid"`
	DevicePhysicalName string         `db:"d_physicalname"`
	DeviceCreatedAt    string         `db:"d_createdat"`
	DeviceUpdatedAt    string         `db:"d_updatedat"`
}

type deviceRes struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	GatewayID    string `json:"gatewayId"`
	PhysicalID   string `json:"physicalId"`
	PhysicalName string `json:"physicalName"`
	Plugin       string `json:"plugin"`
	RoomID       string `json:"room_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Creator      User   `json:"creator"`
	Read         int    `json:"read"`
	Write        int    `json:"write"`
	Manage       int    `json:"manage"`
	Admin        int    `json:"admin"`
}

// GetDevices route get list of user devices
func GetDevices(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT permissions.*, users.*,
		devices.id as d_id,	devices.name AS d_name, devices.icon AS d_icon, devices.room_id AS d_roomid, devices.gateway_id AS d_gatewayid, devices.physical_id AS d_physicalid, devices.physical_name AS d_physicalname, devices.plugin AS d_plugin, devices.plugin AS d_plugin, devices.created_at AS d_createdat FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		JOIN users ON devices.creator_id = users.id
		WHERE type=$1 AND user_id=$2 AND (permissions.read=1 OR permissions.admin=1)
	`, "device", user.ID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDGDS001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDGDS001",
			Error: "Devices can't be found",
		})
	}

	var devices []deviceRes
	for rows.Next() {
		var permission permissionDevice
		err := rows.StructScan(&permission)
		if err != nil {
			contextLogger := logger.WithFields(logger.Fields{"code": "CSDGDS002"})
			contextLogger.Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:  "CSDGDS002",
				Error: "Devices can't be found",
			})
		}
		devices = append(devices, deviceRes{
			ID:           permission.DeviceID,
			Name:         permission.DeviceName,
			RoomID:       permission.DeviceRoomID,
			GatewayID:    permission.DeviceGatewayID,
			PhysicalID:   permission.DevicePhysicalID,
			PhysicalName: permission.DeviceName,
			CreatedAt:    permission.DeviceCreatedAt,
			Creator:      permission.User,
			Read:         permission.Permission.Read,
			Write:        permission.Permission.Write,
			Manage:       permission.Permission.Manage,
			Admin:        permission.Permission.Admin,
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: devices,
	})
}

// GetDevice route get specific device with id
func GetDevice(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT permissions.*, users.*,
		devices.id as d_id,	devices.name AS d_name, devices.room_id AS d_roomid, devices.gateway_id AS d_gatewayid, devices.physical_id AS d_physicalid, devices.physical_name AS d_physicalname, devices.plugin AS d_plugin, devices.created_at AS d_createdat FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		JOIN users ON devices.creator_id = users.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "device", c.Param("deviceId"), user.ID)

	if row == nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDGD001"})
		contextLogger.Errorf("QueryRowx: Select permissions")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSDGD001",
			Error: "Device not found",
		})
	}

	var permission permissionDevice
	err := row.StructScan(&permission)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSDGD002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSDGD002",
			Error: "Error 4: Can't retrieve devices",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: deviceRes{
			ID:           permission.DeviceID,
			Name:         permission.DeviceName,
			RoomID:       permission.DeviceRoomID,
			GatewayID:    permission.DeviceGatewayID,
			PhysicalID:   permission.DevicePhysicalID,
			PhysicalName: permission.DeviceName,
			CreatedAt:    permission.DeviceCreatedAt,
			Creator:      permission.User,
			Read:         permission.Permission.Read,
			Write:        permission.Permission.Write,
			Manage:       permission.Permission.Manage,
			Admin:        permission.Permission.Admin,
		},
	})
}
