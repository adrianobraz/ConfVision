package monitor

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/xano"
)

const keepAliveLimiteMinutos = 10

type statusSetorResp struct {
	IdSetor string `json:"idSetor"`
	Status  string `json:"status"`
}

func calcularStatusMapa(mapaAmbienteId int) ([]statusSetorResp, error) {
	setores, err := xano.ListarSetoresPorMapa(mapaAmbienteId)
	if err != nil {
		return nil, err
	}
	if len(setores) == 0 {
		return []statusSetorResp{}, nil
	}

	idSetores := make([]string, 0, len(setores))
	dispSet := make(map[string]struct{})
	for _, s := range setores {
		if s.IdSetor != "" {
			idSetores = append(idSetores, s.IdSetor)
		}
		if s.IdDispositivo != "" {
			dispSet[s.IdDispositivo] = struct{}{}
		}
	}

	idDispositivos := make([]string, 0, len(dispSet))
	for id := range dispSet {
		idDispositivos = append(idDispositivos, id)
	}

	alarme, err := buscarSetoresEmAlarme(idSetores)
	if err != nil {
		return nil, err
	}

	offline, err := buscarDispositivosKeepAliveOffline(idDispositivos)
	if err != nil {
		return nil, err
	}

	out := make([]statusSetorResp, 0, len(setores))
	for _, s := range setores {
		st := "normal"
		if alarme[s.IdSetor] {
			st = "alarme"
		} else if offline[s.IdDispositivo] {
			st = "falha"
		}
		out = append(out, statusSetorResp{
			IdSetor: s.IdSetor,
			Status:  st,
		})
	}
	return out, nil
}

// StatusMapasBatch calcula o pior status de cada mapa (alarme > falha > normal)
// em uma unica varredura, buscando setores/dispositivos em lote para evitar N consultas.
func StatusMapasBatch(ids []int) (map[int]string, error) {
	out := make(map[int]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	pedido := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		pedido[id] = struct{}{}
		out[id] = "normal"
	}

	todosSetores, err := xano.ListarTodosSetores()
	if err != nil {
		return nil, err
	}

	setoresPorMapa := make(map[int][]string)
	dispPorMapa := make(map[int][]string)
	allSetores := make(map[string]struct{})
	allDisp := make(map[string]struct{})

	for _, s := range todosSetores {
		if _, ok := pedido[s.MapaAmbienteId]; !ok {
			continue
		}
		if s.IdSetor != "" {
			setoresPorMapa[s.MapaAmbienteId] = append(setoresPorMapa[s.MapaAmbienteId], s.IdSetor)
			allSetores[s.IdSetor] = struct{}{}
		}
		if s.IdDispositivo != "" {
			dispPorMapa[s.MapaAmbienteId] = append(dispPorMapa[s.MapaAmbienteId], s.IdDispositivo)
			allDisp[s.IdDispositivo] = struct{}{}
		}
	}

	listaSetores := make([]string, 0, len(allSetores))
	for id := range allSetores {
		listaSetores = append(listaSetores, id)
	}
	listaDisp := make([]string, 0, len(allDisp))
	for id := range allDisp {
		listaDisp = append(listaDisp, id)
	}

	alarme, err := buscarSetoresEmAlarme(listaSetores)
	if err != nil {
		return nil, err
	}
	offline, err := buscarDispositivosKeepAliveOffline(listaDisp)
	if err != nil {
		return nil, err
	}

	for id := range pedido {
		st := "normal"
		for _, s := range setoresPorMapa[id] {
			if alarme[s] {
				st = "alarme"
				break
			}
		}
		if st == "normal" {
			for _, d := range dispPorMapa[id] {
				if offline[d] {
					st = "falha"
					break
				}
			}
		}
		out[id] = st
	}
	return out, nil
}

func buscarSetoresEmAlarme(idSetores []string) (map[string]bool, error) {
	out := make(map[string]bool)
	if len(idSetores) == 0 {
		return out, nil
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ph := strings.Repeat("?,", len(idSetores))
	ph = strings.TrimSuffix(ph, ",")
	args := make([]interface{}, len(idSetores))
	for i, id := range idSetores {
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT setorAlarme.ID_Setor
		FROM setorAlarme
		INNER JOIN processo
			ON processo.ID_Dispositivo = setorAlarme.ID_Dispositivo
			AND processo.DataAtenFim IS NULL
			AND processo.Nivel > 0
		INNER JOIN evento
			ON evento.ID_Processo = processo.ID_Processo
			AND evento.ZonaUser = setorAlarme.Numero
		WHERE setorAlarme.ID_Setor IN (%s)
	`, ph)

	tab, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	for tab.Next() {
		var id sql.NullString
		if err := tab.Scan(&id); err != nil {
			return nil, err
		}
		if id.String != "" {
			out[id.String] = true
		}
	}
	return out, nil
}

func buscarDispositivosKeepAliveOffline(idDispositivos []string) (map[string]bool, error) {
	out := make(map[string]bool)
	if len(idDispositivos) == 0 {
		return out, nil
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ph := strings.Repeat("?,", len(idDispositivos))
	ph = strings.TrimSuffix(ph, ",")
	args := make([]interface{}, len(idDispositivos))
	for i, id := range idDispositivos {
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT dispositivo.ID_Dispositivo, dispositivo.KeepAlive, dispositivo.DataUltimoEvento
		FROM dispositivo
		WHERE dispositivo.ID_Dispositivo IN (%s)
	`, ph)

	tab, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	limite := time.Duration(keepAliveLimiteMinutos) * time.Minute
	agora := time.Now()

	for tab.Next() {
		var idDisp, keepAlive sql.NullString
		var dataUltimo sql.NullTime
		if err := tab.Scan(&idDisp, &keepAlive, &dataUltimo); err != nil {
			return nil, err
		}

		keepMin, _ := strconv.Atoi(strings.TrimSpace(keepAlive.String))
		if keepMin <= 0 {
			continue
		}

		if !dataUltimo.Valid {
			out[idDisp.String] = true
			continue
		}

		if agora.Sub(dataUltimo.Time) > limite {
			out[idDisp.String] = true
		}
	}
	return out, nil
}
