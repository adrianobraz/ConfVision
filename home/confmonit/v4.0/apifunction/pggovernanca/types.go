package pggovernanca

const (
	MotivoPublico     = "sistema_indisponivel"
	MensagemPublica   = "Sistema temporariamente fora"
	NivelOK           = "ok"
	NivelAlerta       = "alerta"
	NivelBloqueado    = "bloqueado_carteira"
	EntidadeREP       = "REP"
	EntidadeCEN       = "CEN"
	TipoRepCentral    = "repasse_rep_central"
	TipoRepBreakglass = "repasse_central_breakglass"
)

type RepasseRow struct {
	FaturaID        int
	Tipo            string
	IDRepresentante string
	IDCentral       string
	Status          string
	Vencimento      int64 // unix ms ou sec
	ValorTotal      float64
}

type RestricaoRow struct {
	EntidadeTipo   string
	EntidadeID     string
	FaturaID       int
	Nivel          string
	Vencimento     int64
	DiasRestantes  int
	Ativo          bool
}

type EstadoEntidade struct {
	PodeOperar      bool   `json:"pode_operar"`
	Nivel           string `json:"nivel"`
	DiasRestantes   int    `json:"dias_restantes"`
	DiasVencido     int    `json:"dias_vencido"`
	MensagemPublica string `json:"mensagem_publica"`
	RepasseFaturaID int    `json:"repasse_fatura_id"`
	ValorRepasse    float64 `json:"valor_repasse"`
	TipoRepasse     string `json:"tipo_repasse"`
}

type FranqueadoEstado struct {
	BloqueadoCascata bool   `json:"bloqueado_cascata"`
	Motivo           string `json:"motivo"`
	MensagemPublica  string `json:"mensagem_publica"`
}

type TickResult struct {
	RepasseSincronizados int `json:"repasse_sincronizados"`
	RestricoesAtivas     int `json:"restricoes_ativas"`
	SuspensoesNovas      int `json:"suspensoes_novas"`
	Reativacoes          int `json:"reativacoes"`
}
