package autofim

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

type decision struct {
	Motivo              string
	Acao                string
	FinalizarAgora      bool
	TemFalhas           bool
	TemAlarme           bool
	TemDesarme          bool
	TemArme             bool
	TemRestaure         bool
	TemParAlarmeRest50  bool
	QtdCiclos3mDiffProc int
	PerfilLocal         string
	QtdMonitorados      int
	QtdRestaure         int
}

func evaluateBotFinalizaEventoAuto(events []contextEvent, hist histAgg, processoJaFinalizado bool) decision {
	d := decision{
		Motivo: "SEM_REGRA",
		Acao:   "NAO_FINALIZOU",
	}
	if hist.QtdCiclos3mDiffProc > 0 {
		d.QtdCiclos3mDiffProc = hist.QtdCiclos3mDiffProc
	} else {
		d.QtdCiclos3mDiffProc = hist.QtdCiclos3mDiffProcCamel
	}
	if strings.TrimSpace(hist.PerfilLocal) != "" {
		d.PerfilLocal = strings.ToUpper(strings.TrimSpace(hist.PerfilLocal))
	} else if strings.TrimSpace(hist.PerfilLocalCamel) != "" {
		d.PerfilLocal = strings.ToUpper(strings.TrimSpace(hist.PerfilLocalCamel))
	} else {
		d.PerfilLocal = "NORMAL"
	}

	if processoJaFinalizado {
		d.Motivo = "PROCESSO_JA_FINALIZADO_NO_INICIO"
		d.Acao = "JA_FINALIZADO"
		return d
	}

	sort.Slice(events, func(i, j int) bool {
		return normalizeTs(events[i].CreatedAt) < normalizeTs(events[j].CreatedAt)
	})

	var pendingAlarmeTs int64
	var pendingAlarmePart, pendingAlarmeZona string
	var pendingRestaureTs int64
	var pendingRestaurePart, pendingRestaureZona string

	for _, e := range events {
		grupo := normalizeGrupo(e.CtiGrupo, e.CtiDescricao)
		ts := normalizeTs(e.CreatedAt)

		if grupo == "ALARME" || grupo == "ARME" || grupo == "DESARME" || grupo == "FALHAS" || grupo == "RESTAURE" {
			d.QtdMonitorados++
		}

		switch grupo {
		case "FALHAS":
			d.TemFalhas = true
		case "DESARME":
			d.TemDesarme = true
		case "ARME":
			d.TemArme = true
		case "ALARME":
			d.TemAlarme = true
			if pendingRestaureTs > 0 && sameSector(e.Particao, e.ZonaUser, pendingRestaurePart, pendingRestaureZona) {
				diff := ts - pendingRestaureTs
				if diff >= 0 && diff <= 50 {
					d.TemParAlarmeRest50 = true
					pendingRestaureTs = 0
					pendingRestaurePart = ""
					pendingRestaureZona = ""
				}
			}
			pendingAlarmeTs = ts
			pendingAlarmePart = e.Particao
			pendingAlarmeZona = e.ZonaUser
		case "RESTAURE":
			d.TemRestaure = true
			d.QtdRestaure++
			if pendingAlarmeTs > 0 && sameSector(e.Particao, e.ZonaUser, pendingAlarmePart, pendingAlarmeZona) {
				diff := ts - pendingAlarmeTs
				if diff >= 0 && diff <= 50 {
					d.TemParAlarmeRest50 = true
					pendingAlarmeTs = 0
					pendingAlarmePart = ""
					pendingAlarmeZona = ""
				}
			}
			pendingRestaureTs = ts
			pendingRestaurePart = e.Particao
			pendingRestaureZona = e.ZonaUser
		}
	}

	// Regra ajustada: ARME ou DESARME finaliza.
	if d.TemDesarme || d.TemArme {
		d.Motivo = "FINALIZA_ARME_OU_DESARME"
		d.FinalizarAgora = true
		return d
	}

	if d.TemAlarme && d.TemParAlarmeRest50 {
		if d.PerfilLocal == "NORMAL" && d.QtdCiclos3mDiffProc > 1 {
			d.Motivo = "NAO_FINALIZA_NORMAL_REPETICAO_DIFF_PROCESSOS_3M"
			d.Acao = "NAO_FINALIZOU"
			return d
		}
		d.Motivo = "FINALIZA_ALARME_RESTAURE_50S"
		d.FinalizarAgora = true
		return d
	}

	if d.TemAlarme && !d.TemParAlarmeRest50 && d.TemRestaure {
		d.Motivo = "NAO_FINALIZA_ALARME_RESTAURE_ACIMA_50S"
		return d
	}

	if !d.TemAlarme && d.TemFalhas {
		d.Motivo = "FINALIZA_FALHAS"
		d.FinalizarAgora = true
		return d
	}

	if !d.TemAlarme && d.QtdMonitorados > 0 && d.QtdMonitorados == d.QtdRestaure {
		d.Motivo = "FINALIZA_RESTAURE_ISOLADO"
		d.FinalizarAgora = true
		return d
	}

	return d
}

func normalizeGrupo(ctiGrupo, ctiDescricao string) string {
	g := strings.ToUpper(strings.TrimSpace(ctiGrupo))
	if strings.Contains(strings.ToUpper(strings.TrimSpace(ctiDescricao)), "RESTAUR") {
		return "RESTAURE"
	}
	return g
}

func normalizeTs(v any) int64 {
	ts := toInt64(v)
	if ts > 9999999999 {
		return ts / 1000
	}
	return ts
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return n
	default:
		return 0
	}
}

func sameSector(partA, zonaA, partB, zonaB string) bool {
	if strings.TrimSpace(partA) != strings.TrimSpace(partB) {
		return false
	}
	za := strings.TrimSpace(zonaA)
	zb := strings.TrimSpace(zonaB)
	if za == zb {
		return true
	}
	if za == "" || zb == "" {
		return false
	}
	a, errA := strconv.Atoi(za)
	b, errB := strconv.Atoi(zb)
	if errA != nil || errB != nil {
		return false
	}
	return int(math.Abs(float64(a-b))) == 10
}
