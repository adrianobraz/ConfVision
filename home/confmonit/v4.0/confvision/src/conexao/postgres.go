package conexao

import (
	"confvision/src/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConectarPostgres() (*sql.DB, error) {
	if config.ConexaoPostgres == "" {
		return nil, fmt.Errorf("conexao postgres nao configurada (POSTGRES_URL)")
	}

	db, err := sql.Open("pgx", config.ConexaoPostgres)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
