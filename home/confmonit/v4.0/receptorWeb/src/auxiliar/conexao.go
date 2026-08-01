package auxiliar

import (
	"database/sql"
	"fmt"
	"receptorWeb/src/config"
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver
)

// Conectar abre a conexão com o banco de dados e a retorna
func Conectar() (*sql.DB, error) {

	db, erro := sql.Open("mysql", config.StringConexaoBanco)
	if erro != nil {
		return nil, erro
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Second * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if erro = db.Ping(); erro != nil {
		db.Close()
		fmt.Println(erro)
		return nil, erro
	}

	return db, nil

}
