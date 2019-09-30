package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

type permissionMember struct {
	Permission
	User
}

type memberRes struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Birthdate string `json:"birthdate"`
	CreatedAt string `json:"createdAt"`
	Read      int    `json:"read"`
	Write     int    `json:"write"`
	Manage    int    `json:"manage"`
	Admin     int    `json:"admin"`
}

// GetMembers route get list of home members
func GetMembers(c echo.Context) error {
	rows, err := DB.Queryx(`
		SELECT * FROM permissions
		JOIN users ON permissions.user_id = users.id
		WHERE permissions.type=$1 AND permissions.type_id=$2
	`, "home", c.Param("homeId"))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 2: Can't retrieve homes",
		})
	}

	var members []memberRes
	for rows.Next() {
		var permission permissionMember
		err := rows.StructScan(&permission)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, MessageResponse{
				Message: "Error 3: Can't retrieve members",
			})
		}
		members = append(members, memberRes{
			ID:        permission.User.ID,
			Firstname: permission.User.Firstname,
			Lastname:  permission.User.Lastname,
			Email:     permission.User.Email,
			Birthdate: permission.User.Birthdate,
			CreatedAt: permission.User.CreatedAt,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: members,
	})
}
