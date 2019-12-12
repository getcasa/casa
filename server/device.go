package server

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/getcasa/sdk"
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

type getDatasDeviceReq struct {
	Field string
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

	request := `UPDATE devices SET
	name = COALESCE($1, name),
	room_id = COALESCE($2, room_id),
	icon = COALESCE($3, icon)
	WHERE id = $4 RETURNING *`
	rows, err := DB.Queryx(request, utils.NewNullString(req.Name), utils.NewNullString(req.RoomID), utils.NewNullString(req.Icon), c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDUD005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDUD005",
			Message: "Device can't be updated",
		})
	}

	defer rows.Close()

	rows.Next()
	var device Device
	err = rows.StructScan(&device)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDUD006"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDUD006",
			Message: "Device can't be updated",
		})
	}

	return c.JSON(http.StatusOK, device)
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
	DeviceID           string `db:"d_id"`
	DeviceName         string `db:"d_name"`
	DeviceIcon         string `db:"d_icon"`
	DeviceRoomID       string `db:"d_roomid"`
	DevicePlugin       string `db:"d_plugin"`
	DeviceGatewayID    string `db:"d_gatewayid"`
	DevicePhysicalID   string `db:"d_physicalid"`
	DevicePhysicalName string `db:"d_physicalname"`
	DeviceConfig       string `db:"d_config"`
	DeviceCreatedAt    string `db:"d_createdat"`
	DeviceUpdatedAt    string `db:"d_updatedat"`
}

type deviceRes struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Icon          string       `json:"icon"`
	GatewayID     string       `json:"gatewayId"`
	PhysicalID    string       `json:"physicalId"`
	PhysicalName  string       `json:"physicalName"`
	Config        string       `json:"config"`
	Plugin        string       `json:"plugin"`
	RoomID        string       `json:"roomId"`
	CreatedAt     string       `json:"createdAt"`
	UpdatedAt     string       `json:"updatedAt"`
	Creator       User         `json:"creator"`
	Read          bool         `json:"read"`
	Write         bool         `json:"write"`
	Manage        bool         `json:"manage"`
	Admin         bool         `json:"admin"`
	PluginDevice  sdk.Device   `json:"pluginDevice"`
	PluginActions []sdk.Action `json:"pluginActions"`
}

// GetDevices route get list of user devices
func GetDevices(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT permissions.*, users.*,
		devices.id as d_id,	devices.name AS d_name, devices.icon AS d_icon, devices.room_id AS d_roomid, devices.gateway_id AS d_gatewayid, devices.physical_id AS d_physicalid, devices.physical_name AS d_physicalname, devices.config AS d_config, devices.plugin AS d_plugin, devices.plugin AS d_plugin, devices.created_at AS d_createdat FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		JOIN users ON devices.creator_id = users.id
		WHERE permissions.type=$1 AND permissions.user_id=$2 AND devices.room_id=$3 AND (permissions.read=true OR permissions.admin=true)
	`, "device", user.ID, c.Param("roomId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDGDS001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDGDS001",
			Message: "Devices can't be found",
		})
	}
	defer rows.Close()

	devices := []deviceRes{}
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

		var pluginDevice sdk.Device
		pluginActions := []sdk.Action{}
		for _, config := range Configs {
			if config.Name != permission.DevicePlugin {
				continue
			}
			for _, device := range config.Devices {
				if device.Name != permission.DevicePhysicalName {
					continue
				}
				pluginDevice = device
				for _, availableAction := range device.Actions {
					for _, action := range config.Actions {
						if action.Name != availableAction {
							continue
						}
						pluginActions = append(pluginActions, action)
					}
				}
			}
		}

		devices = append(devices, deviceRes{
			ID:            permission.DeviceID,
			Name:          permission.DeviceName,
			Icon:          permission.DeviceIcon,
			RoomID:        permission.DeviceRoomID,
			GatewayID:     permission.DeviceGatewayID,
			PhysicalID:    permission.DevicePhysicalID,
			PhysicalName:  permission.DevicePhysicalName,
			Config:        permission.DeviceConfig,
			CreatedAt:     permission.DeviceCreatedAt,
			Creator:       permission.User,
			Read:          permission.Permission.Read,
			Write:         permission.Permission.Write,
			Manage:        permission.Permission.Manage,
			Admin:         permission.Permission.Admin,
			PluginDevice:  pluginDevice,
			PluginActions: pluginActions,
		})
	}

	totalDevices := strconv.Itoa(len(devices))
	c.Response().Header().Set("Content-Range", "0-"+totalDevices+"/"+totalDevices)
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
		WHERE permissions.type=$1 AND permissions.type_id=$2 AND permissions.user_id=$3 AND devices.room_id=$4
	`, "device", c.Param("deviceId"), user.ID, c.Param("roomId"))

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

	var pluginDevice sdk.Device
	pluginActions := []sdk.Action{}
	for _, config := range Configs {
		if config.Name != permission.DevicePlugin {
			continue
		}
		for _, device := range config.Devices {
			if device.Name != permission.DevicePhysicalName {
				continue
			}
			pluginDevice = device
			for _, availableAction := range device.Actions {
				for _, action := range config.Actions {
					if action.Name != availableAction {
						continue
					}
					pluginActions = append(pluginActions, action)
				}
			}
		}
	}

	return c.JSON(http.StatusOK, deviceRes{
		ID:            permission.DeviceID,
		Name:          permission.DeviceName,
		Icon:          permission.DeviceIcon,
		RoomID:        permission.DeviceRoomID,
		GatewayID:     permission.DeviceGatewayID,
		PhysicalID:    permission.DevicePhysicalID,
		PhysicalName:  permission.DevicePhysicalName,
		Config:        permission.DeviceConfig,
		CreatedAt:     permission.DeviceCreatedAt,
		Creator:       permission.User,
		Read:          permission.Permission.Read,
		Write:         permission.Permission.Write,
		Manage:        permission.Permission.Manage,
		Admin:         permission.Permission.Admin,
		PluginDevice:  pluginDevice,
		PluginActions: pluginActions,
	})
}

