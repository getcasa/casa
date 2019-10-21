package server

import (
	"database/sql"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
	"github.com/lib/pq"
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

type syncedDatas struct {
	Home        Home
	Gateway     Gateway
	Users       []User
	Automations []Automation
	Devices     []Device
	Rooms       []Room
	Permissions []Permission
}

type automationScan struct {
	ID              string
	HomeID          string `db:"home_id" json:"homeID"`
	Name            string
	Trigger         []string
	TriggerKey      []string `db:"trigger_key" json:"triggerKey"`
	TriggerValue    []string `db:"trigger_value" json:"triggerValue"`
	TriggerOperator []string `db:"trigger_operator" json:"triggerOperator"`
	Action          []string
	ActionCall      []string `db:"action_call" json:"actionCall"`
	ActionValue     []string `db:"action_value" json:"actionValue"`
	Status          bool
	CreatedAt       string `db:"created_at" json:"createdAt"`
	UpdatedAt       string `db:"updated_at" json:"updatedAt"`
	CreatorID       string `db:"creator_id" json:"creatorID"`
}

// SyncGateway sync datas with gateway
func SyncGateway(c echo.Context) error {
	id := c.Param("id")

	var missingFields []string
	if id == "" {
		missingFields = append(missingFields, "id")
	}
	if len(missingFields) > 0 {
		err := errors.New("Some fields missing: " + strings.Join(missingFields, ", "))
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG001",
			Error: err.Error(),
		})
	}

	var synced syncedDatas
	err := DB.Get(&synced.Gateway, `SELECT * FROM gateways WHERE id=$1`, id)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG002",
			Error: "Gateway can't be found",
		})
	}

	err = DB.Get(&synced.Home, `SELECT * FROM homes WHERE id=$1`, synced.Gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG003",
			Error: "Home can't be found",
		})
	}

	err = DB.Select(&synced.Rooms, `SELECT * FROM rooms WHERE home_id=$1`, synced.Gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG004",
			Error: "Rooms can't be found",
		})
	}

	err = DB.Select(&synced.Devices, `SELECT DISTINCT devices.* FROM devices JOIN rooms ON devices.room_id = rooms.id WHERE rooms.home_id = $1`, synced.Gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG005",
			Error: "Devices can't be found",
		})
	}

	err = DB.Select(&synced.Permissions, `SELECT DISTINCT permissions.* FROM permissions
	JOIN users ON permissions.user_id = users.id
	JOIN permissions AS permi ON users.id = permi.user_id WHERE permi.type_id = $1 AND permi.type = 'home'`, synced.Gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG006"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG006",
			Error: "Permissions can't be found",
		})
	}

	err = DB.Select(&synced.Users, `SELECT DISTINCT users.id, users.firstname, users.lastname, users.email, users.birthdate, users.created_at, users.updated_at FROM users JOIN permissions ON users.id = permissions.user_id WHERE permissions.type_id = $1 AND permissions.type = 'home'`, synced.Gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG007"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSGSG007",
			Error: "Members can't be found",
		})
	}

	rows, err := DB.Queryx(`SELECT * FROM automations WHERE home_id=$1`, synced.Gateway.HomeID)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG008"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSGSG008",
			Error: "Automations can't be found",
		})
	}

	var automations []Automation
	for rows.Next() {
		var auto automationScan
		err := rows.Scan(&auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerValue), pq.Array(&auto.TriggerOperator), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.UpdatedAt, &auto.CreatorID)
		if err != nil {
			contextLogger := logger.WithFields(logger.Fields{"code": "CSGSG009"})
			contextLogger.Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:  "CSGSG009",
				Error: "Automations can't be found",
			})
		}
		automations = append(automations, Automation{
			ID:              auto.ID,
			Name:            auto.Name,
			Trigger:         auto.Trigger,
			TriggerKey:      auto.TriggerKey,
			TriggerValue:    auto.TriggerValue,
			TriggerOperator: auto.TriggerOperator,
			Action:          auto.Action,
			ActionCall:      auto.ActionCall,
			ActionValue:     auto.ActionValue,
			Status:          auto.Status,
			CreatedAt:       auto.CreatedAt,
			UpdatedAt:       auto.UpdatedAt,
			CreatorID:       auto.CreatorID,
			HomeID:          auto.HomeID,
		})
	}

	synced.Automations = automations

	return c.JSON(http.StatusOK, DataReponse{
		Data: synced,
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
