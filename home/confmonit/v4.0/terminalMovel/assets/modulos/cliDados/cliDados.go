package cliDados

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/cliDados/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
	{
		Uri:    "/cliDados/buscarDispositivosById",
		Metodo: http.MethodPost,
		Funcao: buscarDispositivoById,
	},
	{
		Uri:    "/cliDados/buscarDispositivoByCliente",
		Metodo: http.MethodPost,
		Funcao: buscarDispositivoByCliente,
	},
}

type usuarios struct {
	Numero string `json:"codigo"`
	Nome   string `json:"nome"`
}

type setores struct {
	Setor    string `json:"setor"`
	Camera   string `json:"camera"`
	Particao string `json:"particao"`
	Nome     string `json:"nome"`
}

type dispositivo struct {
	IdCliente      string     `json:"idCliente"`
	IdDispositivo  string     `json:"idDispositivo"`
	Nome           string     `json:"nome"`
	Particao       string     `json:"particao"`
	Conta          string     `json:"conta"`
	NomeDisp       string     `json:"nomeDisp"`
	NomeFabricante string     `json:"nomeFabricante"`
	Armado         string     `json:"armado"`
	Senha          string     `json:"senha"`
	RazaoSocial    string     `json:"razaoSocial"`
	CodBenuvem     string     `json:"codFranq"`
	Usuarios       []usuarios `json:"usuarios"`
	Setores        []setores  `json:"setores"`
}

type cliente struct {
	IdCliente   string `json:"idCliente"`
	Nome        string `json:"nome"`
	Endereco    string `json:"endereco"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	Telefone1   string `json:"telefone1"`
	Telefone2   string `json:"telefone2"`
}

func buscarDados(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj cliente
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT
			cliente.ID_Cliente,
			cliente.Nome,
			cliente.Endereco,
			cliente.Complemento,
			cliente.Bairro,
			cliente.Cidade, 	
			cliente.Telefone1,
			cliente.Telefone2
			
		FROM cliente
			
		WHERE cliente.ID_Cliente = ?
	`, obj.IdCliente)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	for tab.Next() {
		var (
			idCliente   sql.NullString
			nome        sql.NullString
			endereco    sql.NullString
			complemento sql.NullString
			bairro      sql.NullString
			cidade      sql.NullString
			telefone2   sql.NullString
			telefone1   sql.NullString
		)

		if erro := tab.Scan(
			&idCliente,
			&nome,
			&endereco,
			&complemento,
			&bairro,
			&cidade,
			&telefone1,
			&telefone2,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		obj.IdCliente = idCliente.String
		obj.Nome = nome.String
		obj.Endereco = endereco.String
		obj.Complemento = complemento.String
		obj.Bairro = bairro.String
		obj.Cidade = cidade.String
		obj.Telefone1 = telefone1.String
		obj.Telefone2 = telefone2.String

	}
	aux.RespostaJsonDados(w, http.StatusOK, obj)
}

func buscarDispositivoById(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var disp dispositivo
	if erro := json.Unmarshal(body, &disp); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			dispositivo.Particao,
			dispositivo.Conta,
			dispositivo.Armado,
			dispositivo.Senha,

			franqueado.RazaoSocial,
			franqueado.CodBenuvem,
			IFNULL(fabricantes.Nome, '')

		FROM dispositivo

		LEFT JOIN cliente
		ON cliente.ID_Cliente = dispositivo.ID_Cliente

		LEFT JOIN franqueado
		ON franqueado.ID_Franqueado = cliente.ID_Franqueado

		LEFT JOIN fabricantes
		ON fabricantes.ID_Fabricante = dispositivo.ID_Fabricante

		WHERE dispositivo.ID_Dispositivo = ?
	`, disp.IdDispositivo)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	if tab.Next() {
		var (
			particao       sql.NullString
			razaoSocial    sql.NullString
			codBenuvem     sql.NullString
			nomeFabricante sql.NullString
		)

		if erro := tab.Scan(
			&disp.IdDispositivo,
			&disp.Nome,
			&particao,
			&disp.Conta,
			&disp.Armado,
			&disp.Senha,
			&razaoSocial,
			&codBenuvem,
			&nomeFabricante,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		disp.Particao = fmt.Sprintf("%02s", particao.String)
		disp.RazaoSocial = razaoSocial.String
		disp.CodBenuvem = codBenuvem.String
		disp.NomeFabricante = nomeFabricante.String

		if err := buscaSetor(db, disp.IdDispositivo, &disp.Setores); err != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		if err := buscUsuarios(db, disp.IdDispositivo, &disp.Usuarios); err != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		aux.RespostaJsonDados(w, http.StatusOK, disp)
		return
	}

	aux.RespostaJsonVazio(w)
}

func buscarDispositivoByCliente(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var disp dispositivo
	if erro := json.Unmarshal(body, &disp); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT
			dispositivo.ID_Dispositivo,
			dispositivo.Nome	

		FROM dispositivo
		
		WHERE dispositivo.ID_Cliente = ?
	`, disp.IdCliente)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var listaDisp []dispositivo

	for tab.Next() {
		var item dispositivo

		if erro := tab.Scan(
			&item.IdDispositivo,
			&item.Nome,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		listaDisp = append(listaDisp, item)
	}

	if len(listaDisp) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, listaDisp)
	} else {
		aux.RespostaJsonVazio(w)
	}

}

func buscaSetor(db *sql.DB, idDisp string, lista *[]setores) error {
	// Setores
	tab, err := db.Query(`
		SELECT
			setorAlarme.Numero,
			setorAlarme.Camera,
			setorAlarme.Particao,
			setorAlarme.Nome
		FROM setorAlarme

		WHERE setorAlarme.ID_Dispositivo = ?

	`, idDisp)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item setores

		if err := tab.Scan(
			&item.Setor,
			&item.Camera,
			&item.Particao,
			&item.Nome,
		); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func buscUsuarios(db *sql.DB, idDisp string, lista *[]usuarios) error {
	tab, err := db.Query(`
		SELECT
			usuariosAlarme.Codigo,
			usuariosAlarme.Nome

		FROM usuariosAlarme

		WHERE usuariosAlarme.ID_Dispositivo = ?

	`, idDisp)
	if err != nil {
		return err
	}

	for tab.Next() {
		var item usuarios

		if err := tab.Scan(
			&item.Numero,
			&item.Nome,
		); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}
