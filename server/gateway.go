package server

import (
	"database/sql"
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
	ID     string
	User   string
	HomeID string
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGAG001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGAG001",
			Error: "Wrong parameters",
		})
	} else if req.ID == "" {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGAG002"})
		contextLogger.Errorf("Missing ID")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGAG002",
			Error: "Empty ID",
		})
	} else if checkUlid(req.ID) != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGAG003"})
		contextLogger.Errorf("ID mismatch ULID")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGAG003",
			Error: "Not an ULID",
		})
	}

	newGateway := Gateway{
		ID:    req.ID,
		Model: req.Model,
	}
	_, err = DB.NamedExec("INSERT INTO gateways (id, model) VALUES (:id, :model)", newGateway)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGAG004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGAG004",
			Error: "Error: can't create gateway",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: "Gateway created",
	})
}

// UpdateGateway route update gateway
func UpdateGateway(c echo.Context) error {
	id := c.Param("id")
	req := new(updateGatewayReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGUG001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGUG001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGUG002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGUG003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSGUG003",
			Error: "Gateway can't be found",
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGUG004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSGUG004",
			Error: "Gateway can't be found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGUG005"})
		contextLogger.Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSGUG005",
			Error: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET Name=$1 WHERE id=$2", req.Name, gateway.ID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGUG006"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGUG006",
			Error: "Gateway can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Gateway updated",
	})
}

// DeleteGateway route delete gateway
func DeleteGateway(c echo.Context) error {
	id := c.Param("id")
	user := c.Get("user").(User)

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGDG001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSGDG001",
			Error: "Gateway can't be found",
		})
	}

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGDG002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSGDG002",
			Error: "Gateway can't be found",
		})
	}

	if permission.Admin == 0 {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGDG003"})
		contextLogger.Warnf("Unauthorized")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:  "CSGDG003",
			Error: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM gateways WHERE id=$1", id)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGDG004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGDG004",
			Error: "Gateway can't be deleted",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGLG001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGLG001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"ID", "User", "HomeID"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGLG002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	var gateway Gateway
	err = DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1 AND creator_id = '' AND home_id = ''", req.ID)
	if err == nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGLG003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGLG003",
			Error: "Gateway is already linked",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET creator_id=$1, home_id=$2 WHERE id=$3", req.User, req.HomeID, req.ID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGLG004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGLG004",
			Error: "Error 5: Can't link Gateway",
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
	Read      int            `json:"read"`
	Write     int            `json:"write"`
	Manage    int            `json:"manage"`
	Admin     int            `json:"admin"`
}

// GetGateway route get specific gateway with id
func GetGateway(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT * FROM gateways
		WHERE id=$1
	 `, c.Param("id"))

	if row == nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGGG001"})
		contextLogger.Errorf("QueryRowx: Select gateways")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSGGG001",
			Error: "Gateway not found",
		})
	}

	var gateway Gateway
	err := row.StructScan(&gateway)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGSG002",
			Error: "Gateway can't be found",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG003"})
		contextLogger.Errorf("QueryRowx: Select gateways")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSGSG003",
			Error: "Gateway can't be found",
		})
	}

	var permission permissionGateway
	err = row.StructScan(&permission)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG04"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGSG04",
			Error: "Gateway can't be found",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGAP001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGAP001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"ID", "Name", "Config"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSHAH002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGAP002",
			Error: err.Error(),
		})
	}

	_, err := DB.Query("INSERT INTO plugins (id, gateway_id, name, config) VALUES (generate_ulid(), $1, $2, $3)", c.Param("gatewayId"), req.Name, req.Config)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGAP003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGAP003",
			Error: "Plugin can't be added",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: req.Name + " has been added",
	})
}

// GetPlugin route get a gateway plugin
func GetPlugin(c echo.Context) error {
	row := DB.QueryRowx(`
		SELECT * FROM plugins
		WHERE gateway_id=$1 AND name=$2
	 `, c.Param("gatewayId"), c.Param("pluginName"))

	if row == nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSSGGP001"})
		contextLogger.Errorf("QueryRowx: Select plugin")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSSGGP001",
			Error: "Plugin not found",
		})
	}

	var plugin Plugin
	err := row.StructScan(&plugin)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSSGGP002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSSGGP002",
			Error: "Plugin can't be found",
		})
	}

	return c.JSON(http.StatusOK, plugin)
}
