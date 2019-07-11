package server

import (
	"database/sql"
	"log"

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
}

// Token structure in database
type Token struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"userId"`
	Type      string `db:"type" json:"type"`
	IP        string `db:"ip" json:"ip"`
	UserAgent string `db:"user_agent" json:"userAgent"`
	Read      int    `db:"read" json:"read"`
	Write     int    `db:"write" json:"write"`
	Manage    int    `db:"manage" json:"manage"`
	Admin     int    `db:"admin" json:"admin"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	ExpireAt  string `db:"expire_at" json:"expireAt"`
}

// Gateway structure in database
type Gateway struct {
	ID        string         `db:"id" json:"id"`
	HomeID    sql.NullString `db:"home_id" json:"homeId"`
	Name      sql.NullString `db:"name" json:"name"`
	Model     string         `db:"model" json:"model"`
	CreatedAt string         `db:"created_at" json:"createdAt"`
	CreatorID sql.NullString `db:"creator_id" json:"creatorId"`
}

// Home structure in database
type Home struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Address   string `db:"address" json:"address"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	CreatorID string `db:"creator_id" json:"creatorId"`
}

// Room structure in database
type Room struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	HomeID    string `db:"home_id" json:"homeId"`
	CreatedAt string `db:"created_at" json:"createdAt"`
	CreatorID string `db:"creator_id" json:"creatorId"`
}

// Device structure in database
type Device struct {
	ID         string `db:"id" json:"id"`
	GatewayID  string `db:"gateway_id" json:"gatewayId"`
	Name       string `db:"name" json:"name"`
	PhysicalID string `db:"physical_id" json:"physicalId"`
	RoomID     string `db:"room_id" json:"roomId"`
	CreatedAt  string `db:"created_at" json:"createdAt"`
	CreatorID  string `db:"creator_id" json:"creatorId"`
}

// Permission structure in database
type Permission struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"userId"`
	Type      string `db:"type" json:"type"`
	TypeID    string `db:"type_id" json:"typeId"`
	Read      int    `db:"read" json:"read"`
	Write     int    `db:"write" json:"write"`
	Manage    int    `db:"manage" json:"manage"`
	Admin     int    `db:"admin" json:"admin"`
	UpdatedAt string `db:"updated_at" json:"updatedAt"`
}

// DB define the database object
var DB *sqlx.DB

// InitDB check and create tables
func InitDB() {
	var err error
	connStr := "postgres://postgres:password@localhost/?sslmode=disable"
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Panic(err)
	}

	_, err = db.Exec("CREATE database casadb")
	if err != nil {
		log.Panic(err)
	}

	db.Close()
}

// StartDB start the database to use it in server
func StartDB() {
	var err error
	connStr := "postgres://postgres:password@localhost/casadb?sslmode=disable"
	DB, err = sqlx.Open("postgres", connStr)
	if err != nil {
		log.Panic(err)
	}

	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS users (id BYTEA PRIMARY KEY, firstname TEXT, lastname TEXT, email TEXT, password TEXT, birthdate TEXT, created_at TEXT)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS tokens (id BYTEA PRIMARY KEY, user_id BYTEA, type TEXT, ip TEXT, user_agent TEXT, read INTEGER, write INTEGER, manage INTEGER, admin INTEGER, created_at TEXT, expire_at TEXT)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS gateways (id BYTEA PRIMARY KEY, home_id BYTEA, name TEXT, model TEXT, created_at TEXT, creator_id BYTEA)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS homes (id BYTEA PRIMARY KEY, name TEXT, address TEXT, created_at TEXT, creator_id BYTEA)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS rooms (id BYTEA PRIMARY KEY, name TEXT, home_id BYTEA, created_at TEXT, creator_id BYTEA)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS devices (id BYTEA PRIMARY KEY, name TEXT, physical_id TEXT, room_id BYTEA, created_at TEXT, creator_id BYTEA)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS permissions (id BYTEA PRIMARY KEY, user_id BYTEA, type TEXT, type_id BYTEA, read INTEGER, write INTEGER, manage INTEGER, admin INTEGER, updated_at TEXT)")
	if err != nil {
		log.Panic(err)
	}
}
