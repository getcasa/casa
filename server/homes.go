package server

import (
	"net/http"
	"reflect"

	"github.com/ItsJimi/casa/logger"
	"github.com/ItsJimi/casa/utils"
	"github.com/labstack/echo"
)

type addHomeReq struct {
	Name     string
	Address  string
	WifiSSID string
}

// AddHome route create and add user to an home
func AddHome(c echo.Context) error {
	req := new(addHomeReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSHAH001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSRAR001",
			Message: "Wrong parameters",
		})
	}

	if err := utils.MissingFields(c, reflect.ValueOf(req).Elem(), []string{"Name"}); err != nil {
		logger.WithFields(logger.Fields{"code": "CSHAH002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSHAH002",
			Message: err.Error(),
		})
	}

	user := c.Get("user").(User)

	row, err := DB.Query("INSERT INTO homes (id, name, address, creator_id) VALUES (generate_ulid(), $1, $2, $3) RETURNING id;", req.Name, req.Address, user.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHAH003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSHAH003",
			Message: "Home can't be added",
		})
	}
	var homeID string
	row.Next()
	err = row.Scan(&homeID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHAH004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSHAH004",
			Message: "Home can't be added",
		})
	}

	newPermission := Permission{
		UserID: user.ID,
		Type:   "home",
		TypeID: homeID,
		Read:   true,
		Write:  true,
		Manage: true,
		Admin:  true,
	}
	_, err = DB.NamedExec("INSERT INTO permissions (id, user_id, type, type_id, read, write, manage, admin) VALUES (generate_ulid(), :user_id, :type, :type_id, :read, :write, :manage, :admin)", newPermission)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHAH005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSHAH005",
			Message: "Home can't be added",
		})
	}

	return c.JSON(http.StatusCreated, MessageResponse{
		Message: homeID,
	})
}

// UpdateHome route update home
func UpdateHome(c echo.Context) error {
	req := new(addHomeReq)
	if err := c.Bind(req); err != nil {
		logger.WithFields(logger.Fields{"code": "CSHUH001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSHUH001",
			Message: "Wrong parameters",
		})
	}

	_, err := DB.Exec("UPDATE homes SET name=COALESCE($1, name), address=COALESCE($2, address), wifi_ssid=COALESCE($3, wifi_ssid) WHERE id=$4", utils.NewNullString(req.Name), utils.NewNullString(req.Address), utils.NewNullString(req.WifiSSID), c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHUH005"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSHUH005",
			Message: "Home can't be updated",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Home updated",
	})
}

// DeleteHome route delete home
func DeleteHome(c echo.Context) error {
	_, err := DB.Exec("DELETE FROM homes WHERE id=$1", c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHDH001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSHDH001",
			Message: "Home can't be deleted",
		})
	}
	_, err = DB.Exec("DELETE FROM permissions WHERE type=$1 AND type_id=$2", "home", c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHDH002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSHDH002",
			Message: "Home can't be deleted",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Home deleted",
	})
}

type permissionHome struct {
	Permission
	User
	HomeID        string `db:"h_id"`
	HomeName      string `db:"h_name"`
	HomeAddress   string `db:"h_address"`
	HomeCreatedAt string `db:"h_createdat"`
}

type homeRes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	CreatedAt string `json:"created_at"`
	Creator   User   `json:"creator"`
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Manage    bool   `json:"manage"`
	Admin     bool   `json:"admin"`
}

// GetHomes route get list of user homes
func GetHomes(c echo.Context) error {
	user := c.Get("user").(User)

	rows, err := DB.Queryx(`
		SELECT permissions.*, users.*, homes.id as h_id, homes.name AS h_name, homes.address AS h_address, homes.created_at AS h_createdat FROM permissions
		JOIN homes ON permissions.type_id = homes.id
		JOIN users ON homes.creator_id = users.id
		WHERE permissions.type=$1 AND permissions.user_id=$2
	`, "home", user.ID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHGHS001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSHGHS001",
			Message: "Homes can't be retrieved",
		})
	}
	defer rows.Close()

	var homes []homeRes
	for rows.Next() {
		var permission permissionHome
		err := rows.StructScan(&permission)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSHGHS002"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSHGHS002",
				Message: "Homes can't be retrieved",
			})
		}
		homes = append(homes, homeRes{
			ID:        permission.HomeID,
			Name:      permission.HomeName,
			Address:   permission.HomeAddress,
			CreatedAt: permission.HomeCreatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: homes,
	})
}

// GetHome route get specific home with id
func GetHome(c echo.Context) error {
	user := c.Get("user").(User)

	row := DB.QueryRowx(`
		SELECT permissions.*, users.*, homes.id as h_id, homes.name AS h_name, homes.address AS h_address, homes.created_at AS h_createdat FROM permissions
		JOIN homes ON permissions.type_id = homes.id
		JOIN users ON homes.creator_id = users.id
		WHERE type=$1 AND type_id=$2 AND user_id=$3
	`, "home", c.Param("homeId"), user.ID)

	if row == nil {
		logger.WithFields(logger.Fields{"code": "CSHGH001"}).Errorf("QueryRowx: Select error")
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "CSHGH001",
			Message: "Home can't be found",
		})
	}

	var permission permissionHome
	err := row.StructScan(&permission)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSHGH002"}).Errorf("QueryRowx: Select error")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSHGH002",
			Message: "Home can't be found",
		})
	}

	return c.JSON(http.StatusOK, DataReponse{
		Data: homeRes{
			ID:        permission.HomeID,
			Name:      permission.HomeName,
			Address:   permission.HomeAddress,
			CreatedAt: permission.HomeCreatedAt,
			Creator:   permission.User,
			Read:      permission.Permission.Read,
			Write:     permission.Permission.Write,
			Manage:    permission.Permission.Manage,
			Admin:     permission.Permission.Admin,
		},
	})
}
