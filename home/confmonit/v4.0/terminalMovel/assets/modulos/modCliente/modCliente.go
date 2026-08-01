package modCliente

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
		Uri:    "/modCliente/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type Usuarios struct {
		Numero string `json:"codigo"`
		Nome   string `json:"nome"`
	}

	type Setores struct {
		Setor string `json:"setor"`
		Nome  string `json:"nome"`
	}

	type objeto struct {
		IdDispositivo string     `json:"idDispositivo"`
		IdCliente     string     `json:"idCliente"`
		Conta         string     `json:"conta"`
		Nome          string     `json:"nome"`
		Telefone      string     `json:"telefone"`
		Celular       string     `json:"celular"`
		Endereco      string     `json:"endereco"`
		Bairro        string     `json:"bairro"`
		Cidade        string     `json:"cidade"`
		NomeDisp      string     `json:"nomeDisp"`
		Armado        string     `json:"armado"`
		RazaoSocial   string     `json:"razaoSocial"`
		Usuarios      []Usuarios `json:"usuarios"`
		Setores       []Setores  `json:"setores"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
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
			dispositivos.ID_Dispositivo,
			clientes.ID_Cliente,
			dispositivos.contaAlarme,
			clientes.nome,
			clientes.telefone,
			clientes.celular,
			clientes.endereco,
			clientes.bairro,
			clientes.cidade,
			dispositivos.nome,
			dispositivos.armado,
			franqueado.razaoSocial,
			setorAlarme.numero,
			setorAlarme.nome,
			usuariosAlarme.codigo,
			usuariosAlarme.nome

		FROM dispositivos			

		LEFT JOIN clientes
		ON clientes.ID_Cliente = dispositivos.ID_Cliente

		LEFT JOIN franqueado
		ON franqueado.ID_Franqueado = dispositivos.ID_Franqueado

		LEFT JOIN setorAlarme
		ON setorAlarme.ID_Dispositivo= dispositivos.ID_Dispositivo

		LEFT JOIN usuariosAlarme
		ON usuariosAlarme.ID_Dispositivo = dispositivos.ID_Dispositivo
			
		WHERE dispositivos.ID_Dispositivo = ?
	`, obj.IdDispositivo)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	for tab.Next() {
		var (
			usuarios      Usuarios
			setores       Setores
			idDispositivo sql.NullString
			idCliente     sql.NullString
			conta         sql.NullString
			nome          sql.NullString
			telefone      sql.NullString
			celular       sql.NullString
			endereco      sql.NullString
			bairro        sql.NullString
			cidade        sql.NullString
			nomeDisp      sql.NullString
			armado        sql.NullString
			razaoSocial   sql.NullString
			numeroSetor   sql.NullString
			nomeSetor     sql.NullString
			codigoUsuario sql.NullString
			nomeUsuario   sql.NullString
		)

		if erro := tab.Scan(
			&idDispositivo,
			&idCliente,
			&conta,
			&nome,
			&telefone,
			&celular,
			&endereco,
			&bairro,
			&cidade,
			&nomeDisp,
			&armado,
			&razaoSocial,
			&numeroSetor,
			&nomeSetor,
			&codigoUsuario,
			&nomeUsuario,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		obj.IdDispositivo = idDispositivo.String
		obj.IdCliente = idCliente.String
		obj.Conta = conta.String
		obj.Nome = nome.String
		obj.Telefone = telefone.String
		obj.Celular = celular.String
		obj.Endereco = endereco.String
		obj.Bairro = bairro.String
		obj.Cidade = cidade.String
		obj.NomeDisp = nomeDisp.String
		obj.Armado = armado.String
		obj.RazaoSocial = razaoSocial.String
		if numeroSetor.String != "" {
			setores.Setor = numeroSetor.String
			setores.Nome = nomeSetor.String
			obj.Setores = append(obj.Setores, setores)
		}

		if codigoUsuario.String != "" {
			usuarios.Numero = codigoUsuario.String
			usuarios.Nome = nomeUsuario.String
			obj.Usuarios = append(obj.Usuarios, usuarios)
		}

	}

	aux.RespostaJsonDados(w, http.StatusOK, obj)

}
