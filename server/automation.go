package server

import (
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

type automationStruct struct {
	ID              string
	HomeID          string `db:"home_id" json:"homeID"`
	Name            string
	Trigger         []string
	TriggerKey      []string
	TriggerOperator []string
	TriggerValue    []string
	Action          []string
	ActionCall      []string
	ActionValue     []string
	Status          bool
	CreatedAt       string `db:"created_at" json:"createdAt"`
	CreatorID       string `db:"creator_id" json:"creatorID"`
	User            User
}

// GetAutomations route get list of user automations
func GetAutomations(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT * FROM automations
		WHERE creator_id=$1`, user.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAGAS001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSAGAS001",
			Message: "Automations can't be retrieved",
		})
	}

	var automations []automationStruct
	for rows.Next() {

		var auto automationStruct
		err := rows.Scan(&auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerOperator), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.CreatorID)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSAGAS002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSAGAS002",
				Message: "Error 3: Can't retrieve automations",
			})
		}
		automations = append(automations, automationStruct{
			ID:              auto.ID,
			HomeID:          auto.HomeID,
			Name:            auto.Name,
			Trigger:         auto.Trigger,
			TriggerKey:      auto.TriggerKey,
			TriggerOperator: auto.TriggerOperator,
			TriggerValue:    auto.TriggerValue,
			Action:          auto.Action,
			ActionCall:      auto.ActionCall,
			ActionValue:     auto.ActionValue,
			Status:          auto.Status,
			CreatedAt:       auto.CreatedAt,
			CreatorID:       auto.CreatorID,
			User:            user,
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: automations,
	})
}

// GetAutomation route get specific automation with id
func GetAutomation(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT * FROM automations
		WHERE creator_id=$1 AND id=$2`, user.ID, c.Param("automationId"))
	if row == nil {
		logger.WithFields(logger.Fields{"code": "CSAGA001"}).Errorf("QueryRowx: Select automations")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSAGA001",
			Message: "Automation can't be found",
		})
	}

	var auto automationStruct
	err := row.Scan(&auto.ID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerOperator), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.CreatorID, &auto.HomeID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSAGA002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSAGA002",
			Message: "Automation can't be found",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: automationStruct{
			ID:              auto.ID,
			HomeID:          auto.HomeID,
			Name:            auto.Name,
			Trigger:         auto.Trigger,
			TriggerKey:      auto.TriggerKey,
			TriggerOperator: auto.TriggerOperator,
			TriggerValue:    auto.TriggerValue,
			Action:          auto.Action,
			ActionCall:      auto.ActionCall,
			ActionValue:     auto.ActionValue,
			Status:          auto.Status,
			CreatedAt:       auto.CreatedAt,
			CreatorID:       auto.CreatorID,
			User:            user,
		},
	})

}
