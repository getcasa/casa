package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
	"github.com/oklog/ulid/v2"
)

type addGatewayReq struct {
	ID    string
	Model string `default0:"custom"`
}

type linkGatewayReq struct {
	ID   string
	User string
}

type updateGatewayReq struct {
	Name string
}

// CheckUlid check if id is a real ulid
func checkUlid(id string) error {
	_, err := ulid.ParseStrict(id)

	return err
}

// AddGateway route add new gateway in system
func AddGateway(c echo.Context) error {
	req := new(addGatewayReq)
	err := c.Bind(req)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGAG001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGAG001",
			Message: "Wrong parameters",
		})
	} else if req.ID == "" {
		logger.WithFields(logger.Fields{"code": "CSGAG002"}).Errorf("Missing ID")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGAG002",
			Message: "Empty ID",
		})
	} else if checkUlid(req.ID) != nil {
		logger.WithFields(logger.Fields{"code": "CSGAG003"}).Errorf("ID mismatch ULID")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGAG003",
			Message: "Not an ULID",
		})
	}

	newGateway := Gateway{
		ID:    req.ID,
		Model: req.Model,
	}
	_, err = DB.NamedExec("INSERT INTO gateways (id, model) VALUES (:id, :model)", newGateway)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGAG004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSGAG004",
			Message: "Error: can't create gateway",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Gateway created",
	})
}

// UpdateGateway route update gateway
func UpdateGateway(c echo.Context) error {
	id := c.Param("gatewayId")
	req := new(updateGatewayReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSGUG001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGUG001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSGUG002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
	}

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGUG003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSGUG003",
			Message: "Gateway can't be found",
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGUG004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSGUG004",
			Message: "Gateway can't be found",
		})
	}

	if permission.Manage == false && permission.Admin == false {
		logger.WithFields(logger.Fields{"code": "CSGUG005"}).Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSGUG005",
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET Name=COALESCE($1, name) WHERE id=$2", utils.NewNullString(req.Name), gateway.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGUG006"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSGUG006",
			Message: "Gateway can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Gateway updated",
	})
}

// DeleteGateway route delete gateway
func DeleteGateway(c echo.Context) error {
	id := c.Param("gatewayId")
	user := c.Get("user").(User)

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGDG001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSGDG001",
			Message: "Gateway can't be found",
		})
	}

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGDG002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSGDG002",
			Message: "Gateway can't be found",
		})
	}

	if permission.Admin == false {
		logger.WithFields(logger.Fields{"code": "CSGDG003"}).Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "CSGDG003",
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM gateways WHERE id=$1", id)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGDG004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSGDG004",
			Message: "Gateway can't be deleted",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Gateway deleted",
	})
}

// LinkGateway route link gateway with user & home
func LinkGateway(c echo.Context) error {
	req := new(linkGatewayReq)
	err := c.Bind(req)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGLG001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGLG001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"ID", "User", "HomeID"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSGLG002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
	}

	var gateway Gateway
	err = DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1 AND creator_id = '' AND home_id = ''", req.ID)
	if err == nil {
		logger.WithFields(logger.Fields{"code": "CSGLG003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGLG003",
			Message: "Gateway is already linked",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET creator_id=$1, home_id=$2 WHERE id=$3", req.User, c.Param("homeId"), req.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGLG004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSGLG004",
			Message: "Error 5: Can't link Gateway",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Gateway linked",
	})
}

type permissionGateway struct {
	Permission
	User
	GatewayID        string         `db:"g_id"`
	GatewayName      sql.NullString `db:"g_name"`
	GatewayHomeID    sql.NullString `db:"g_homeid"`
	GatewayModel     string         `db:"g_model"`
	GatewayCreatedAt string         `db:"g_createdat"`
	GatewayUpdatedAt string         `db:"g_updatedat"`
}

