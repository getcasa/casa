package server

import (
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMGM001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSMGM001",
			Error: "Members can't be retrieved",
		})
	}

	var members []memberRes
	for rows.Next() {
		var permission permissionMember
		err := rows.StructScan(&permission)
		if err != nil {
			contextLogger := logger.WithFields(logger.Fields{"code": "CSMGM002"})
			contextLogger.Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:  "CSMGM002",
				Error: "Members can't be retrieved",
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
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMAM001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMAM001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMAM002"})
		contextLogger.Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMAM002",
			Error: err.Error(),
		})
	}

	var reqUser User
	err := DB.QueryRowx("SELECT * FROM users WHERE email=$1", req.Email).StructScan(&reqUser)

	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMAM003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:  "CSMAM003",
			Error: "User can't be found",
		})
	}

	var permission Permission
	err = DB.QueryRowx("SELECT * FROM permissions WHERE user_id=$1 AND type_id=$2", reqUser.ID, c.Param("homeId")).StructScan(&permission)

	if err == nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMAM004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMAM004",
			Error: reqUser.Firstname + " was already added to your home",
		})
	}

	permissionID := NewULID().String()
	_, err = DB.Exec(`
		INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, permissionID, reqUser.ID, "home", c.Param("homeId"), 1, 0, 0, 0, time.Now().Format(time.RFC1123))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMAM005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSMAM005",
			Error: "User can't be added to your home",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: reqUser.Firstname + " has been added to your home",
	})
}

// RemoveMember route remove a member to an home
func RemoveMember(c echo.Context) error {
	var reqHome Home
	err := DB.QueryRowx("SELECT * FROM homes WHERE id=$1", c.Param("homeId")).StructScan(&reqHome)

	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMRM001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Members can't be found",
		})
	}

	var reqUser User
	err = DB.QueryRowx("SELECT * FROM users WHERE id=$1", c.Param("userId")).StructScan(&reqUser)

	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMRM002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Members can't be found",
		})
	}

	if reqHome.CreatorID == reqUser.ID {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMRM003"})
		contextLogger.Warnf("%s == %s", reqHome.CreatorID, reqUser.ID)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Home's creator can't be removed",
		})
	}

	_, err = DB.Exec(`
		DELETE FROM permissions WHERE user_id=$1 AND type='home' AND type_id=$2
	`, c.Param("userId"), c.Param("homeId"))

	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMRM004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Member can't be deleted",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: reqUser.Firstname + " has been removed from your home",
	})
}

type editMemberReq struct {
	Read   string
	Write  string
	Manage string
	Admin  string
}

// EditMember route create a new permission to authorize an user
func EditMember(c echo.Context) error {
	req := new(editMemberReq)
	if err := c.Bind(req); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM001"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMEM001",
			Error: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Read", "Write", "Manage", "Admin"}); err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM002"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMEM002",
			Error: err.Error(),
		})
	}

	read, err := strconv.Atoi(req.Read)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM003"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMEM003",
			Error: "Read has a wrong value",
		})
	}
	write, err := strconv.Atoi(req.Write)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM004"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMEM004",
			Error: "Write has a wrong value",
		})
	}
	manage, err := strconv.Atoi(req.Manage)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM005"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMEM005",
			Error: "Manage has a wrong value",
		})
	}
	admin, err := strconv.Atoi(req.Admin)
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM006"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:  "CSMEM006",
			Error: "Admin has a wrong value",
		})
	}

	_, err = DB.Exec(`
		UPDATE permissions
		SET read=$1, write=$2, manage=$3, admin=$4
		WHERE user_id=$5 AND type=$6 AND type_id=$7
	`, read, write, manage, admin, c.Param("userId"), "home", c.Param("homeId"))
	if err != nil {
		contextLogger := logger.WithFields(logger.Fields{"code": "CSMEM007"})
		contextLogger.Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:  "CSMEM007",
			Error: "Member can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Member has been updated",
	})
}