// GetLogsDevice return list of log for a device
func GetLogsDevice(c echo.Context) error {
	rows, err := DB.Queryx(`
	SELECT logs.* FROM logs
	JOIN devices ON logs.type_id = devices.id
	JOIN rooms ON devices.room_id = rooms.id
	WHERE rooms.home_id = $1 AND devices.room_id=$2 AND type_id=$3 AND type = 'device'
	ORDER BY created_at DESC
	`, c.Param("homeId"), c.Param("roomId"), c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDGLD002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDGLD002",
			Message: "Logs can't be found",
		})
	}

	defer rows.Close()

	logs := []Logs{}
	for rows.Next() {
		var log Logs
		err := rows.StructScan(&log)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDGLD003"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSDGLD003",
				Message: "Logs can't be found",
			})
		}

		logs = append(logs, log)
	}

	return c.JSON(http.StatusOK, logs)
}

// GetDatasDevice return list of datas for a device
func GetDatasDevice(c echo.Context) error {
	req := new(getDatasDeviceReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSAGLA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAGLA001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Field"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSAGLA002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAGLA002",
			Message: err.Error(),
		})
	}

	rows, err := DB.Queryx(`
	SELECT datas.* FROM datas
	JOIN devices ON datas.device_id = devices.id
	JOIN rooms ON devices.room_id = rooms.id
	WHERE rooms.home_id = $1 AND devices.room_id=$2 AND device_id=$3 AND field = $4
	ORDER BY created_at DESC
	`, c.Param("homeId"), c.Param("roomId"), c.Param("deviceId"), req.Field)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDGLD003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDGLD003",
			Message: "Datas can't be found",
		})
	}

	defer rows.Close()

	datas := []Datas{}
	for rows.Next() {
		var data Datas
		err := rows.StructScan(&data)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDGLD004"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSDGLD004",
				Message: "Datas can't be found",
			})
		}

		datas = append(datas, data)
	}

	return c.JSON(http.StatusOK, datas)
}

// GetDeviceMembers route get list of device users
func GetDeviceMembers(c echo.Context) error {
	rows, err := DB.Queryx(`
		SELECT * FROM permissions
		JOIN users ON permissions.user_id = users.id
		WHERE (permissions.type=$1 AND permissions.type_id=$2) OR (permissions.type=$3 AND permissions.type_id=$4)
	`, "room", c.Param("roomId"), "device", c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDGDM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDGDM001",
			Message: "Members can't be retrieved",
		})
	}
	defer rows.Close()

	var permissions []permissionDevice
	for rows.Next() {
		var permission permissionDevice
		err := rows.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDGDM002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSDGDM002",
				Message: "Members can't be retrieved",
			})
		}

		permissions = append(permissions, permission)
	}

	var members []memberRes
	for _, permission := range permissions {
		if permission.Permission.Type == "device" {
			continue
		}

		member := memberRes{
			ID:        permission.User.ID,
			Firstname: permission.User.Firstname,
			Lastname:  permission.User.Lastname,
			Email:     permission.User.Email,
			Birthdate: permission.User.Birthdate,
			CreatedAt: permission.User.CreatedAt,
			Read:      false,
			Write:     false,
			Manage:    false,
			Admin:     false,
		}

		for _, _permission := range permissions {
			if _permission.Permission.Type == "device" && permission.Permission.UserID == _permission.Permission.UserID {
				member.Read = _permission.Permission.Read
				member.Write = _permission.Permission.Write
				member.Manage = _permission.Permission.Manage
				member.Admin = _permission.Permission.Admin
				break
			}
		}
		members = append(members, member)
	}

	totalMembers := strconv.Itoa(len(members))
	c.Response().Header().Set("Content-Range", "0-"+totalMembers+"/"+totalMembers)
	return c.JSON(http.StatusOK, members)
}

// EditDeviceMember route create a new permission to authorize an useron a device
func EditDeviceMember(c echo.Context) error {
	req := new(editMemberReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSDEDM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDEDM001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Read", "Write", "Manage", "Admin"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSDEDM002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSDEDM002",
			Message: err.Error(),
		})
	}

	var permission Permission
	err := DB.QueryRowx(`
		SELECT * FROM permissions
		WHERE user_id=$1 AND type=$2 AND type_id=$3
	 `, c.Param("userId"), "device", c.Param("deviceId")).StructScan(&permission)

	if err != nil {
		_, err = DB.Exec(`
			INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin) 
			VALUES (generate_ulid(), $1, $2, $3, $4, $5, $6, $7)
		`, c.Param("userId"), "device", c.Param("deviceId"), req.Read, req.Write, req.Manage, req.Admin)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDEDM003"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSDEDM003",
				Message: "Member can't be updated",
			})
		}

		return c.JSON(http.StatusOK, MessageResponse{
			Message: "Member has been updated",
		})
	}

	_, err = DB.Exec(`
		UPDATE permissions
		SET read=$1, write=$2, manage=$3, admin=$4
		WHERE user_id=$5 AND type=$6 AND type_id=$7
	`, req.Read, req.Write, req.Manage, req.Admin, c.Param("userId"), "device", c.Param("deviceId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDEDM004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSDEDM004",
			Message: "Member can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Member has been updated",
	})
}
