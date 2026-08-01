package erroConexao

import (
	connV4 "api/src/V4/conexao"
	"time"
)

type ErroConexao struct {
	ID_ErroConexao string `json:"idErro"`
	Codigo         string `json:"codigo"`
	Serial         string `json:"serial"`
	Imei           string `json:"imei"`
	IPRemoto       string `json:"ipRemoto"`
	Modelo         string `json:"modelo"`
	Versao         string `json:"versao"`
	Protocolo      string `json:"protocolo"`
	Fabricante     string `json:"fabricante"`
	Conta          string `json:"conta"`
	DataTentativa  string `json:"dataTentativa"`
	Comentario     string `json:"comentario"`
}

func (ec *ErroConexao) listar(lista *[]ErroConexao) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
		erroConexao.ID_ErroConexao,
		erroConexao.Codigo,
		erroConexao.Serial,
		erroConexao.Imei,
		erroConexao.IP_Remoto,
		erroConexao.Modelo,
		erroConexao.Versao,
		erroConexao.Protocolo,
		erroConexao.Fabricante,
		erroConexao.Conta,
		erroConexao.DataTentativa,
		erroConexao.Comentario
		FROM erroConexao	
	`)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item ErroConexao

		if err := tab.Scan(
			&item.ID_ErroConexao,
			&item.Codigo,
			&item.Serial,
			&item.Imei,
			&item.IPRemoto,
			&item.Modelo,
			&item.Versao,
			&item.Protocolo,
			&item.Fabricante,
			&item.Conta,
			&item.DataTentativa,
			&item.Comentario,
		); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

func (ec *ErroConexao) limpar() error {

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	referencia := time.Now().Add(-30 * time.Minute).Format("2006-01-02 15:04:05")

	stm, err := db.Prepare(`DELETE FROM erroConexao WHERE DataTentativa < ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(referencia); err != nil {
		return err
	}

	return nil
}
