package visdata

import "context"

const (
	CanalLigacao   = "ligacao"
	TipoMovRecarga = "recarga"
)

func PodeUsarCanal(ctx context.Context, idFranqueado, idCentral, canal string) (bool, float64, float64, error) {
	_ = ctx
	_ = idFranqueado
	_ = idCentral
	_ = canal
	return true, 0, 0, nil
}

func GetCreditoResumo(ctx context.Context, idFranqueado string) (map[string]any, error) {
	saldos, err := ListCreditoSaldos(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	return map[string]any{"saldos": saldos, "id_franqueado": idFranqueado}, nil
}

func CreditarSaldo(ctx context.Context, idFranqueado, servico, tipo string, valor float64, idFatura *int64, obs, criadoPor string) error {
	_ = ctx
	_ = idFranqueado
	_ = servico
	_ = tipo
	_ = valor
	_ = idFatura
	_ = obs
	_ = criadoPor
	return nil
}

func InicializarSaldoZero(ctx context.Context, idFranqueado string) error {
	_ = ctx
	_ = idFranqueado
	return nil
}

func DebitarSaldo(ctx context.Context, idFranqueado, servico string, valor float64, idProcesso, obs string) (bool, error) {
	_ = ctx
	_ = idFranqueado
	_ = servico
	_ = valor
	_ = idProcesso
	_ = obs
	return false, nil
}

type TarifaOperacional struct {
	IDFranqueado string  `json:"id_franqueado"`
	Canal        string  `json:"canal"`
	ValorMinimo  float64 `json:"valor_minimo"`
}

func GetTarifaOperacional(ctx context.Context, idCentral, canal string) (TarifaOperacional, error) {
	_ = ctx
	return TarifaOperacional{IDFranqueado: idCentral, Canal: canal, ValorMinimo: 0}, nil
}

func SaveTarifaOperacional(ctx context.Context, t TarifaOperacional) error {
	_ = ctx
	_ = t
	return nil
}
