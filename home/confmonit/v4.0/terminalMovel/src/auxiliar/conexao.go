package aux

import (
	"database/sql"
	"fmt"
	"terminal/src/setup"

	_ "github.com/go-sql-driver/mysql" // Driver
)

// Conectar abre a conexão com o banco de dados e a retorna
func Conectar() (*sql.DB, error) {

	db, erro := sql.Open("mysql", setup.StringConexaoBanco)
	if erro != nil {
		return nil, erro
	}
	//db.SetConnMaxLifetime(time.Nanosecond)
	if erro = db.Ping(); erro != nil {
		db.Close()
		fmt.Println(erro)
		return nil, erro
	}

	return db, nil

}
