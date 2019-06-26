package server

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// User structure in database
type User struct {
	ID        string `db:"id"`
	Firstname string `db:"firstname"`
	Lastname  string `db:"lastname"`
	Email     string `db:"email"`
	Password  string `db:"password"`
	Birthdate string `db:"birthdate"`
	CreatedAt string `db:"created_at"`
}

// Token structure in database
type Token struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	Type      string `db:"type"`
	IP        string `db:"ip"`
	UserAgent string `db:"user_agent"`
	Read      int    `db:"read"`
	Write     int    `db:"write"`
	Manage    int    `db:"manage"`
	Admin     int    `db:"admin"`
	CreatedAt string `db:"created_at"`
	ExpireAt  string `db:"expire_at"`
}

// Home structure in database
type Home struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Address   string `db:"address"`
	CreatedAt string `db:"created_at"`
	CreatorID string `db:"creator_id"`
}

// Room structure in database
type Room struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	HomeID    string `db:"home_id"`
	CreatedAt string `db:"created_at"`
	CreatorID string `db:"creator_id"`
}

// Device structure in database
type Device struct {
	ID         string `db:"id"`
	Name       string `db:"name"`
	PhysicalID string `db:"physical_id"`
	RoomID     string `db:"room_id"`
	CreatedAt  string `db:"created_at"`
	CreatorID  string `db:"creator_id"`
}

// Permission structure in database
type Permission struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	Type      string `db:"type"`
	TypeID    string `db:"type_id"`
	Read      int    `db:"read"`
	Write     int    `db:"write"`
	Manage    int    `db:"manage"`
	Admin     int    `db:"admin"`
	UpdatedAt string `db:"updated_at"`
}

// DB define the database object
var DB *sqlx.DB

// InitDB check and create tables
func InitDB() {
	var err error
	DB, err = sqlx.Open("sqlite3", "./casa.db")
	if err != nil {
		log.Panic(err)
	}

	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS user (id BLOB PRIMARY KEY, firstname TEXT, lastname TEXT, email TEXT, password TEXT, birthdate TEXT, created_at TEXT)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS token (id BLOB PRIMARY KEY, user_id BLOB, type TEXT, ip TEXT, user_agent TEXT, read INTEGER, write INTEGER, manage INTEGER, admin INTEGER, created_at TEXT, expire_at TEXT)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS home (id BLOB PRIMARY KEY, name TEXT, address TEXT, created_at TEXT, creator_id BLOB)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS room (id BLOB PRIMARY KEY, name TEXT, home_id BLOB, created_at TEXT, creator_id BLOB)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS device (id BLOB PRIMARY KEY, name TEXT, physical_id TEXT, room_id BLOB, created_at TEXT, creator_id BLOB)")
	if err != nil {
		log.Panic(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS permission (id BLOB PRIMARY KEY, user_id BLOB, type TEXT, type_id BLOB, read INTEGER, write INTEGER, manage INTEGER, admin INTEGER, updated_at TEXT)")
	if err != nil {
		log.Panic(err)
	}
}
