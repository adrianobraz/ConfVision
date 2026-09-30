package pgcatalogo

import "encoding/json"

type Produto struct {
	ID                   int             `json:"id"`
	Produto              string          `json:"produto"`
	Plano                string          `json:"plano"`
	NomeExibicao         string          `json:"nome_exibicao"`
	ValorMensal          float64         `json:"valor_mensal"`
	ValorPisoBreakglass  float64         `json:"valor_piso_breakglass"`
	ValorPisoMinimo      float64         `json:"valor_piso_minimo"`
	PeriodicidadePadrao  string          `json:"periodicidade_padrao,omitempty"`
	RetencaoDias         int             `json:"retencao_dias"`
	LimitesJSON          json.RawMessage `json:"limites_json,omitempty"`
	ModulosJSON          json.RawMessage `json:"modulos_json,omitempty"`
	FPPacoteCotaID       *int            `json:"fp_pacote_cota_id"`
	PacoteCotaNome       string          `json:"pacote_cota_nome,omitempty"`
	PacoteCotaQuantidade int             `json:"pacote_cota_quantidade,omitempty"`
	ValorCota            float64         `json:"valor_cota,omitempty"`
	ValorTotalMensal     float64         `json:"valor_total_mensal,omitempty"`
	Ativo                string          `json:"ativo"`
	Observacao           string          `json:"observacao,omitempty"`
	ModoPreco            string          `json:"modo_preco,omitempty"`
}

type PacoteCota struct {
	ID         int     `json:"id"`
	Nome       string  `json:"nome"`
	Quantidade int     `json:"quantidade"`
	Valor      float64 `json:"valor"`
	ValorPiso  float64 `json:"valor_piso"`
	ValorVenda float64 `json:"valor_venda,omitempty"`
	PrecoRepID *int    `json:"preco_rep_id,omitempty"`
	Ativo      string  `json:"ativo"`
	Observacao string  `json:"observacao,omitempty"`
}

type SalvarProdutoInput struct {
	ID              int             `json:"id"`
	Produto         string          `json:"produto"`
	Plano           string          `json:"plano"`
	NomeExibicao    string          `json:"nome_exibicao"`
	ValorMensal     float64         `json:"valor_mensal"`
	Periodicidade   string          `json:"periodicidade_padrao"`
	RetencaoDias    int             `json:"retencao_dias"`
	LimitesJSON     json.RawMessage `json:"limites_json"`
	ModulosJSON     json.RawMessage `json:"modulos_json"`
	FPPacoteCotaID  *int            `json:"fp_pacote_cota_id"`
	Ativo           string          `json:"ativo"`
	Observacao      string          `json:"observacao"`
}

type SalvarPacoteInput struct {
	ID         int     `json:"id"`
	Nome       string  `json:"nome"`
	Quantidade int     `json:"quantidade"`
	Valor      float64 `json:"valor"`
	Ativo      string  `json:"ativo"`
	Observacao string  `json:"observacao"`
}

type PrecoConfig struct {
	IDCentral               string  `json:"id_central"`
	ModoPreco               string  `json:"modo_preco"`
	ValorUnitarioMinimoCota float64 `json:"valor_unitario_minimo_cota"`
	Observacao              string  `json:"observacao,omitempty"`
	Nome                    string  `json:"nome,omitempty"`
}

type SalvarPrecoPacoteRepInput struct {
	FPPacoteCotaID int     `json:"fp_pacote_cota_id"`
	ValorVenda     float64 `json:"valor_venda"`
	IDRepresentante string `json:"id_representante"`
	Observacao     string  `json:"observacao"`
}
