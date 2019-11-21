package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
	"github.com/lib/pq"
)

type addAutomationReq struct {
	Name            string
	Trigger         []string
	TriggerKey      []string
	TriggerValue    []string
	TriggerOperator []string
	Action          []string
	ActionCall      []string
	ActionValue     []string
	Status          bool
}

// AddAutomation route create and add user to an automation
func AddAutomation(c echo.Context) error {
	req := new(addAutomationReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSAAA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAAA001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name", "Trigger", "TriggerValue", "TriggerKey", "TriggerOperator", "Action", "ActionCall", "ActionValue"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSAAA002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAAA002",
			Message: err.Error(),
		})
	}

	if len(req.TriggerOperator) != (len(req.Trigger) - 1) {
		logger.WithFields(logger.Fields{"code": "CSAAA003"}).Errorf("%s", "Number of operator can't match with number of trigger")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAAA003",
			Message: "Number of operator can't match with number of trigger",
		})
	}

	user := c.Get("user").(User)

	newAutomation := Automation{
		Name:            req.Name,
		Trigger:         req.Trigger,
		TriggerKey:      req.TriggerKey,
		TriggerOperator: req.TriggerOperator,
		TriggerValue:    req.TriggerValue,
		Action:          req.Action,
		ActionCall:      req.ActionCall,
		ActionValue:     req.ActionValue,
		HomeID:          c.Param("homeId"),
		Status:          true,
		CreatorID:       user.ID,
	}

	var device Device
	for _, trigg := range req.Trigger {
		err := DB.Get(&device, `SELECT * FROM devices WHERE id = $1`, trigg)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSAAA004"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSAAA004",
				Message: "Trigger device can't be found",
			})
		}
	}

	for _, act := range req.Action {
		err := DB.Get(&device, `SELECT * FROM devices WHERE id = $1`, act)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSAAA005"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSAAA005",
				Message: "Action device can't be found",
			})
		}
	}

	row, err := DB.Query("INSERT INTO automations (id, name, trigger, trigger_key, trigger_operator, trigger_value, action, action_call, action_value, status, creator_id, home_id) VALUES (generate_ulid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
		newAutomation.Name, pq.Array(newAutomation.Trigger), pq.Array(newAutomation.TriggerKey), pq.Array(newAutomation.TriggerOperator), pq.Array(newAutomation.TriggerValue), pq.Array(newAutomation.Action), pq.Array(newAutomation.ActionCall), pq.Array(newAutomation.ActionValue), newAutomation.Status, newAutomation.CreatorID, newAutomation.HomeID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAAA005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAAA005",
			Message: "Automation can't be created",
		})
	}

	var automationID string
	row.Next()
	err = row.Scan(&automationID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAAA006"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAAA006",
			Message: "Automation can't be created",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: automationID,
	})
}

// UpdateAutomation route update automation
func UpdateAutomation(c echo.Context) error {
	req := new(addAutomationReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSAUA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSAAA001",
			Message: "Wrong parameters",
		})
	}

	request := `UPDATE automations
	SET name = COALESCE($1, name),

	trigger = COALESCE($2, trigger),
	trigger_key = COALESCE($3, trigger_key),
	trigger_operator = COALESCE($4, trigger_operator),
	trigger_value = COALESCE($5, trigger_value),
	action = COALESCE($6, action),
	action_call = COALESCE($7, action_call),
	action_value = COALESCE($8, action_value)

	WHERE id=$9`

	fmt.Println(req.Trigger)

	_, err := DB.Exec(request, utils.NewNullString(req.Name), pq.Array(req.Trigger), pq.Array(req.TriggerKey), pq.Array(req.TriggerOperator), pq.Array(req.TriggerValue), pq.Array(req.Action), pq.Array(req.ActionCall), pq.Array(req.ActionValue), c.Param("automationId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAUA002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSAUA002",
			Message: "Automation can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Automation updated",
	})
}

// DeleteAutomation route delete automation
func DeleteAutomation(c echo.Context) error {
	user := c.Get("user").(User)

	_, err := DB.Exec("DELETE FROM automations WHERE creator_id=$1 AND id=$2", user.ID, c.Param("automationId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSADA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSADA001",
			Message: "Automation can't be deleted",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Automation deleted",
	})
}

type permissionAutomations struct {
	User            User
	ID              string   `db:"a_id"`
	Name            string   `db:"a_name"`
	HomeID          string   `db:"a_homeid"`
	Trigger         []string `db:"a_trigger"`
	TriggerKey      []string `db:"a_triggerkey"`
	TriggerOperator []string `db:"a_triggeroperator"`
	TriggerValue    []string `db:"a_triggervalue"`
	Action          []string `db:"a_action"`
	ActionCall      []string `db:"a_actioncall"`
	ActionValue     []string `db:"a_actionvalue"`
	Status          bool     `db:"a_status"`
	CreatedAt       string   `db:"a_createdat"`
	UpdatedAt       string   `db:"a_updatedat"`
	Triggers        string
	Actions         string
}

type automationStruct struct {
	ID              string   `json:"id"`
	HomeID          string   `db:"home_id" json:"homeId"`
	Name            string   `json:"name"`
	Trigger         []Device `json:"trigger"`
	TriggerKey      []string `json:"triggerKey"`
	TriggerOperator []string `json:"triggerOperator"`
	TriggerValue    []string `json:"triggerValue"`
	Action          []Device `json:"action"`
	ActionCall      []string `json:"actionCall"`
	ActionValue     []string `json:"actionValue"`
	Status          bool     `json:"status"`
	CreatedAt       string   `db:"created_at" json:"createdAt"`
	UpdatedAt       string   `db:"updated_at" json:"updatedAt"`
	Creator         User     `json:"creator"`
}

