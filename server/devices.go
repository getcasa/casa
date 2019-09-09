package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

type addDeviceReq struct {
	GatewayID  string
	Name       string
	PhysicalID string
	RoomID     string
}

// AddDevice route create and add user to an device
func AddDevice(c echo.Context) error {
	req := new(addDeviceReq)
	if err := c.Bind(req); err != nil {
		return err
	}
	var missingFields []string
	if req.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if req.GatewayID == "" {
		missingFields = append(missingFields, "GatewayID")
	}
	if req.PhysicalID == "" {
		missingFields = append(missingFields, "PhysicalID")
	}
	if req.RoomID == "" {
		missingFields = append(missingFields, "RoomID")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	user := c.Get("user").(User)

	deviceID := NewULID().String()
	newDevice := Device{
		ID:         deviceID,
		Name:       req.Name,
		RoomID:     req.RoomID,
		GatewayID:  req.GatewayID,
		PhysicalID: req.PhysicalID,
		CreatedAt:  time.Now().Format(time.RFC1123),
		CreatorID:  user.ID,
	}
	_, err := DB.NamedExec("INSERT INTO devices (id, name, room_id, gateway_id, physical_id, created_at, creator_id) VALUES (:id, :name, :room_id, :gateway_id, :physical_id, :created_at, :creator_id)", newDevice)
	log.Println(err)
	if err != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 2: Can't create device",
		})
	}

	permissionID := NewULID().String()
	newPermission := Permission{
		ID:        permissionID,
		UserID:    user.ID,
		Type:      "device",
		TypeID:    deviceID,
		Read:      1,
		Write:     1,
		Manage:    1,
		Admin:     1,
		UpdatedAt: time.Now().Format(time.RFC1123),
	}
	DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin, updated_at) VALUES (:id, :user_id, :type, :type_id, :read, :write, :manage, :admin, :updated_at)", newPermission)

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: deviceID,
	})
}

// UpdateDevice route update device
func UpdateDevice(c echo.Context) error {
	req := new(addDeviceReq)
	if err := c.Bind(req); err != nil {
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
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "device", c.Param("id"))
	if err != nil {
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
	_, err = DB.Exec(request, c.Param("id"))
	if err != nil {
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
	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "device", c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Device not found",
		})
	}

	if permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM devices WHERE id=$1", c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 6: Can't delete device",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "device", c.Param("id"))
	if err != nil {
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
	Device
}

type deviceRes struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RoomID     string `json:"room_id"`
	GatewayID  string `json:"gatewayId"`
	PhysicalID string `json:"physicalId"`
	CreatedAt  string `json:"created_at"`
	Creator    User   `json:"creator"`
	Read       int    `json:"read"`
	Write      int    `json:"write"`
	Manage     int    `json:"manage"`
	Admin      int    `json:"admin"`
}

// GetDevices route get list of user devices
func GetDevices(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT * FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		WHERE type=$1 AND user_id=$2
	`, "device", user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve devices",
		})
	}

	var devices []deviceRes
	for rows.Next() {
		var permission permissionDevice
		err := rows.StructScan(&permission)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Error 3: Can't retrieve devices",
			})
		}
		devices = append(devices, deviceRes{
			ID:         permission.Device.ID,
			Name:       permission.Device.Name,
			RoomID:     permission.Device.RoomID,
			GatewayID:  permission.Device.GatewayID,
			PhysicalID: permission.Device.PhysicalID,
			CreatedAt:  permission.Device.CreatedAt,
			Creator:    user,
			Read:       permission.Permission.Read,
			Write:      permission.Permission.Write,
			Manage:     permission.Permission.Manage,
			Admin:      permission.Permission.Admin,
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
		SELECT * FROM permissions
		JOIN devices ON permissions.type_id = devices.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "device", c.Param("id"), user.ID)

	if row == nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Device not found",
		})
	}

	var permission permissionDevice
	err := row.StructScan(&permission)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 4: Can't retrieve devices",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: deviceRes{
			ID:         permission.Device.ID,
			Name:       permission.Device.Name,
			RoomID:     permission.Device.RoomID,
			GatewayID:  permission.Device.GatewayID,
			PhysicalID: permission.Device.PhysicalID,
			CreatedAt:  permission.Device.CreatedAt,
			Creator:    user,
			Read:       permission.Permission.Read,
			Write:      permission.Permission.Write,
			Manage:     permission.Permission.Manage,
			Admin:      permission.Permission.Admin,
		},
	})
}