type gatewayRes struct {
	ID        string         `json:"id"`
	HomeID    sql.NullString `json:"homeId"`
	Name      sql.NullString `json:"name"`
	Model     string         `json:"model"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	Creator   User           `json:"creator"`
	Read      bool           `json:"read"`
	Write     bool           `json:"write"`
	Manage    bool           `json:"manage"`
	Admin     bool           `json:"admin"`
}

// GetGateway route get specific gateway with id
func GetGateway(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT * FROM gateways
		WHERE id=$1
	 `, c.Param("gatewayId"))

	if row == nil {
		logger.WithFields(logger.Fields{"code": "CSGGG001"}).Errorf("QueryRowx: Select gateways")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSGGG001",
			Message: "Gateway not found",
		})
	}

	var gateway Gateway
	err := row.StructScan(&gateway)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGSG002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSGSG002",
			Message: "Gateway can't be found",
		})
	}

	row = DB.QueryRowx(`
		SELECT permissions.*, users.*,
		gateways.id as g_id,	gateways.name AS g_name, gateways.home_id AS g_homeid, gateways.model AS g_model, gateways.created_at AS g_createdat, gateways.updated_at AS g_updatedat FROM permissions
		JOIN gateways ON permissions.type_id = gateways.home_id
		JOIN users ON gateways.creator_id = users.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "home", gateway.HomeID, user.ID)

	if row == nil {
		logger.WithFields(logger.Fields{"code": "CSGSG003"}).Errorf("QueryRowx: Select gateways")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSGSG003",
			Message: "Gateway can't be found",
		})
	}

	var permission permissionGateway
	err = row.StructScan(&permission)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGSG04"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSGSG04",
			Message: "Gateway can't be found",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: gatewayRes{
			ID:        permission.GatewayID,
			HomeID:    permission.GatewayHomeID,
			Name:      permission.GatewayName,
			Model:     permission.GatewayModel,
			CreatedAt: permission.GatewayCreatedAt,
			UpdatedAt: permission.GatewayUpdatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		},
	})
}

type addPluginReq struct {
	Name   string
	Config string
}

// AddPlugin add a plugin configuration for gateway
func AddPlugin(c echo.Context) error {
	req := new(addPluginReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSGAP001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGAP001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"ID", "Name", "Config"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSHAH002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGAP002",
			Message: err.Error(),
		})
	}

	_, err := DB.Exec(`
	INSERT INTO plugins (id, gateway_id, name, config)
	SELECT COALESCE((SELECT id FROM plugins WHERE plugins.gateway_id = $1 AND name = $2), generate_ulid()), $1, $2, $3
	ON CONFLICT (id) DO
	UPDATE SET config = $3 WHERE plugins.gateway_id = $1 AND plugins.name = $2`, c.Param("gatewayId"), req.Name, req.Config)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSGAP003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSGAP003",
			Message: "Plugin can't be added",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: req.Name + " has been added",
	})
}

// GetPlugin route get a gateway plugin
func GetPlugin(c echo.Context) error {
	var plugin Plugin
	err := DB.QueryRowx(`
		SELECT * FROM plugins
		WHERE gateway_id=$1 AND name=$2
	 `, c.Param("gatewayId"), c.Param("pluginName")).StructScan(&plugin)

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGGP001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSSGGP001",
			Message: "Plugin not found",
		})
	}

	return c.JSON(http.StatusOK, plugin)
}

type callActionReq struct {
	Action string
	Params string
}

// CallAction call an action on selected gateway
func CallAction(c echo.Context) error {
	req := new(callActionReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGCA001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Action"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCA002"}).Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGCA002",
			Message: err.Error(),
		})
	}

	var device Device
	err := DB.QueryRowx("SELECT * FROM devices WHERE id=$1", c.Param("deviceId")).StructScan(&device)

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCA003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSSGCA003",
			Message: "Device can't be found",
		})
	}

	action := ActionMessage{
		PhysicalID: device.PhysicalID,
		Plugin:     device.Plugin,
		Call:       req.Action,
		Config:     device.Config,
		Params:     req.Params,
	}

	byteAction, _ := json.Marshal(action)

	message := WebsocketMessage{
		Action: "callAction",
		Body:   byteAction,
	}

	marshMessage, _ := json.Marshal(message)
	err = WebsocketWriteMessage(GatewayConn, marshMessage)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCA004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Action can't be sent",
		})
	}

	_, err = DB.Exec("INSERT INTO logs (id, type, type_id, value) VALUES (generate_ulid(), $1, $2, $3)", "device", device.ID, string(byteAction))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCA005"}).Errorf("%s", err.Error())
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Action sent to gateway",
	})
}
