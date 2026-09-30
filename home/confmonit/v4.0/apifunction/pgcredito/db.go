package pgcredito

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	conn *sql.DB
	once sync.Once
	openErr error
)

func Abrir(dsn string) error {
	if dsn == "" {
		return nil
	}
	once.Do(func() {
		conn, openErr = sql.Open("postgres", dsn)
		if openErr != nil {
			return
		}
		conn.SetMaxOpenConns(15)
		conn.SetMaxIdleConns(3)
		conn.SetConnMaxLifetime(30 * time.Minute)
		openErr = conn.Ping()
		if openErr == nil {
			log.Println("apifunction postgres (credito atendimento FP) ok")
		}
	})
	return openErr
}

func DB() (*sql.DB, error) {
	if conn == nil {
		return nil, errPostgresNaoConfigurado
	}
	return conn, nil
}

func Configurado() bool {
	return conn != nil
}
