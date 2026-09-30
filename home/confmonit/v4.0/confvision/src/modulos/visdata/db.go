package visdata

import (
	"confvision/src/conexao"
	"confvision/src/config"
	"database/sql"
	"errors"
	"sync"
)

var (
	ErrNotHandled  = errors.New("visdata: rota nao tratada")
	ErrPostgresOff = errors.New("visdata: postgres desabilitado")

	dbMu   sync.Mutex
	dbInst *sql.DB
	dbErr  error
)

func DB() (*sql.DB, error) {
	if !config.VisPostgresEnabled {
		return nil, ErrPostgresOff
	}

	dbMu.Lock()
	defer dbMu.Unlock()

	if dbInst == nil {
		dbInst, dbErr = conexao.ConectarPostgres()
		return dbInst, dbErr
	}
	if dbErr != nil {
		return nil, dbErr
	}

	if err := dbInst.Ping(); err != nil {
		_ = dbInst.Close()
		dbInst = nil
		dbInst, dbErr = conexao.ConectarPostgres()
	}
	return dbInst, dbErr
}

func Ready() bool {
	if !config.VisPostgresEnabled {
		return false
	}
	db, err := DB()
	return err == nil && db != nil
}
