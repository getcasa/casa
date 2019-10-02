package server

import (
	"fmt"
	"net/http"
	"reflect"

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
}

// AddDevice route create a device
func AddDevice(c echo.Context) error {
	req := new(addDeviceReq)
	if err := c.Bind(req); err != nil {
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name", "GatewayID", "PhysicalID", "PhysicalName", "Plugin"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	var device Device
	err := DB.Get(&device, "SELECT * FROM devices WHERE physical_id=$1 AND gateway_id=$2", req.PhysicalID, req.GatewayID)
	if err == nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Device with the same physical id already exist in this gateway",
		})
	}

	row, err := DB.Query("INSERT INTO devices (id, name, room_id, gateway_id, physical_id, physical_name, plugin, created_at, creator_id) VALUES (generate_ulid(), :name, :room_id, :gateway_id, :physical_id, :physical_name, :plugin, :creator_id) RETURNING id;", req.Name, c.Param("roomId"), req.GatewayID, req.PhysicalID, req.PhysicalName, req.Plugin, user.ID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 1: Token can't be create",
		})
	}
	var deviceID string
	row.Next()
	err = row.Scan(&deviceID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 2: Token can't be create",
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
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Can't add new permission: " + err.Error(),
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
		fmt.Println(err)
		return err
	}
	var missingFields []string
	if req.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if req.RoomID == "" {
		missingFields = append(missingFields, "RoomID")
	}
	if len(missingFields) >= 2 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Need one field (Name, RoomID)",
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "device", c.Param("deviceId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Device not found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
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
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't update device",
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
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Device not found",
		})
	}

	if permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM devices WHERE id=$1", c.Param("deviceId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 6: Can't delete device",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "device", c.Param("deviceId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 7: Can't delete device",
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
	DeviceRoomID       string `db:"d_roomid"`
	DeviceGatewayID    string `db:"d_gatewayid"`
	DevicePhysicalID   string `db:"d_physicalid"`
	DevicePhysicalName string `db:"d_physicalname"`
	DeviceCreatedAt    string `db:"d_createdat"`
}

type deviceRes struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RoomID       string `json:"room_id"`
	GatewayID    string `json:"gatewayId"`
	PhysicalID   string `json:"physicalId"`
	PhysicalName string `json:"physicalName"`
	CreatedAt    string `json:"created_at"`
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
		devices.id as d_id,	devices.name AS d_name, devices.room_id AS d_roomid, devices.gateway_id AS d_gatewayid, devices.physical_id AS d_physicalid, , devices.physical_name AS d_physicalname, devices.plugin AS d_plugin, devices.created_at AS d_createdat FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		JOIN users ON devices.creator_id = users.id
		WHERE type=$1 AND user_id=$2 AND (permissions.read=1 OR permissions.admin=1)
	`, "device", user.ID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve devices",
		})
	}

	var devices []deviceRes
	for rows.Next() {
		var permission permissionDevice
		err := rows.StructScan(&permission)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Error 3: Can't retrieve devices",
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
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Device not found",
		})
	}

	var permission permissionDevice
	err := row.StructScan(&permission)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 4: Can't retrieve devices",
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
