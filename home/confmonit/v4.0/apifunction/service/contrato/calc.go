package contrato

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const maxDescontoPercentual = 90.0

func ValidarDesconto(valorBase float64, descontoTipo string, descontoValor float64) error {
	if descontoValor <= 0 {
		return nil
	}
	descontoTipo = strings.ToLower(strings.TrimSpace(descontoTipo))
	if descontoTipo == "" {
		return nil
	}
	if valorBase <= 0 {
		return errors.New("subtotal invalido para aplicar desconto")
	}
	maxValor := math.Round(valorBase*maxDescontoPercentual/100*100) / 100
	switch descontoTipo {
	case "percentual":
		if descontoValor > maxDescontoPercentual {
			return fmt.Errorf("desconto percentual nao pode ser maior que %.0f%%", maxDescontoPercentual)
		}
	case "valor":
		if descontoValor > maxValor {
			return fmt.Errorf("desconto em valor nao pode ser maior que 90%% do subtotal (max %.2f)", maxValor)
		}
	}
	return nil
}

func CalcularPreview(itens []ItemInput, descontoTipo string, descontoValor float64) PreviewResult {
	var calculados []ItemCalculado
	var valorBase float64
	var planoFP string
	var modulos map[string]bool
	var limites map[string]int
	retencao := 30
	qtdCotaTotal := 0
	camerasTotal := 0

	for _, it := range itens {
		if it.Quantidade <= 0 {
			continue
		}
		vt := it.ValorUnitario * float64(it.Quantidade)
		if it.Tipo == "cota" && it.QtdPorPacote > 0 {
			qtdCotaTotal += it.QtdPorPacote * it.Quantidade
		}
		if it.Tipo == "cv_licenca" {
			camerasTotal += it.Quantidade
		}
		if it.Tipo == "produto" {
			parts := strings.SplitN(it.Chave, "|", 2)
			if len(parts) == 2 && parts[0] == "franqueadopro" {
				planoFP = parts[1]
				modulos = modulosPorPlanoFP(planoFP)
				limites = limitesPorPlanoFP(planoFP)
				retencao = retencaoPorPlanoFP(planoFP)
			}
		}
		valorBase += vt
		calculados = append(calculados, ItemCalculado{
			Tipo: it.Tipo, Chave: it.Chave, Descricao: it.Descricao,
			Quantidade: it.Quantidade, QtdPorPacote: it.QtdPorPacote,
			ValorUnitario: it.ValorUnitario, ValorTotal: vt, RefID: it.RefID,
		})
	}

	if qtdCotaTotal > 0 {
		limCota := limitesDeCotaQtd(qtdCotaTotal)
		if limites == nil {
			limites = limCota
		} else {
			limites = mergeLimites(limites, limCota)
		}
	}
	if camerasTotal > 0 {
		if limites == nil {
			limites = map[string]int{}
		}
		if camerasTotal > limites["cameras_max"] {
			limites["cameras_max"] = camerasTotal
		}
	}
	if modulos == nil {
		modulos = map[string]bool{}
	}
	if limites == nil {
		limites = map[string]int{}
	}

	valorFinal := valorBase
	descontoTipo = strings.ToLower(strings.TrimSpace(descontoTipo))
	switch descontoTipo {
	case "percentual":
		if descontoValor > 0 {
			valorFinal = valorBase * (1 - descontoValor/100)
		}
	case "valor":
		valorFinal = valorBase - descontoValor
	}
	if valorFinal < 0 {
		valorFinal = 0
	}
	valorFinal = math.Round(valorFinal*100) / 100
	valorBase = math.Round(valorBase*100) / 100

	return PreviewResult{
		ValorBase: valorBase, ValorFinal: valorFinal,
		Limites: limites, Modulos: modulos, RetencaoDias: retencao,
		Itens: calculados, PlanoFP: planoFP,
	}
}

func diasPeriodicidade(periodicidade string) int {
	switch strings.ToLower(strings.TrimSpace(periodicidade)) {
	case "quinzenal":
		return 15
	case "bimestral":
		return 60
	case "trimestral":
		return 90
	case "semestral":
		return 180
	case "anual":
		return 360
	default:
		return 30
	}
}

// ValorPorPeriodicidade aplica fator dias/30 sobre o valor mensal base.
func ValorPorPeriodicidade(valorMensal float64, periodicidade string) float64 {
	dias := diasPeriodicidade(periodicidade)
	return math.Round(valorMensal*(float64(dias)/30)*100) / 100
}

func proximaCobranca(base time.Time, periodicidade string) time.Time {
	switch strings.ToLower(periodicidade) {
	case "quinzenal":
		return base.AddDate(0, 0, 15)
	case "bimestral":
		return base.AddDate(0, 2, 0)
	case "trimestral":
		return base.AddDate(0, 3, 0)
	case "semestral":
		return base.AddDate(0, 6, 0)
	case "anual":
		return base.AddDate(1, 0, 0)
	default:
		return base.AddDate(0, 1, 0)
	}
}
