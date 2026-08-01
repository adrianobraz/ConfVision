package conexao

import (
	"confvision/src/config"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func Conectar() (*sql.DB, error) {
	if config.ConexaoMySQL == "" {
		return nil, fmt.Errorf("conexao mysql nao configurada")
	}

	db, erro := sql.Open("mysql", config.ConexaoMySQL)
	if erro != nil {
		return nil, erro
	}

	if erro = db.Ping(); erro != nil {
		db.Close()
		return nil, erro
	}

	return db, nil
}
