package credito

import (
	"apifunction/auth"
	"apifunction/config"
	"apifunction/db"
	"apifunction/service/contrato"
	"apifunction/xano"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	MinRecarga        = 50.0
	TipoFaturaRecarga = "recarga_credito_servicos"
)

type SolicitarInput struct {
	IDFranqueado      string  `json:"id_franqueado"`
	Valor             float64 `json:"valor"`
	SolicitadoPor     string  `json:"solicitado_por"`
	SolicitadoUsuario string  `json:"solicitado_usuario"`
	AdminUsuario      string  `json:"admin_usuario"` // alias admConfmonit legado
}

type RecargaRes struct {
	IDRecarga        int     `json:"id_recarga"`
	IDFaturaContabil int     `json:"id_fatura_contabil"`
	Valor            float64 `json:"valor"`
	Status           string  `json:"status"`
}

func SolicitarRecarga(sess auth.SessaoAdm, in SolicitarInput) (RecargaRes, error) {
	out := RecargaRes{}
	if strings.TrimSpace(in.SolicitadoUsuario) == "" {
		in.SolicitadoUsuario = strings.TrimSpace(in.AdminUsuario)
	}
	idFra := strings.TrimSpace(in.IDFranqueado)
	if idFra == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	if in.Valor < MinRecarga {
		return out, fmt.Errorf("valor minimo R$ %.2f", MinRecarga)
	}

	solicitadoPor := strings.ToUpper(strings.TrimSpace(in.SolicitadoPor))
	if solicitadoPor == "" {
		if strings.TrimSpace(sess.IDUsuario) == "BREAKGLASS" {
			solicitadoPor = "BG"
		} else if sess.UserTipo == "REP" {
			solicitadoPor = "REP"
		} else if sess.UserTipo == "CEN" {
			solicitadoPor = "CEN"
		} else {
			solicitadoPor = "FRA"
		}
	}

	if solicitadoPor != "FRA" {
		if err := contrato.AssertPodeEditarContrato(sess, idFra); err != nil {
			return out, err
		}
	}

	idRep, idCentral, err := contrato.FranqueadoRepresentante(idFra)
	if err != nil {
		return out, err
	}
	if idCentral == "" {
		idCentral = sess.IDCentralCatalogo
	}

	res, err := db.Conn.Exec(`
INSERT INTO fp_credito_recarga
  (ID_Franqueado, ID_Central, ID_Representante, Canal, Valor, Status, SolicitadoPor, SolicitadoUsuario)
VALUES (?,?,?, 'credito', ?, 'aberta', ?, ?)`,
		idFra, idCentral, nullIfEmpty(idRep), in.Valor, solicitadoPor, strings.TrimSpace(in.SolicitadoUsuario))
	if err != nil {
		return out, err
	}
	idRec, _ := res.LastInsertId()

	xc := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	idFatContabil := 0
	if xc != nil && xc.Enabled() {
		desc := fmt.Sprintf("Recarga credito servicos — R$ %.2f", in.Valor)
		syncOut, err := xc.Post("/fp_credito_recarga_sync", map[string]any{
			"worker_key":       config.WorkerSecret,
			"id_recarga_mysql": idRec,
			"id_franqueado":    idFra,
			"id_central":       idCentral,
			"id_representante": idRep,
			"canal":            "credito",
			"valor_total":      in.Valor,
			"admin_usuario":    in.SolicitadoUsuario,
			"descricao":        desc,
		})
		if err != nil {
			return out, fmt.Errorf("sync fatura: %w", err)
		}
		idFatContabil = intFromMap(syncOut, "id_fatura_contabil")
		if idFatContabil == 0 {
			idFatContabil = intFromMap(syncOut, "id_fatura")
		}
		_, _ = db.Conn.Exec(`UPDATE fp_credito_recarga SET ID_FaturaContabil = ? WHERE ID_Recarga = ?`, idFatContabil, idRec)
	}

	out = RecargaRes{
		IDRecarga:        int(idRec),
		IDFaturaContabil: idFatContabil,
		Valor:            in.Valor,
		Status:           "aberta",
	}
	return out, nil
}

func ConfirmarRecarga(idFaturaContabil int, adminUsuario string) error {
	if idFaturaContabil <= 0 {
		return errors.New("id_fatura_contabil obrigatorio")
	}
	var idRec int
	var idFra, status, aplicado string
	var valor float64
	err := db.Conn.QueryRow(`
SELECT ID_Recarga, ID_Franqueado, Valor, Status, CreditoAplicado
FROM fp_credito_recarga WHERE ID_FaturaContabil = ? LIMIT 1`, idFaturaContabil).Scan(
		&idRec, &idFra, &valor, &status, &aplicado)
	if err == sql.ErrNoRows {
		return errors.New("recarga nao encontrada para esta fatura")
	}
	if err != nil {
		return err
	}
	if aplicado == "S" {
		return nil
	}
	if status != "aberta" && status != "paga" {
		return fmt.Errorf("recarga status invalido: %s", status)
	}

	obs := fmt.Sprintf("Recarga paga fatura #%d", idFaturaContabil)
	if err := creditarWallet(idFra, "credito", "recarga", valor, int64(idFaturaContabil), obs, adminUsuario); err != nil {
		return err
	}
	_, err = db.Conn.Exec(`
UPDATE fp_credito_recarga SET Status = 'paga', CreditoAplicado = 'S', PagoEm = NOW() WHERE ID_Recarga = ?`, idRec)
	return err
}

func ConfirmarRecargaPorIDRecarga(idRecarga int, adminUsuario string) error {
	var idFat sql.NullInt64
	err := db.Conn.QueryRow(`SELECT ID_FaturaContabil FROM fp_credito_recarga WHERE ID_Recarga = ?`, idRecarga).Scan(&idFat)
	if err != nil {
		return err
	}
	if !idFat.Valid || idFat.Int64 <= 0 {
		return errors.New("recarga sem fatura contabil")
	}
	return ConfirmarRecarga(int(idFat.Int64), adminUsuario)
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func intFromMap(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		if d, ok := m["dados"].(map[string]any); ok {
			return intFromMap(d, key)
		}
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func ListarRecargas(idFranqueado string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `
SELECT ID_Recarga, ID_Franqueado, Canal, Valor, Status, SolicitadoPor,
       ID_FaturaContabil, CreditoAplicado, CreatedAt, PagoEm
FROM fp_credito_recarga`
	args := []any{}
	if strings.TrimSpace(idFranqueado) != "" {
		q += ` WHERE ID_Franqueado = ?`
		args = append(args, idFranqueado)
	}
	q += ` ORDER BY ID_Recarga DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var idRec int
		var idFra, canal, status, solPor, aplicado string
		var valor float64
		var idFat sql.NullInt64
		var criado, pago sql.NullTime
		if err := rows.Scan(&idRec, &idFra, &canal, &valor, &status, &solPor, &idFat, &aplicado, &criado, &pago); err != nil {
			return nil, err
		}
		item := map[string]any{
			"id_recarga": idRec, "id_franqueado": idFra, "canal": canal,
			"valor": valor, "status": status, "solicitado_por": solPor,
			"credito_aplicado": aplicado,
		}
		if idFat.Valid {
			item["id_fatura_contabil"] = idFat.Int64
		}
		if criado.Valid {
			item["created_at"] = criado.Time.Format(time.RFC3339)
		}
		if pago.Valid {
			item["pago_em"] = pago.Time.Format(time.RFC3339)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
