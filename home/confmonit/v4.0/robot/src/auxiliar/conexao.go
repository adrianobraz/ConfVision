package auxiliar

import (
	"database/sql"
	"robot/src/config"
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver
)

func Conectar() (*sql.DB, error) {
	db, erro := sql.Open("mysql", config.StringConexao)
	if erro != nil {
		return nil, erro
	}
	db.SetConnMaxLifetime(time.Nanosecond)

	// Verifica se a conexão esta operante
	if erro = db.Ping(); erro != nil {
		db.Close()
		return nil, erro
	}
	return db, nil
}
