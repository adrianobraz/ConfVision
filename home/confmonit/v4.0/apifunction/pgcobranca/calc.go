package pgcobranca

import (
	"fmt"
	"math"
	"time"
)

var locSP *time.Location

func init() {
	locSP, _ = time.LoadLocation("America/Sao_Paulo")
	if locSP == nil {
		locSP = time.FixedZone("BRT", -3*3600)
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.In(locSP).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, locSP)
}

func daysInclusive(a, b time.Time) int {
	a = dateOnly(a)
	b = dateOnly(b)
	if b.Before(a) {
		return 0
	}
	return int(b.Sub(a).Hours()/24) + 1
}

func diaria(valorCiclo float64, diasCiclo int) float64 {
	if diasCiclo <= 0 {
		diasCiclo = 15
	}
	return valorCiclo / float64(diasCiclo)
}

// PeriodoForAnchor retorna inicio/fim do periodo cobrado na data ancora (14 ou 29).
func PeriodoForAnchor(anchor time.Time, periodicidade string) (inicio, fim time.Time) {
	anchor = dateOnly(anchor)
	y, m, d := anchor.Date()
	day := d

	switch periodicidade {
	case PeriodicidadeMensal:
		// Mensal no dia 29: periodo ~29 deste mes ate dia antes do 29 do mes seguinte
		inicio = anchor
		ny, nm := y, m+1
		if nm > 12 {
			nm = 1
			ny++
		}
		nextAnchor := clampDay(ny, nm, 29)
		fim = nextAnchor.AddDate(0, 0, -1)
		return inicio, fim

	default: // quinzenal
		if day <= 14 {
			inicio = time.Date(y, m, 1, 0, 0, 0, 0, locSP)
			if day == 14 {
				inicio = anchor
			}
			fim = clampDay(y, m, 29).AddDate(0, 0, -1)
			if day == 14 {
				fim = clampDay(y, m, 29)
				if fim.Month() != m {
					fim = time.Date(y, m+1, 0, 0, 0, 0, 0, locSP) // ultimo dia do mes
				}
			}
		} else {
			// ancora 29: periodo 29 -> ~12/13 do mes seguinte (~15 dias)
			inicio = anchor
			ny, nm := y, m+1
			if nm > 12 {
				nm = 1
				ny++
			}
			fim = time.Date(ny, nm, 13, 0, 0, 0, 0, locSP)
		}
		return inicio, fim
	}
}

func clampDay(y int, m time.Month, day int) time.Time {
	last := time.Date(y, m+1, 0, 0, 0, 0, 0, locSP).Day()
	if day > last {
		day = last
	}
	return time.Date(y, m, day, 0, 0, 0, 0, locSP)
}

func anchorMatches(cfg Config, anchor time.Time) bool {
	anchor = dateOnly(anchor)
	d := anchor.Day()
	if cfg.Modo == ModoOriginal {
		return false
	}
	if d == cfg.DiaQuinzenal1 || d == cfg.DiaQuinzenal2 {
		return true
	}
	return false
}

func servicoCobraNoAnchor(s Servico, cfg Config, anchor time.Time) bool {
	anchor = dateOnly(anchor)
	d := anchor.Day()
	switch s.Periodicidade {
	case PeriodicidadeMensal:
		return d == cfg.DiaMensal
	default:
		return d == cfg.DiaQuinzenal1 || d == cfg.DiaQuinzenal2
	}
}

func CalcLinha(s Servico, cfg Config, anchor time.Time) LinhaOrdem {
	inicio, fim := PeriodoForAnchor(anchor, s.Periodicidade)
	valorCheio := s.ValorCiclo
	diasPeriodo := daysInclusive(inicio, fim)
	diasRef := s.DiasCiclo
	if diasRef <= 0 {
		diasRef = 15
	}
	dia := diaria(s.ValorCiclo, diasRef)

	// Proporcional se periodo parcial vs ciclo padrao
	if diasPeriodo > 0 && diasPeriodo < diasRef {
		valorCheio = round2(dia * float64(diasPeriodo))
	}

	credito := 0.0
	if s.PagoAte != nil {
		pagoAte := dateOnly(*s.PagoAte)
		if !pagoAte.Before(inicio) {
			overlapEnd := pagoAte
			if overlapEnd.After(fim) {
				overlapEnd = fim
			}
			diasPagos := daysInclusive(inicio, overlapEnd)
			credito = round2(dia * float64(diasPagos))
		}
	}
	credito += s.SaldoAjuste
	if credito < 0 {
		credito = 0
	}
	if credito > valorCheio {
		credito = valorCheio
	}

	liquido := round2(valorCheio - credito)
	ajuste := round2(credito)

	desc := fmt.Sprintf("%s (%s a %s)", s.Descricao,
		inicio.Format("02/01/2006"), fim.Format("02/01/2006"))
	if ajuste > 0 {
		desc += fmt.Sprintf(" | credito -R$ %.2f", ajuste)
	}

	return LinhaOrdem{
		ServicoID:     s.ID,
		RefTipo:       s.RefTipo,
		RefID:         s.RefID,
		Descricao:     desc,
		PeriodoInicio: inicio,
		PeriodoFim:    fim,
		ValorCheio:    valorCheio,
		Credito:       credito,
		Ajuste:        -ajuste,
		ValorLiquido:  liquido,
		ValorPiso:     s.ValorPiso,
		MargemCentral: s.MargemCentral,
		MargemRep:     s.MargemRep,
	}
}

func MontarOrdem(idFranqueado string, cfg Config, anchor time.Time, servicos []Servico) *OrdemSimulada {
	anchor = dateOnly(anchor)
	var linhas []LinhaOrdem
	for _, s := range servicos {
		if !s.Ativo {
			continue
		}
		if !servicoCobraNoAnchor(s, cfg, anchor) {
			continue
		}
		linhas = append(linhas, CalcLinha(s, cfg, anchor))
	}
	if len(linhas) == 0 {
		return nil
	}
	ord := &OrdemSimulada{
		IDFranqueado: idFranqueado,
		CicloRef:     fmt.Sprintf("ORD-%s-%s", idFranqueado, anchor.Format("20060102")),
		Vencimento:   anchor,
		Linhas:       linhas,
	}
	for _, l := range linhas {
		ord.SubtotalCheio += l.ValorCheio
		ord.TotalCreditos += l.Credito
		ord.ValorTotal += l.ValorLiquido
	}
	ord.SubtotalCheio = round2(ord.SubtotalCheio)
	ord.TotalCreditos = round2(ord.TotalCreditos)
	ord.TotalAjustes = round2(-ord.TotalCreditos)
	ord.ValorTotal = round2(ord.ValorTotal)
	return ord
}
