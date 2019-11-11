package server

import (
	"io/ioutil"
	"net/http"

	"github.com/ItsJimi/casa/logger"
	"github.com/jmoiron/sqlx"
)

// User structure in database
type User struct {
	ID        string `db:"id" json:"id"`
	Firstname string `db:"firstname" json:"firstname"`
	Lastname  string `db:"lastname" json:"lastname"`
	Email     string `db:"email" json:"email"`
	Password  string `db:"password" json:"-"`
	Birthdate string `db:"birthdate" json:"birthdate"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
}

// Token structure in database
type Token struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"userId"`
	Type      string `db:"type" json:"type"`
	IP        string `db:"ip" json:"ip"`
	UserAgent string `db:"user_agent" json:"userAgent"`
	Read      bool   `db:"read" json:"read"`
	Write     bool   `db:"write" json:"write"`
	Manage    bool   `db:"manage" json:"manage"`
	Admin     bool   `db:"admin" json:"admin"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
	ExpireAt  string `db:"expire_at" json:"expireAt"`
}

// Gateway structure in database
type Gateway struct {
	ID        string `db:"id" json:"id"`
	HomeID    string `db:"home_id" json:"homeId"`
	Name      string `db:"name" json:"name"`
	Model     string `db:"model" json:"model"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
	CreatorID string `db:"creator_id" json:"creatorId"`
}

// Plugin structure in database
type Plugin struct {
	ID        string `db:"id" json:"id"`
	GatewayID string `db:"gateway_id" json:"gatewayId"`
	Name      string `db:"name" json:"name"`
	Config    string `db:"config" json:"config"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
}

// Home structure in database
type Home struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Address   string `db:"address" json:"address"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
	CreatorID string `db:"creator_id" json:"creatorId"`
}

// Room structure in database
type Room struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Icon      string `db:"icon" json:"icon"`
	HomeID    string `db:"home_id" json:"homeId"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
	CreatorID string `db:"creator_id" json:"creatorId"`
}

// Device structure in database
type Device struct {
	ID           string `db:"id" json:"id"`
	GatewayID    string `db:"gateway_id" json:"gatewayId"`
	Name         string `db:"name" json:"name"`
	Icon         string `db:"icon" json:"icon"`
	PhysicalID   string `db:"physical_id" json:"physicalId"`
	PhysicalName string `db:"physical_name" json:"physicalName"`
	Config       string `db:"config" json:"config"`
	Plugin       string `db:"plugin" json:"plugin"`
	RoomID       string `db:"room_id" json:"roomId"`
	CreatedAt    string `db:"created_at" json:"createdAt"`
	UpdatedAt    string `db:"updated_at" json:"updatedAt"`
	CreatorID    string `db:"creator_id" json:"creatorId"`
}

// Permission structure in database
type Permission struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"userId"`
	Type      string `db:"type" json:"type"` //home, room, device
	TypeID    string `db:"type_id" json:"typeId"`
	Read      bool   `db:"read" json:"read"`
	Write     bool   `db:"write" json:"write"`
	Manage    bool   `db:"manage" json:"manage"`
	Admin     bool   `db:"admin" json:"admin"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
}

// Automation struct in database
type Automation struct {
	ID              string   `db:"id" json:"id"`
	HomeID          string   `db:"home_id" json:"homeId"`
	Name            string   `db:"name" json:"name"`
	Trigger         []string `db:"trigger" json:"trigger"`
	TriggerKey      []string `db:"trigger_key" json:"triggerKey"`
	TriggerValue    []string `db:"trigger_value" json:"triggerValue"`
	TriggerOperator []string `db:"trigger_operator" json:"triggerOperator"`
	Action          []string `db:"action" json:"action"`
	ActionCall      []string `db:"action_call" json:"actionCall"`
	ActionValue     []string `db:"action_value" json:"actionValue"`
	Status          bool     `db:"status" json:"status"`
	CreatedAt       string   `db:"created_at" json:"createdAt"`
	UpdatedAt       string   `db:"updated_at" json:"updatedAt"`
	CreatorID       string   `db:"creator_id" json:"creatorId"`
}

// Datas struct in database
type Datas struct {
	ID        string  `db:"id" json:"id"`
	DeviceID  string  `db:"device_id" json:"deviceId"`
	Field     string  `db:"field" json:"field"`
	ValueNbr  float64 `db:"value_nbr" json:"valueNbr"`
	ValueStr  string  `db:"value_str" json:"valueStr"`
	ValueBool bool    `db:"value_bool" json:"valueBool"`
	CreatedAt string  `db:"created_at" json:"createdAt"`
}

// DB define the database object
var DB *sqlx.DB

// InitDB check and create tables
func InitDB() {
	var err error
	connStr := "postgres://postgres:password@localhost/?sslmode=disable"
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB001"}).Panicf("%s", err.Error())
	}

	_, err = db.Exec("CREATE database casadb")
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB002"}).Errorf("%s", err.Error())
	}

	db.Close()

	connStr = "postgres://postgres:password@localhost/casadb?sslmode=disable"
	db, err = sqlx.Open("postgres", connStr)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB003"}).Panicf("%s", err.Error())
	}

	file, err := ioutil.ReadFile("database.sql")
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB004"}).Panicf("%s", err.Error())
	}

	_, err = db.Exec(string(file))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB005"}).Errorf("%s", err.Error())
	}

	resp, err := http.Get("https://raw.githubusercontent.com/geckoboard/pgulid/master/pgulid.sql")
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB006"}).Panicf("%s", err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB007"}).Panicf("%s", err.Error())
	}

	_, err = db.Exec(string(body))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIDB008"}).Errorf("%s", err.Error())
	}

	db.Close()
}

// StartDB start the database to use it in server
func StartDB() {
	var err error
	connStr := "postgres://postgres:password@localhost/casadb?sslmode=disable"
	DB, err = sqlx.Open("postgres", connStr)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDSDB001"}).Panicf("%s", err.Error())
	}
}
