package visdata

import (
	"confvision/src/conexao"
	"confvision/src/config"
	"database/sql"
	"errors"
	"sync"
)

var (
	ErrNotHandled = errors.New("visdata: rota nao tratada")
	ErrPostgresOff = errors.New("visdata: postgres desabilitado")

	dbOnce sync.Once
	dbInst *sql.DB
	dbErr  error
)

func DB() (*sql.DB, error) {
	if !config.VisPostgresEnabled {
		return nil, ErrPostgresOff
	}
	dbOnce.Do(func() {
		dbInst, dbErr = conexao.ConectarPostgres()
	})
	return dbInst, dbErr
}

func Ready() bool {
	if !config.VisPostgresEnabled {
		return false
	}
	_, err := DB()
	return err == nil
}
