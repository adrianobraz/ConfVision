package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// OpsArmeAgendaGET retorna janelas de arme/desarme na faixa temporal (desarme=inicio, arme=fim).
func OpsArmeAgendaGET(ctx context.Context, diaSemana int) (map[string]any, error) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return nil, fmt.Errorf("timezone America/Sao_Paulo: %w", err)
	}

	now := time.Now().In(loc)
	minStr := now.Add(-1140 * time.Second).Format("15:04")
	maxStr := now.Add(1200 * time.Second).Format("15:04")
	crossMidnight := minStr > maxStr
	diaStr := fmt.Sprintf("%d", diaSemana)

	db, err := DB()
	if err != nil {
		return nil, err
	}

	desarme, err := opsArmeQueryJanela(ctx, db, "hora_inicio", minStr, maxStr, crossMidnight, diaStr)
	if err != nil {
		return nil, err
	}
	arme, err := opsArmeQueryJanela(ctx, db, "hora_fim", minStr, maxStr, crossMidnight, diaStr)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"dados": map[string]any{
			"desarme": desarme,
			"arme":    arme,
		},
	}, nil
}

func opsArmeQueryJanela(ctx context.Context, db *sql.DB, col, minStr, maxStr string, crossMidnight bool, diaStr string) ([]map[string]any, error) {
	if col != "hora_inicio" && col != "hora_fim" {
		return nil, fmt.Errorf("coluna invalida")
	}

	var query string
	if crossMidnight {
		query = fmt.Sprintf(`
SELECT id, created_at, whatsappeventocad_id, dias, hora_inicio, hora_fim,
       id_cliente, id_franqueado, whatsapp, nome, id_dispositivo, nome_dispositivo
FROM ops_arme_janela
WHERE ($1 = ANY(dias)) AND (%s >= $2 OR %s <= $3)
ORDER BY id ASC`, col, col)
	} else {
		query = fmt.Sprintf(`
SELECT id, created_at, whatsappeventocad_id, dias, hora_inicio, hora_fim,
       id_cliente, id_franqueado, whatsapp, nome, id_dispositivo, nome_dispositivo
FROM ops_arme_janela
WHERE ($1 = ANY(dias)) AND (%s >= $2 AND %s <= $3)
ORDER BY id ASC`, col, col)
	}

	rows, err := db.QueryContext(ctx, query, diaStr, minStr, maxStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		item, err := scanOpsArmeJanela(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanOpsArmeJanela(rows *sql.Rows) (map[string]any, error) {
	var id int
	var createdAt time.Time
	var whatsAppEventoCadID sql.NullInt64
	var dias []string
	var horaInicio, horaFim sql.NullString
	var idCliente, idFranqueado, whatsapp, nome, idDispositivo, nomeDispositivo sql.NullString

	if err := rows.Scan(
		&id, &createdAt, &whatsAppEventoCadID, &dias, &horaInicio, &horaFim,
		&idCliente, &idFranqueado, &whatsapp, &nome, &idDispositivo, &nomeDispositivo,
	); err != nil {
		return nil, err
	}

	item := map[string]any{
		"id":                  id,
		"created_at":          createdAt.UTC().Format(time.RFC3339),
		"whatsappeventocad_id": nullInt(whatsAppEventoCadID),
		"dias":                dias,
		"hora_inicio":         nullStr(horaInicio),
		"hora_fim":            nullStr(horaFim),
		"IdCliente":           nullStr(idCliente),
		"IdFranqueado":        nullStr(idFranqueado),
		"whatsapp":            nullStr(whatsapp),
		"nome":                nullStr(nome),
		"idDispositivo":       nullStr(idDispositivo),
		"NomeDispositivo":     nullStr(nomeDispositivo),
	}
	return item, nil
}
