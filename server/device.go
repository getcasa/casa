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
	Config       string
	RoomID       string
	Plugin       string
	Icon         string
}

// AddDevice route create a device
func AddDevice(c echo.Context) error {
	req := new(addDeviceReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSDAD001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDAD001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name", "GatewayID", "PhysicalID", "PhysicalName", "Plugin", "Config"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSDAD002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDAD002",
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var device Device
	err := DB.Get(&device, "SELECT * FROM devices WHERE physical_id=$1 AND gateway_id=$2", req.PhysicalID, req.GatewayID)
	if err == nil {
		logger.WithFields(logger.Fields{"code": "CSDAD003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDAD003",
			Message: "Device with the same physical id already exist in this gateway",
		})
	}

	row, err := DB.Query("INSERT INTO devices (id, name, icon, room_id, gateway_id, physical_id, physical_name, config, plugin, creator_id) VALUES (generate_ulid(), $1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id;",
		req.Name, req.Icon, c.Param("roomId"), req.GatewayID, req.PhysicalID, req.PhysicalName, req.Config, req.Plugin, user.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDAD004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDAD004",
			Message: "Device can't be created",
		})
	}
	var deviceID string
	row.Next()
	err = row.Scan(&deviceID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDAD005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDAD005",
			Message: "Device can't be created",
		})
	}

	newPermission := Permission{
		UserID: user.ID,
		Type:   "device",
		TypeID: deviceID,
		Read:   true,
		Write:  true,
		Manage: true,
		Admin:  true,
	}
	_, err = DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin) VALUES (generate_ulid(), :user_id, :type, :type_id, :read, :write, :manage, :admin)", newPermission)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDAD006"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDAD006",
			Message: "Device can't be created",
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
		logger.WithFields(logger.Fields{"code": "CSDUD001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDUD001",
			Message: "Wrong parameters",
		})
	}

	request := "UPDATE devices SET name = COALESCE($1, name), room_id = COALESCE($2, room_id) WHERE id = $3"
	_, err := DB.Exec(request, utils.NewNullString(req.Name), utils.NewNullString(req.RoomID), c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDUD005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDUD005",
			Message: "Device can't be updated",
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
		logger.WithFields(logger.Fields{"code": "CSDDD001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSDDD001",
			Message: "Device can't be found",
		})
	}

	if permission.Admin == false {
		logger.WithFields(logger.Fields{"code": "CSDDD002"}).Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSDDD002",
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM devices WHERE id=$1", c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDDD003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDDD003",
			Message: "Device can't be deleted",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "device", c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDDD004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDDD004",
			Message: "Device can't be deleted",
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
	DeviceConfig       string         `db:"d_config"`
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
	Config       string `json:"config"`
	Plugin       string `json:"plugin"`
	RoomID       string `json:"room_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Creator      User   `json:"creator"`
	Read         bool   `json:"read"`
	Write        bool   `json:"write"`
	Manage       bool   `json:"manage"`
	Admin        bool   `json:"admin"`
}

// GetDevices route get list of user devices
func GetDevices(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT permissions.*, users.*,
		devices.id as d_id,	devices.name AS d_name, devices.icon AS d_icon, devices.room_id AS d_roomid, devices.gateway_id AS d_gatewayid, devices.physical_id AS d_physicalid, devices.physical_name AS d_physicalname, devices.config AS d_config, devices.plugin AS d_plugin, devices.plugin AS d_plugin, devices.created_at AS d_createdat FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		JOIN users ON devices.creator_id = users.id
		WHERE type=$1 AND user_id=$2 AND (permissions.read=true OR permissions.admin=true)
	`, "device", user.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDGDS001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDGDS001",
			Message: "Devices can't be found",
		})
	}

	var devices []deviceRes
	for rows.Next() {
		var permission permissionDevice
		err := rows.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDGDS002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSDGDS002",
				Message: "Devices can't be found",
			})
		}
		devices = append(devices, deviceRes{
			ID:           permission.DeviceID,
			Name:         permission.DeviceName,
			RoomID:       permission.DeviceRoomID,
			GatewayID:    permission.DeviceGatewayID,
			PhysicalID:   permission.DevicePhysicalID,
			PhysicalName: permission.DevicePhysicalName,
			Config:       permission.DeviceConfig,
			CreatedAt:    permission.DeviceCreatedAt,
			Creator:      permission.User,
			Read:         permission.Permission.Read,
			Write:        permission.Permission.Write,
			Manage:       permission.Permission.Manage,
			Admin:        permission.Permission.Admin,
		})
	}

	return c.JSON(http.StatusOK, devices)
}

// GetDevice route get specific device with id
func GetDevice(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT permissions.*, users.*,
		devices.id as d_id,	devices.name AS d_name, devices.room_id AS d_roomid, devices.gateway_id AS d_gatewayid, devices.physical_id AS d_physicalid, devices.physical_name AS d_physicalname, devices.config AS d_config, devices.plugin AS d_plugin, devices.created_at AS d_createdat FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		JOIN users ON devices.creator_id = users.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "device", c.Param("deviceId"), user.ID)

	if row == nil {
		logger.WithFields(logger.Fields{"code": "CSDGD001"}).Errorf("QueryRowx: Select permissions")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSDGD001",
			Message: "Device not found",
		})
	}

	var permission permissionDevice
	err := row.StructScan(&permission)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDGD002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDGD002",
			Message: "Error 4: Can't retrieve devices",
		})
	}

	return c.JSON(http.StatusOK, deviceRes{
		ID:           permission.DeviceID,
		Name:         permission.DeviceName,
		RoomID:       permission.DeviceRoomID,
		GatewayID:    permission.DeviceGatewayID,
		PhysicalID:   permission.DevicePhysicalID,
		PhysicalName: permission.DevicePhysicalName,
		Config:       permission.DeviceConfig,
		CreatedAt:    permission.DeviceCreatedAt,
		Creator:      permission.User,
		Read:         permission.Permission.Read,
		Write:        permission.Permission.Write,
		Manage:       permission.Permission.Manage,
		Admin:        permission.Permission.Admin,
	})
}
