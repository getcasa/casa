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
		logger.WithFields(logger.Fields{"code": "CSMGM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSMGM001",
			Message: "Members can't be retrieved",
		})
	}

	var members []memberRes
	for rows.Next() {
		var permission permissionMember
		err := rows.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSMGM002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSMGM002",
				Message: "Members can't be retrieved",
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

	totalMembers := strconv.Itoa(len(members))
	c.Response().Header().Set("Content-Range", "0-"+totalMembers+"/"+totalMembers)
	return c.JSON(http.StatusOK, members)
}

type addMemberReq struct {
	Email string
}

// AddMember route create a new permission to authorize an user
func AddMember(c echo.Context) error {
	req := new(addMemberReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSMAM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMAM001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Email"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSMAM002"}).Warnf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMAM002",
			Message: err.Error(),
		})
	}

	var reqUser User
	err := DB.QueryRowx("SELECT * FROM users WHERE email=$1", req.Email).StructScan(&reqUser)

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMAM003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSMAM003",
			Message: "User can't be found",
		})
	}

	var permission Permission
	err = DB.QueryRowx("SELECT * FROM permissions WHERE user_id=$1 AND type_id=$2", reqUser.ID, c.Param("homeId")).StructScan(&permission)

	if err == nil {
		logger.WithFields(logger.Fields{"code": "CSMAM004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMAM004",
			Message: reqUser.Firstname + " was already added to your home",
		})
	}

	_, err = DB.Exec(`
		INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin, updated_at) 
		VALUES (generate_ulid(), $1, $2, $3, $4, $5, $6, $7, $8)
	`, reqUser.ID, "home", c.Param("homeId"), 1, 0, 0, 0, time.Now().Format(time.RFC1123))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMAM005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSMAM005",
			Message: "User can't be added to your home",
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
		logger.WithFields(logger.Fields{"code": "CSMRM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "Members can't be found",
		})
	}

	var reqUser User
	err = DB.QueryRowx("SELECT * FROM users WHERE id=$1", c.Param("userId")).StructScan(&reqUser)

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMRM002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "Members can't be found",
		})
	}

	if reqHome.CreatorID == reqUser.ID {
		logger.WithFields(logger.Fields{"code": "CSMRM003"}).Warnf("%s == %s", reqHome.CreatorID, reqUser.ID)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Home's creator can't be removed",
		})
	}

	_, err = DB.Exec(`
		DELETE FROM permissions WHERE user_id=$1 AND type='home' AND type_id=$2
	`, c.Param("userId"), c.Param("homeId"))

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMRM004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "Member can't be deleted",
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
		logger.WithFields(logger.Fields{"code": "CSMEM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMEM001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Read", "Write", "Manage", "Admin"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSMEM002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMEM002",
			Message: err.Error(),
		})
	}

	read, err := strconv.Atoi(req.Read)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMEM003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMEM003",
			Message: "Read has a wrong value",
		})
	}
	write, err := strconv.Atoi(req.Write)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMEM004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMEM004",
			Message: "Write has a wrong value",
		})
	}
	manage, err := strconv.Atoi(req.Manage)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMEM005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMEM005",
			Message: "Manage has a wrong value",
		})
	}
	admin, err := strconv.Atoi(req.Admin)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMEM006"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSMEM006",
			Message: "Admin has a wrong value",
		})
	}

	_, err = DB.Exec(`
		UPDATE permissions
		SET read=$1, write=$2, manage=$3, admin=$4
		WHERE user_id=$5 AND type=$6 AND type_id=$7
	`, read, write, manage, admin, c.Param("userId"), "home", c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMEM007"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSMEM007",
			Message: "Member can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Member has been updated",
	})
}

// GetRoomMembers route get list of home members
func GetRoomMembers(c echo.Context) error {
	rows, err := DB.Queryx(`
		SELECT * FROM permissions
		JOIN users ON permissions.user_id = users.id
		WHERE (permissions.type=$1 AND permissions.type_id=$2) OR (permissions.type=$3 AND permissions.type_id=$4)
	`, "home", c.Param("homeId"), "room", c.Param("roomId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSMGM001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSMGM001",
			Message: "Members can't be retrieved",
		})
	}

	var permissions []permissionMember
	for rows.Next() {
		var permission permissionMember
		err := rows.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSMGM002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSMGM002",
				Message: "Members can't be retrieved",
			})
		}

		permissions = append(permissions, permission)
	}

	var members []memberRes
	for _, permission := range permissions {
		if permission.Permission.Type == "room" {
			continue
		}

		member := memberRes{
			ID:        permission.User.ID,
			Firstname: permission.User.Firstname,
			Lastname:  permission.User.Lastname,
			Email:     permission.User.Email,
			Birthdate: permission.User.Birthdate,
			CreatedAt: permission.User.CreatedAt,
			Read:      0,
			Write:     0,
			Manage:    0,
			Admin:     0,
		}

		for _, _permission := range permissions {
			if _permission.Permission.Type == "room" && permission.Permission.UserID == _permission.Permission.UserID {
				member.Read = _permission.Permission.Read
				member.Write = _permission.Permission.Write
				member.Manage = _permission.Permission.Manage
				member.Admin = _permission.Permission.Admin
				break
			}
		}
		members = append(members, member)
	}

	totalMembers := strconv.Itoa(len(members))
	c.Response().Header().Set("Content-Range", "0-"+totalMembers+"/"+totalMembers)
	return c.JSON(http.StatusOK, members)
}
