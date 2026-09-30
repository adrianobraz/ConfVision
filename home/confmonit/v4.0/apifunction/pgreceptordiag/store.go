package pgreceptordiag

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"apifunction/pgcredito"
)

func UpsertPresenca(p ReqPresenca) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
INSERT INTO fp_receptor_presenca (
  id_franqueado, fabricante, modulo, conta, id_dispositivo, ip_remoto, ultimo_evento, ultimo_sinal_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
ON CONFLICT (fabricante, conta) DO UPDATE SET
  id_franqueado = EXCLUDED.id_franqueado,
  modulo = EXCLUDED.modulo,
  id_dispositivo = EXCLUDED.id_dispositivo,
  ip_remoto = EXCLUDED.ip_remoto,
  ultimo_evento = CASE WHEN EXCLUDED.ultimo_evento <> '' THEN EXCLUDED.ultimo_evento ELSE fp_receptor_presenca.ultimo_evento END,
  ultimo_sinal_at = NOW(),
  updated_at = NOW()`, strings.TrimSpace(p.IDFranqueado), strings.TrimSpace(p.Fabricante),
		strings.TrimSpace(p.Modulo), strings.TrimSpace(p.Conta), strings.TrimSpace(p.IDDispositivo),
		strings.TrimSpace(p.IPRemoto), strings.TrimSpace(p.UltimoEvento))
	return err
}

func ObterPresenca(fabricante, conta string) (*Presenca, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	var p Presenca
	var ult sql.NullTime
	err = db.QueryRow(`
SELECT id_franqueado, fabricante, modulo, conta, id_dispositivo, ip_remoto, ultimo_evento, ultimo_sinal_at
FROM fp_receptor_presenca
WHERE fabricante = $1 AND conta = $2`, strings.TrimSpace(fabricante), strings.TrimSpace(conta)).Scan(
		&p.IDFranqueado, &p.Fabricante, &p.Modulo, &p.Conta, &p.IDDispositivo, &p.IPRemoto, &p.UltimoEvento, &ult,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if ult.Valid {
		t := ult.Time
		p.UltimoSinal = &t
	}
	return &p, nil
}

func LogDiagnostico(idFra, fabricante, conta, ipTec, status, conclusao string) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
INSERT INTO fp_receptor_diag_log (id_franqueado, fabricante, conta, ip_tecnico, status, conclusao, created_at)
VALUES ($1,$2,$3,$4,$5,$6,NOW())`,
		strings.TrimSpace(idFra), strings.TrimSpace(fabricante), strings.TrimSpace(conta),
		strings.TrimSpace(ipTec), strings.TrimSpace(status), strings.TrimSpace(conclusao))
	return err
}

func ListarPresenca(idFra, fabricante string, limite int) ([]Presenca, error) {
	if limite <= 0 || limite > 200 {
		limite = 50
	}
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	q := `
SELECT id_franqueado, fabricante, modulo, conta, id_dispositivo, ip_remoto, ultimo_evento, ultimo_sinal_at
FROM fp_receptor_presenca WHERE 1=1`
	args := []any{}
	n := 1
	if strings.TrimSpace(idFra) != "" {
		q += ` AND id_franqueado = $` + itoa(n)
		args = append(args, strings.TrimSpace(idFra))
		n++
	}
	if strings.TrimSpace(fabricante) != "" {
		q += ` AND fabricante = $` + itoa(n)
		args = append(args, strings.ToUpper(strings.TrimSpace(fabricante)))
		n++
	}
	q += ` ORDER BY ultimo_sinal_at DESC NULLS LAST LIMIT $` + itoa(n)
	args = append(args, limite)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Presenca
	for rows.Next() {
		var p Presenca
		var ult sql.NullTime
		if err := rows.Scan(&p.IDFranqueado, &p.Fabricante, &p.Modulo, &p.Conta, &p.IDDispositivo, &p.IPRemoto, &p.UltimoEvento, &ult); err != nil {
		 return nil, err
		}
		if ult.Valid {
			t := ult.Time
			p.UltimoSinal = &t
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
