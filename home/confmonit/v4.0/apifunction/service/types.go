package service

const (
	TipoCLI = "CLI"
	TipoFRA = "FRA"
	TipoREP = "REP"

	MaxEventos = 100
)

type ReqTransferencia struct {
	Tipo      string `json:"tipo"`
	IDOrigem  string `json:"idOrigem"`
	IDDestino string `json:"idDestino"`
	Motivo    string `json:"motivo"`
}

type Contagens struct {
	Eventos          int `json:"eventos"`
	Faturas          int `json:"faturas"`
	ProcessosAbertos int `json:"processosAbertos"`
	TicketsAbertos   int `json:"ticketsAbertos"`
	Clientes         int `json:"clientes,omitempty"`
	Franqueados      int `json:"franqueados,omitempty"`
}

type Bloqueio struct {
	Codigo    string `json:"codigo"`
	Mensagem  string `json:"mensagem"`
	Contagem  int    `json:"contagem,omitempty"`
	Entidade  string `json:"entidade,omitempty"`
	IDEntidade string `json:"idEntidade,omitempty"`
}

type PreviewResult struct {
	OK         bool       `json:"ok"`
	Bloqueios  []Bloqueio `json:"bloqueios"`
	Contagens  Contagens  `json:"contagens"`
	VinculoAntigo string  `json:"vinculoAntigo,omitempty"`
	VinculoNovo   string  `json:"vinculoNovo,omitempty"`
	DescricaoOrigem  string `json:"descricaoOrigem,omitempty"`
	DescricaoDestino string `json:"descricaoDestino,omitempty"`
}

type ExecResult struct {
	LogID string `json:"logId"`
}
