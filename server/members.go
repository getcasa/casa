package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

type addMemberReq struct {
	Email string
}

// AddMember route create a new permission to authorize an user
func AddMember(c echo.Context) error {
	req := new(addMemberReq)
	if err := c.Bind(req); err != nil {
		fmt.Println(err)
		return err
	}
	var missingFields []string
	if req.Email == "" {
		missingFields = append(missingFields, "email")
	}
	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Some fields missing: " + strings.Join(missingFields, ", "),
		})
	}

	var reqUser User
	err := DB.QueryRowx("SELECT * FROM users WHERE email=$1", req.Email).StructScan(&reqUser)

	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "User not found",
		})
	}

	var permission Permission
	err = DB.QueryRowx("SELECT * FROM permissions WHERE user_id=$1 AND type_id=$2", reqUser.ID, c.Param("homeId")).StructScan(&permission)

	if err == nil {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: reqUser.Firstname + " was already added to your home",
		})
	}

	permissionID := NewULID().String()
	_, err = DB.Exec(`
		INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, permissionID, reqUser.ID, "home", c.Param("homeId"), 1, 0, 0, 0, time.Now().Format(time.RFC1123))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, MessageResponse{
			Message: "Error 5: Can't add user to your home",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: reqUser.Firstname + " has been added to your home",
	})
}

// removeMember route remove a member to an home
func removeMember(c echo.Context) error {
	var reqHome Home
	err := DB.QueryRowx("SELECT * FROM homes WHERE id=$1", c.Param("homeId")).StructScan(&reqHome)

	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "User not found",
		})
	}

	var reqUser User
	err = DB.QueryRowx("SELECT * FROM users WHERE id=$1", c.Param("userId")).StructScan(&reqUser)

	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "User not found",
		})
	}

	if reqHome.CreatorID == reqUser.ID {
		return c.JSON(http.StatusBadRequest, MessageResponse{
			Message: "Can't remove home's creator",
		})
	}

	_, err = DB.Exec(`
		DELETE FROM permissions WHERE user_id=$1 AND type='home' AND type_id=$2
	`, c.Param("userId"), c.Param("homeId"))

	if err != nil {
		return c.JSON(http.StatusNotFound, MessageResponse{
			Message: "Error 3: Can't delete member",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: reqUser.Firstname + " has been removed from your home",
	})
}
