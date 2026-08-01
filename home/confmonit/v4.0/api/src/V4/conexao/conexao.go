package connV4

import (
	"api/src/V4/config"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // Driver
)

func Conectar() (*sql.DB, error) {
	bd, err := masterConectar() // Tenta conexao Master

	if err != nil { // Caso conexão Master falhe ele aciona a Slave

		return slaveConectar() // Tenta conexao Slave
	}

	return bd, err
}

// MasterConectar abre uma conexão com o servidor master
func masterConectar() (*sql.DB, error) {

	db, erro := sql.Open("mysql", config.ConexaoV4Master)
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

func slaveConectar() (*sql.DB, error) {

	db, erro := sql.Open("mysql", config.ConexaoV4Slave)
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
