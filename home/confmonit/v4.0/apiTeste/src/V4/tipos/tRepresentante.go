package tiposV4

import "database/sql"

type TRepresentante struct {
	ID_Representante  string `json:"idRepresentante"`
	ID_UsuarioMaster  string `json:"idUsuarioMaster"`
	ID_Pacote         string `json:"idPacote"`
	RazaoSocial       string `json:"razaoSocial"`
	NomeFantasia      string `json:"nomeFantasia"`
	Cnpj              string `json:"cnpj"`
	InscricaoEstadual string `json:"inscricaoEstadual"`
	Cep               string `json:"cep"`
	Endereco          string `json:"endereco"`
	Complemento       string `json:"complemento"`
	Bairro            string `json:"bairro"`
	Cidade            string `json:"cidade"`
	Uf                string `json:"uf"`
	DataCadastro      string `json:"dataCadastro"`
	DataCancelamento  string `json:"dataCancelamento"`
	Informativo       string `json:"informativo"`
	Ativo             string `json:"ativo"`
}

type SRepresentante struct {
	ID_Representante  sql.NullString
	ID_UsuarioMaster  sql.NullString
	ID_Pacote         sql.NullString
	RazaoSocial       sql.NullString
	NomeFantasia      sql.NullString
	Cnpj              sql.NullString
	InscricaoEstadual sql.NullString
	Cep               sql.NullString
	Endereco          sql.NullString
	Complemento       sql.NullString
	Bairro            sql.NullString
	Cidade            sql.NullString
	Uf                sql.NullString
	DataCadastro      sql.NullTime
	DataCancelamento  sql.NullTime
	Informativo       sql.NullString
	Ativo             sql.NullString
}
