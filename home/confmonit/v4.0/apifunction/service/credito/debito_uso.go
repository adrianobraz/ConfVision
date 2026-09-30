package credito

import (
	"apifunction/service/tarifa"
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	TipoCobrancaMinuto    = "minuto"
	TipoCobrancaUnidade   = "unidade"
	TipoCobrancaTentativa = "tentativa"
)

type DebitarUsoInput struct {
	IDFranqueado    string  `json:"id_franqueado"`
	Servico         string  `json:"servico"`
	TipoCobranca    string  `json:"tipo_cobranca"`
	DuracaoSegundos float64 `json:"duracao_segundos"`
	IDProcesso      string  `json:"id_processo"`
	Observacao      string  `json:"observacao"`
}

type DebitarUsoResult struct {
	OK       bool    `json:"ok"`
	Debitado bool    `json:"debitado"`
	Valor    float64 `json:"valor"`
	Motivo   string  `json:"motivo,omitempty"`
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func CalcValorUso(t tarifa.TarifaRep, tipo string, duracaoSeg float64) float64 {
	switch strings.TrimSpace(strings.ToLower(tipo)) {
	case TipoCobrancaMinuto:
		if duracaoSeg <= 0 {
			return 0
		}
		return round2((t.ValorMinuto / 60.0) * duracaoSeg)
	case TipoCobrancaUnidade:
		return round2(t.ValorUnidade)
	case TipoCobrancaTentativa:
		return round2(t.ValorTentativa)
	default:
		return 0
	}
}

func DebitarUso(in DebitarUsoInput) (DebitarUsoResult, error) {
	out := DebitarUsoResult{OK: true}
	idFra := strings.TrimSpace(in.IDFranqueado)
	servico := strings.TrimSpace(strings.ToLower(in.Servico))
	tipo := strings.TrimSpace(strings.ToLower(in.TipoCobranca))
	idProc := strings.TrimSpace(in.IDProcesso)
	if idFra == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	if servico == "" {
		return out, errors.New("servico obrigatorio")
	}
	if tipo == "" {
		return out, errors.New("tipo_cobranca obrigatorio")
	}
	if idProc == "" {
		return out, errors.New("id_processo obrigatorio")
	}

	t, err := tarifa.GetTarifaEfetiva(idFra, servico)
	if err != nil {
		return out, err
	}
	valor := CalcValorUso(t, tipo, in.DuracaoSegundos)
	out.Valor = valor
	if valor <= 0 {
		out.Debitado = false
		out.Motivo = "tarifa_zero"
		return out, nil
	}

	obs := strings.TrimSpace(in.Observacao)
	if obs == "" {
		obs = fmt.Sprintf("%s %s", servico, tipo)
	}
	res, err := debitarWallet(idFra, servico, valor, idProc, obs)
	if err != nil {
		return out, err
	}
	if res {
		out.Debitado = false
		out.Motivo = "ja_debitado"
		return out, nil
	}
	out.Debitado = true
	return out, nil
}
