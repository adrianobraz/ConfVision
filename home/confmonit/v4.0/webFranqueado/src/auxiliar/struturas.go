package auxiliar

import "net/http"

type Rota struct {
	URI    string
	Metodo string
	Funcao func(http.ResponseWriter, *http.Request)
	Aberto bool
}

type Pagina struct {
	TituloSite  string `json:"tituloSite"`
	NomeTela    string `json:"nomeTela"`
	LinkRetorno string `json:"linkRetorno"`
	LinkJs      string `json:"linkJs"`
	LogoMarca   string `json:"logoMarca"`
}

type Login struct {
	IdFranqueado   string `json:"idVinculo"`
	NomeFranqueado string `json:"nomeFranqueado"`
	IdUsuario      string `json:"idUsuario"`
	NomeUsuario    string `json:"nomeUsuario"`
	EmailUsuario   string `json:"email"`
	Senha          string `json:"senha,omitempty"`
	Token          string `json:"token"`
	TerminalAtivo  string `json:"terminalAtivo"`
	GradeAtivo     string `json:"gradeAtivo"`
}
