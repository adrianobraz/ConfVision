package pgcobranca

import "time"

const (
	ModoOriginal    = "original"
	ModoConsolidado = "consolidado"
	ModoAlinhar     = "alinhar"

	PeriodicidadeMensal    = "mensal"
	PeriodicidadeQuinzenal = "quinzenal"

	StatusCicloNormal      = "normal"
	StatusCicloEmAjuste    = "em_ajuste"
	StatusCicloConsolidado = "consolidado"
)

type Config struct {
	IDFranqueado    string `json:"id_franqueado"`
	Modo            string `json:"modo"`
	DiaMensal       int    `json:"dia_mensal"`
	DiaQuinzenal1   int    `json:"dia_quinzenal_1"`
	DiaQuinzenal2   int    `json:"dia_quinzenal_2"`
	IDCentral       string `json:"id_central"`
	IDRepresentante string `json:"id_representante"`
}

type Servico struct {
	ID              int64     `json:"id"`
	IDFranqueado    string    `json:"id_franqueado"`
	RefTipo         string    `json:"ref_tipo"`
	RefID           string    `json:"ref_id"`
	Descricao       string    `json:"descricao"`
	DataInicio      time.Time `json:"data_inicio"`
	Periodicidade   string    `json:"periodicidade"`
	ValorCiclo      float64   `json:"valor_ciclo"`
	DiasCiclo       int       `json:"dias_ciclo"`
	PagoAte         *time.Time `json:"pago_ate,omitempty"`
	SaldoAjuste     float64   `json:"saldo_ajuste"`
	StatusCiclo     string    `json:"status_ciclo"`
	ValorPiso       float64   `json:"valor_piso"`
	MargemCentral   float64   `json:"margem_central"`
	MargemRep       float64   `json:"margem_rep"`
	Ativo           bool      `json:"ativo"`
}

type LinhaOrdem struct {
	ServicoID     int64   `json:"servico_id"`
	RefTipo       string  `json:"ref_tipo"`
	RefID         string  `json:"ref_id"`
	Descricao     string  `json:"descricao"`
	PeriodoInicio time.Time `json:"periodo_inicio"`
	PeriodoFim    time.Time `json:"periodo_fim"`
	ValorCheio    float64 `json:"valor_cheio"`
	Credito       float64 `json:"credito"`
	Ajuste        float64 `json:"ajuste"`
	ValorLiquido  float64 `json:"valor_liquido"`
	ValorPiso     float64 `json:"valor_piso"`
	MargemCentral float64 `json:"margem_central"`
	MargemRep     float64 `json:"margem_rep"`
}

type OrdemSimulada struct {
	IDFranqueado  string       `json:"id_franqueado"`
	CicloRef      string       `json:"ciclo_ref"`
	Vencimento    time.Time    `json:"vencimento"`
	Linhas        []LinhaOrdem `json:"linhas"`
	SubtotalCheio float64      `json:"subtotal_cheio"`
	TotalCreditos float64      `json:"total_creditos"`
	TotalAjustes  float64      `json:"total_ajustes"`
	ValorTotal    float64      `json:"valor_total"`
}

type GerarOrdensResult struct {
	Vencimento   string `json:"vencimento"`
	Processados  int    `json:"processados"`
	Ordens       int    `json:"ordens"`
	FaturasXano  []int  `json:"faturas_xano"`
	Erros        []string `json:"erros,omitempty"`
}
