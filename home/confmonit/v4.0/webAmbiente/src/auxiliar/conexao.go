package auxiliar

import (
	"database/sql"
	"fmt"
	"webAmbiente/src/config"

	_ "github.com/go-sql-driver/mysql"
)

func Conectar() (*sql.DB, error) {
	db, err := sql.Open("mysql", config.StringConexaoBanco)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		fmt.Println(err)
		return nil, err
	}
	return db, nil
}
