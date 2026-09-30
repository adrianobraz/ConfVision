package credito

import (
	"context"
	"apifunction/ops"
	"apifunction/pgcredito"
)

func creditarWallet(idFranqueado, servico, tipo string, valor float64, idFatura int64, obs, criadoPor string) error {
	if pgcredito.Configurado() {
		var idFat *int64
		if idFatura > 0 {
			idFat = &idFatura
		}
		return pgcredito.CreditarSaldo(context.Background(), idFranqueado, servico, tipo, valor, idFat, obs, criadoPor)
	}
	return ops.Creditar(idFranqueado, servico, tipo, valor, idFatura, obs, criadoPor)
}

func debitarWallet(idFranqueado, servico string, valor float64, idProcesso, obs string) (jaDebitado bool, err error) {
	if pgcredito.Configurado() {
		res, err := pgcredito.DebitarSaldo(context.Background(), idFranqueado, servico, valor, idProcesso, obs)
		return res.JaDebitado, err
	}
	res, err := ops.Debitar(idFranqueado, servico, valor, idProcesso, obs)
	if err != nil {
		return false, err
	}
	if res != nil {
		return res.JaDebitado, nil
	}
	return false, nil
}
