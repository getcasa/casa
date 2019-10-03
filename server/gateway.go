package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"

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
		fmt.Println(err)
		return err
	} else if req.ID == "" {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Empty ID",
		})
	} else if checkUlid(req.ID) != nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Not an ULID",
		})
	}

	newGateway := Gateway{
		ID:    req.ID,
		Model: req.Model,
	}
	_, err = DB.NamedExec("INSERT INTO gateways (id, model) VALUES (:id, :model)", newGateway)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error: can't create gateway",
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
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	var gateway Gateway
	err := DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1", id)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	user := c.Get("user").(User)

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Permission for gateway not found",
		})
	}

	if permission.Manage == 0 && permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET Name=$1 WHERE id=$2", req.Name, gateway.ID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't update Gateway",
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
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	var permission Permission
	err = DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "home", gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Permission for gateway not found",
		})
	}

	if permission.Admin == 0 {
		return c.JSON(http.StatusUnauthorized, MessageResponse{
			Message: "Unauthorized modifications",
		})
	}

	_, err = DB.Exec("DELETE FROM gateways WHERE id=$1", id)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 6: Can't delete gateway",
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
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"ID", "User", "HomeID"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	var gateway Gateway
	err = DB.Get(&gateway, "SELECT * FROM gateways WHERE id=$1 AND creator_id = '' AND home_id = ''", req.ID)
	if err == nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Gateway already linked",
		})
	}

	_, err = DB.Exec("UPDATE gateways SET creator_id=$1, home_id=$2 WHERE id=$3", req.User, req.HomeID, req.ID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't link Gateway",
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
	ID           string
	Name         string
	Trigger      []string
	TriggerValue []string `db:"trigger_value" json:"triggerValue"`
	Action       []string
	ActionValue  []string `db:"action_value" json:"actionValue"`
	Status       bool
	CreatedAt    string `db:"created_at" json:"createdAt"`
	CreatorID    string `db:"creator_id" json:"creatorID"`
	HomeID       string `db:"home_id" json:"homeID"`
}

// SyncGateway sync datas with gateway
func SyncGateway(c echo.Context) error {
	id := c.Param("id")

	var missingFields []string
	if id == "" {
		missingFields = append(missingFields, "id")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	var synced syncedDatas
	err := DB.Get(&synced.Gateway, `SELECT * FROM gateways WHERE id=$1`, id)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 3: not found",
		})
	}

	err = DB.Get(&synced.Home, `SELECT * FROM homes WHERE id=$1`, synced.Gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 3: not found",
		})
	}

	err = DB.Select(&synced.Rooms, `SELECT * FROM rooms WHERE home_id=$1`, synced.Gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 3: not found",
		})
	}

	err = DB.Select(&synced.Devices, `SELECT DISTINCT devices.* FROM devices JOIN rooms ON devices.room_id = rooms.id WHERE rooms.home_id = $1`, synced.Gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 3: not found",
		})
	}

	err = DB.Select(&synced.Permissions, `SELECT DISTINCT permissions.* FROM permissions
	JOIN users ON permissions.user_id = users.id
	JOIN permissions AS permi ON users.id = permi.user_id WHERE permi.type_id = $1 AND permi.type = 'home'`, synced.Gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 3: not found",
		})
	}

	err = DB.Select(&synced.Users, `SELECT DISTINCT users.* FROM users JOIN permissions ON users.id = permissions.user_id WHERE permissions.type_id = $1 AND permissions.type = 'home'`, synced.Gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 3: not found",
		})
	}

	rows, err := DB.Queryx(`SELECT * FROM automations WHERE home_id=$1`, synced.Gateway.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve automations",
		})
	}

	var automations []Automation
	for rows.Next() {
		var auto automationScan
		err := rows.Scan(&auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.CreatorID)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Error 3: Can't retrieve automations",
			})
		}
		automations = append(automations, Automation{
			ID:           auto.ID,
			Name:         auto.Name,
			Trigger:      auto.Trigger,
			TriggerValue: auto.TriggerValue,
			Action:       auto.Action,
			ActionValue:  auto.ActionValue,
			Status:       auto.Status,
			CreatedAt:    auto.CreatedAt,
			CreatorID:    auto.CreatorID,
			HomeID:       auto.HomeID,
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
}

type gatewayRes struct {
	ID        string         `json:"id"`
	HomeID    sql.NullString `json:"homeId"`
	Name      sql.NullString `json:"name"`
	Model     string         `json:"model"`
	CreatedAt string         `json:"created_at"`
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
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Gateway not found",
		})
	}

	var gateway Gateway
	err := row.StructScan(&gateway)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 4: Can't retrieve gateway",
		})
	}

	row = DB.QueryRowx(`
		SELECT permissions.*, users.*,
		gateways.id as g_id,	gateways.name AS g_name, gateways.home_id AS g_homeid, gateways.model AS g_model, gateways.created_at AS g_createdat FROM permissions
		JOIN gateways ON permissions.type_id = gateways.home_id
		JOIN users ON gateways.creator_id = users.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "home", gateway.HomeID, user.ID)

	if row == nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Permission for gateway not found",
		})
	}

	var permission permissionGateway
	err = row.StructScan(&permission)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't retrieve gateway",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: gatewayRes{
			ID:        permission.GatewayID,
			HomeID:    permission.GatewayHomeID,
			Name:      permission.GatewayName,
			Model:     permission.GatewayModel,
			CreatedAt: permission.GatewayCreatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		},
	})
}
