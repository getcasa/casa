package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/lib/pq"
)

type addAutomationReq struct {
	Name         string
	Trigger      []string
	TriggerValue []string
	Action       []string
	ActionValue  []string
	Status       bool
}

// AddAutomation route create and add user to an automation
func AddAutomation(c echo.Context) error {
	req := new(addAutomationReq)
	if err := c.Bind(req); err != nil {
		fmt.Println(err)
		return err
	}
	var missingFields []string
	if req.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if req.Trigger == nil {
		missingFields = append(missingFields, "Trigger")
	}
	if req.TriggerValue == nil {
		missingFields = append(missingFields, "TriggerValue")
	}
	if req.Action == nil {
		missingFields = append(missingFields, "Action")
	}
	if req.ActionValue == nil {
		missingFields = append(missingFields, "ActionValue")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	user := c.Get("user").(User)

	automationID := NewULID().String()
	newAutomation := Automation{
		ID:           automationID,
		Name:         req.Name,
		Trigger:      req.Trigger,
		TriggerValue: req.TriggerValue,
		Action:       req.Action,
		ActionValue:  req.ActionValue,
		Status:       true,
		CreatedAt:    time.Now().Format(time.RFC1123),
		CreatorID:    user.ID,
	}

	fmt.Println(pq.Array(newAutomation.Trigger))
	_, err := DB.Query("INSERT INTO automations (id, name, trigger, trigger_value, action, action_value, status, created_at, creator_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", newAutomation.ID, newAutomation.Name, pq.Array(newAutomation.Trigger), pq.Array(newAutomation.TriggerValue), pq.Array(newAutomation.Action), pq.Array(newAutomation.ActionValue), newAutomation.Status, newAutomation.CreatedAt, newAutomation.CreatorID)
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
// 	err := DB.Get(&permission, "SELECT * FROM permissions WHERE user_id=$1 AND type=$2 AND type_id=$3", user.ID, "automation", c.Param("id"))
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
// 	_, err = DB.Exec(request, c.Param("id"))
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

	_, err := DB.Exec("DELETE FROM automations WHERE creator_id=$1 AND id=$2", user.ID, c.Param("id"))
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

	fmt.Println(rows)

	var automations []Automation
	for rows.Next() {
		var auto Automation
		err := rows.StructScan(&auto)
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
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: automations,
	})
}

// // GetAutomation route get specific automation with id
// func GetAutomation(c echo.Context) error {
// 	user := c.Get("user").(User)

// 	row := DB.QueryRowx(`
// 		SELECT permissions.*, users.*,
// 		automations.id as d_id,	automations.name AS d_name, automations.room_id AS d_roomid, automations.gateway_id AS d_gatewayid, automations.physical_id AS d_physicalid, automations.created_at AS d_createdat FROM permissions
// 		JOIN automations ON permissions.type_id = automations.id
// 		JOIN users ON automations.creator_id = users.id
// 		WHERE type=$1 AND type_id=$2 AND user_id=$3
// 	`, "automation", c.Param("id"), user.ID)

// 	if row == nil {
// 		return c.JSON(http.StatusNotFound, MessageResponse{
// 			Message: "Automation not found",
// 		})
// 	}

// 	var permission permissionAutomation
// 	err := row.StructScan(&permission)
// 	if err != nil {
// 		fmt.Println(err)
// 		return c.JSON(http.StatusInternalServerError, MessageResponse{
// 			Message: "Error 4: Can't retrieve automations",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, DataReponse{
// 		Data: automationRes{
// 			ID:         permission.AutomationID,
// 			Name:       permission.AutomationName,
// 			RoomID:     permission.AutomationRoomID,
// 			GatewayID:  permission.AutomationGatewayID,
// 			PhysicalID: permission.AutomationPhysicalID,
// 			CreatedAt:  permission.AutomationCreatedAt,
// 			Creator:    permission.User,
// 			Read:       permission.Permission.Read,
// 			Write:      permission.Permission.Write,
// 			Manage:     permission.Permission.Manage,
// 			Admin:      permission.Permission.Admin,
// 		},
// 	})
// }
