package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Conn *sql.DB

func Abrir(dsn string) {
	var err error
	Conn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("mysql open: %v", err)
	}
	Conn.SetMaxOpenConns(20)
	Conn.SetMaxIdleConns(5)
	Conn.SetConnMaxLifetime(30 * time.Minute)
	if err = Conn.Ping(); err != nil {
		log.Fatalf("mysql ping: %v", err)
	}
	log.Println("apifunction mysql ok")
}
