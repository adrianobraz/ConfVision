package contrato

import (
	"strings"
	"time"
)

type ItemInput struct {
	Tipo          string  `json:"tipo"`
	Chave         string  `json:"chave"`
	Descricao     string  `json:"descricao,omitempty"`
	Quantidade    int     `json:"quantidade"`
	QtdPorPacote  int     `json:"qtd_por_pacote,omitempty"`
	ValorUnitario float64 `json:"valor_unitario"`
	RefID         int     `json:"ref_id,omitempty"`
}

type SalvarInput struct {
	IDContrato      int         `json:"id_contrato,omitempty"`
	IDFranqueado    string      `json:"id_franqueado"`
	Nome            string      `json:"nome"`
	Itens           []ItemInput `json:"itens"`
	DescontoTipo    string      `json:"desconto_tipo,omitempty"`
	DescontoValor   float64     `json:"desconto_valor,omitempty"`
	Periodicidade   string      `json:"periodicidade"`
	InicioEm        string      `json:"inicio_em"`
	ProximoReajuste string      `json:"proximo_reajuste_em,omitempty"`
	VencimentoEm    string      `json:"vencimento_em,omitempty"`
	VencimentoDia   int         `json:"vencimento_dia,omitempty"`
	PermiteExcedente string     `json:"permite_excedente,omitempty"`
	ValorUnitCamera float64     `json:"valor_unitario_camera,omitempty"`
	Observacao      string      `json:"observacao,omitempty"`
	GerarFatura     bool        `json:"gerar_fatura"`
}

type PreviewResult struct {
	ValorBase    float64            `json:"valor_base"`
	ValorFinal   float64            `json:"valor_final"`
	Limites      map[string]int     `json:"limites_json"`
	Modulos      map[string]bool    `json:"modulos_json"`
	RetencaoDias int                `json:"retencao_dias"`
	Itens        []ItemCalculado    `json:"itens"`
	PlanoFP      string             `json:"plano_franqueadopro,omitempty"`
}

type ItemCalculado struct {
	Tipo          string  `json:"tipo"`
	Chave         string  `json:"chave"`
	Descricao     string  `json:"descricao"`
	Quantidade    int     `json:"quantidade"`
	QtdPorPacote  int     `json:"qtd_por_pacote,omitempty"`
	ValorUnitario float64 `json:"valor_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	RefID         int     `json:"ref_id,omitempty"`
}

type ContratoResumo struct {
	IDContrato       int       `json:"id_contrato"`
	IDFranqueado     string    `json:"id_franqueado"`
	Nome             string    `json:"nome"`
	Status           string    `json:"status"`
	ValorFinal       float64   `json:"valor_final"`
	Periodicidade    string    `json:"periodicidade"`
	InicioEm         string    `json:"inicio_em"`
	ProximaCobranca  string    `json:"proxima_cobranca_em,omitempty"`
	Itens            []ItemCalculado `json:"itens,omitempty"`
}

type ContratoDetalhe struct {
	ContratoResumo
	DescontoTipo        string  `json:"desconto_tipo,omitempty"`
	DescontoValor       float64 `json:"desconto_valor,omitempty"`
	ProximoReajusteEm   string  `json:"proximo_reajuste_em,omitempty"`
	IDRepresentante     string  `json:"id_representante,omitempty"`
	NomeFranqueado      string  `json:"nome_franqueado,omitempty"`
	NomeRepresentante   string  `json:"nome_representante,omitempty"`
	Editavel            bool    `json:"editavel"`
}

type EfetivoResult struct {
	Liberado     bool           `json:"liberado"`
	Motivo       string         `json:"motivo"`
	Produto      string         `json:"produto"`
	Plano        string         `json:"plano_efetivo,omitempty"`
	ModulosJSON  map[string]bool `json:"modulos_json"`
	LimitesJSON  map[string]int  `json:"limites_json"`
	RetencaoDias int            `json:"retencao_dias"`
	ContratoID   int            `json:"contrato_id,omitempty"`
	Status       string         `json:"status,omitempty"`
}

type FaturaResumo struct {
	IDFatura          int              `json:"id_fatura"`
	IDContrato        int              `json:"id_contrato"`
	Tipo              string           `json:"tipo"`
	Referencia        string           `json:"referencia"`
	Status            string           `json:"status"`
	ValorTotal        float64          `json:"valor_total"`
	VencimentoEm      string           `json:"vencimento_em"`
	PagoEm            string           `json:"pago_em,omitempty"`
	IDFaturaContabil  int              `json:"id_fatura_contabil,omitempty"`
	Itens             []ItemCalculado  `json:"itens,omitempty"`
}

type CatalogoProduto struct {
	ID           int     `json:"id"`
	Produto      string  `json:"produto"`
	Plano        string  `json:"plano"`
	NomeExibicao string  `json:"nome_exibicao"`
	GrupoUI      string  `json:"grupo_ui"`
	TipoSelecao  string  `json:"tipo_selecao"`
	ValorMensal  float64 `json:"valor_mensal"`
	RetencaoDias int     `json:"retencao_dias"`
	Ordem        int     `json:"ordem"`
}

type PacoteCota struct {
	ID         int     `json:"id"`
	Nome       string  `json:"nome"`
	Quantidade int     `json:"quantidade"`
	Valor      float64 `json:"valor"`
	Ativo      string  `json:"ativo"`
}

type CVLicencaCat struct {
	ID           int     `json:"id"`
	Plano        string  `json:"plano"`
	NomeExibicao string  `json:"nome_exibicao"`
	Unidade      string  `json:"unidade"`
	ValorMensal  float64 `json:"valor_mensal"`
	Ordem        int     `json:"ordem"`
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now(), nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