// GetAutomations route get list of user automations
func GetAutomations(c echo.Context) error {
	rows, err := DB.Queryx(`
		SELECT t.*,
		array(SELECT `+DeviceJSONSelect+` FROM devices WHERE devices.id = ANY(a_trigger)) AS triggers,
		array(SELECT `+DeviceJSONSelect+` FROM devices WHERE devices.id = ANY(a_action)) AS actions
		FROM(SELECT users.*, automations.id as a_id,	automations.name AS a_name, automations.home_id AS a_homeid, automations.trigger AS a_trigger, automations.trigger_key AS a_triggerkey, automations.trigger_operator AS a_triggeroperator, automations.trigger_value AS a_triggervalue, automations.action AS a_action, automations.action_call AS a_actioncall, automations.action_value AS a_actionvalue, automations.status AS a_status, automations.created_at AS a_createdat, automations.updated_at AS a_updatedat FROM automations
		JOIN users ON automations.creator_id = users.id
		WHERE automations.home_id=$1) AS t
	`, c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAGAS001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSAGAS001",
			Message: "Automations can't be retrieved",
		})
	}
	defer rows.Close()

	var automations []automationStruct
	for rows.Next() {

		var auto permissionAutomations
		err := rows.Scan(&auto.User.ID, &auto.User.Firstname, &auto.User.Lastname, &auto.User.Email, &auto.User.Password, &auto.User.Birthdate, &auto.User.CreatedAt, &auto.User.UpdatedAt, &auto.ID, &auto.Name, &auto.HomeID, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerOperator), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.UpdatedAt, &auto.Triggers, &auto.Actions)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSAGAS002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSAGAS002",
				Message: "Automations can't be retrieved",
			})
		}

		automations = append(automations, automationStruct{
			ID:              auto.ID,
			HomeID:          auto.HomeID,
			Name:            auto.Name,
			Trigger:         deviceJSONToStruct(auto.Triggers),
			TriggerKey:      auto.TriggerKey,
			TriggerOperator: auto.TriggerOperator,
			TriggerValue:    auto.TriggerValue,
			Action:          deviceJSONToStruct(auto.Actions),
			ActionCall:      auto.ActionCall,
			ActionValue:     auto.ActionValue,
			Status:          auto.Status,
			CreatedAt:       auto.CreatedAt,
			UpdatedAt:       auto.UpdatedAt,
			Creator:         auto.User,
		})
	}

	return c.JSON(http.StatusOK, automations)
}

// GetAutomation route get specific automation with id
func GetAutomation(c echo.Context) error {
	row := DB.QueryRowx(`
	SELECT t.*,
	array(SELECT `+DeviceJSONSelect+` FROM devices WHERE devices.id = ANY(a_trigger)) AS triggers,
	array(SELECT `+DeviceJSONSelect+` FROM devices WHERE devices.id = ANY(a_action)) AS actions
	FROM(SELECT users.*, automations.id as a_id,	automations.name AS a_name, automations.home_id AS a_homeid, automations.trigger AS a_trigger, automations.trigger_key AS a_triggerkey, automations.trigger_operator AS a_triggeroperator, automations.trigger_value AS a_triggervalue, automations.action AS a_action, automations.action_call AS a_actioncall, automations.action_value AS a_actionvalue, automations.status AS a_status, automations.created_at AS a_createdat, automations.updated_at AS a_updatedat FROM automations
	JOIN users ON automations.creator_id = users.id
	WHERE automations.home_id=$1 AND automations.id=$2) AS t
`, c.Param("homeId"), c.Param("automationId"))

	var auto permissionAutomations
	err := row.Scan(&auto.User.ID, &auto.User.Firstname, &auto.User.Lastname, &auto.User.Email, &auto.User.Password, &auto.User.Birthdate, &auto.User.CreatedAt, &auto.User.UpdatedAt, &auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerOperator), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.UpdatedAt, &auto.Triggers, &auto.Actions)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAGA001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSAGAS002",
			Message: "Automations can't be retrieved",
		})
	}

	return c.JSON(http.StatusOK, automationStruct{
		ID:              auto.ID,
		HomeID:          auto.HomeID,
		Name:            auto.Name,
		Trigger:         deviceJSONToStruct(auto.Triggers),
		TriggerKey:      auto.TriggerKey,
		TriggerOperator: auto.TriggerOperator,
		TriggerValue:    auto.TriggerValue,
		Action:          deviceJSONToStruct(auto.Actions),
		ActionCall:      auto.ActionCall,
		ActionValue:     auto.ActionValue,
		Status:          auto.Status,
		CreatedAt:       auto.CreatedAt,
		UpdatedAt:       auto.UpdatedAt,
		Creator:         auto.User,
	})
}

func deviceJSONToStruct(str string) []Device {
	str = "[" + str[1:len(str)-1] + "]"
	var arrayJSON []string
	err := json.Unmarshal([]byte(str), &arrayJSON)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSADJTS001"}).Errorf("%s", err.Error())
		return []Device{}
	}

	var listTrigger []Device
	for _, elem := range arrayJSON {
		var trigger Device
		err = json.Unmarshal([]byte(elem), &trigger)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSADJTS002"}).Errorf("%s", err.Error())
			return []Device{}
		}
		listTrigger = append(listTrigger, trigger)
	}

	return listTrigger
}
