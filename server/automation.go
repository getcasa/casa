package server

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
	"github.com/lib/pq"
)

type addAutomationReq struct {
	Name            string
	Trigger         []string
	TriggerKey      []string
	TriggerOperator []string
	TriggerValue    []string
	Action          []string
	ActionCall      []string
	ActionValue     []string
	Status          bool
}

// AddAutomation route create and add user to an automation
func AddAutomation(c echo.Context) error {
	req := new(addAutomationReq)
	if err := c.Bind(req); err != nil {
		fmt.Println(err)
		return err
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name", "Trigger", "TriggerValue", "TriggerKey", "TriggerOperator", "Action", "ActionCall", "ActionValue"}); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	automationID := NewULID().String()
	newAutomation := Automation{
		ID:              automationID,
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
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Trigger device not found",
			})
		}
	}

	for _, act := range req.Action {
		err := DB.Get(&device, `SELECT * FROM devices WHERE id = $1`, act)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Action device not found",
			})
		}
	}

	_, err := DB.Exec("INSERT INTO automations (id, name, trigger, trigger_key, trigger_operator, trigger_value, action, action_call, action_value, status, creator_id, home_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
		newAutomation.ID, newAutomation.Name, pq.Array(newAutomation.Trigger), pq.Array(newAutomation.TriggerKey), pq.Array(newAutomation.TriggerOperator), pq.Array(newAutomation.TriggerValue), pq.Array(newAutomation.Action), pq.Array(newAutomation.ActionCall), pq.Array(newAutomation.ActionValue), newAutomation.Status, newAutomation.CreatorID, newAutomation.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Error 2: Can't create automation",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: automationID,
	})
}

// // UpdateAutomation route update automation
// func UpdateAutomation(c echo.Context) error {
// 	req := new(addAutomationReq)
// 	if err := c.Bind(req); err != nil {
// 		fmt.Println(err)
// 		return err
// 	}
// 	var missingFields []string
// 	if req.Name == "" {
// 		missingFields = append(missingFields, "name")
// 	}
// 	if req.Status == nil {
// 		missingFields = append(missingFields, "Status")
// 	}
// 	if len(missingFields) >= 2 {
// 		return c.JSON(http.StatusBadRequest, MessageResponse{
// 			Message: "Need one field (Name, RoomID)",
// 		})
// 	}

// 	user := c.Get("user").(User)

// 	var permission Permission
// 	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "automation", c.Param("automationId"))
// 	if err != nil {
// 		fmt.Println(err)
// 		return c.JSON(http.StatusNotFound, MessageResponse{
// 			Message: "Automation not found",
// 		})
// 	}

// 	if permission.Manage == 0 && permission.Admin == 0 {
// 		return c.JSON(http.StatusUnauthorized, MessageResponse{
// 			Message: "Unauthorized modifications",
// 		})
// 	}
// 	request := "UPDATE automations SET "
// 	if req.Name != "" {
// 		request += "Name='" + req.Name + "'"
// 		if req.RoomID != "" {
// 			request += ", room_id='" + req.RoomID + "'"
// 		}
// 	} else if req.RoomID != "" {
// 		request += "room_id='" + req.RoomID + "'"
// 	}
// 	request += " WHERE id=$1"
// 	_, err = DB.Exec(request, c.Param("automationId"))
// 	if err != nil {
// 		fmt.Println(err)
// 		return c.JSON(http.StatusInternalServerError, MessageResponse{
// 			Message: "Error 5: Can't update automation",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, MessageResponse{
// 		Message: "Automation updated",
// 	})
// }

// DeleteAutomation route delete automation
func DeleteAutomation(c echo.Context) error {
	user := c.Get("user").(User)

	_, err := DB.Exec("DELETE FROM automations WHERE creator_id=$1 AND id=$2", user.ID, c.Param("automationId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 6: Can't delete automation",
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
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve automations",
		})
	}

	var automations []automationStruct
	for rows.Next() {

		var auto automationStruct
		err := rows.Scan(&auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerOperator), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.CreatorID)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, MessageResponse{
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
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Automation not found",
		})
	}

	var auto automationStruct
	err := row.Scan(&auto.ID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerOperator), pq.Array(&auto.TriggerValue), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.CreatorID, &auto.HomeID)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 3: Can't retrieve automations",
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
