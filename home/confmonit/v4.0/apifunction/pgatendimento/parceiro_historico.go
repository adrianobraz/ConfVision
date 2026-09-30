package pgatendimento

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

const (
	HistEventoExcecaoProposta  = "excecao_proposta"
	HistEventoExcecaoAtivada   = "excecao_ativada"
	HistEventoExcecaoCancelada = "excecao_cancelada"
)

type HistoricoRegistro struct {
	ID             int64          `json:"id"`
	CreatedAt      string         `json:"created_at"`
	IDFranqueado   string         `json:"id_franqueado"`
	IDCliente      string         `json:"id_cliente"`
	NomeCliente    string         `json:"nome_cliente"`
	Evento         string         `json:"evento"`
	IDParceiroDe   string         `json:"id_parceiro_de"`
	IDParceiroPara string         `json:"id_parceiro_para"`
	IDVinculoDe    string         `json:"id_vinculo_de"`
	IDVinculoPara  string         `json:"id_vinculo_para"`
	FPFaturaID     int            `json:"fp_fatura_id,omitempty"`
	Detalhe        map[string]any `json:"detalhe,omitempty"`
}

type HistoricoInput struct {
	IDFranqueado   string
	IDCliente      string
	NomeCliente    string
	Evento         string
	IDParceiroDe   string
	IDParceiroPara string
	IDVinculoDe    string
	IDVinculoPara  string
	FPFaturaID     int
	Detalhe        map[string]any
}

func RegistrarHistoricoParceiro(ctx context.Context, in HistoricoInput) error {
	in.IDFranqueado = strings.TrimSpace(in.IDFranqueado)
	in.Evento = strings.TrimSpace(in.Evento)
	if in.IDFranqueado == "" || in.Evento == "" {
		return nil
	}
	d, err := db(ctx)
	if err != nil {
		return err
	}
	if in.Detalhe == nil {
		in.Detalhe = map[string]any{}
	}
	raw, err := json.Marshal(in.Detalhe)
	if err != nil {
		return err
	}
	var fp any
	if in.FPFaturaID > 0 {
		fp = in.FPFaturaID
	}
	_, err = d.ExecContext(ctx, `
INSERT INTO ops_parceiro_historico (
    id_franqueado, id_cliente, nome_cliente, evento,
    id_parceiro_de, id_parceiro_para, id_vinculo_de, id_vinculo_para,
    fp_fatura_id, detalhe
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb)`,
		in.IDFranqueado, strings.TrimSpace(in.IDCliente), strings.TrimSpace(in.NomeCliente), in.Evento,
		strings.TrimSpace(in.IDParceiroDe), strings.TrimSpace(in.IDParceiroPara),
		strings.TrimSpace(in.IDVinculoDe), strings.TrimSpace(in.IDVinculoPara),
		fp, string(raw),
	)
	return err
}

func ListHistoricoParceiro(ctx context.Context, idFranqueado string, limite int) ([]HistoricoRegistro, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, nil
	}
	if limite <= 0 || limite > 200 {
		limite = 50
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.QueryContext(ctx, `
SELECT id, created_at, id_franqueado, id_cliente, nome_cliente, evento,
       id_parceiro_de, id_parceiro_para, id_vinculo_de, id_vinculo_para,
       fp_fatura_id, COALESCE(detalhe::text, '{}')
FROM ops_parceiro_historico
WHERE id_franqueado = $1
ORDER BY created_at DESC, id DESC
LIMIT $2`, idFranqueado, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HistoricoRegistro
	for rows.Next() {
		var rec HistoricoRegistro
		var created time.Time
		var fp sql.NullInt64
		var detRaw string
		if err := rows.Scan(
			&rec.ID, &created, &rec.IDFranqueado, &rec.IDCliente, &rec.NomeCliente, &rec.Evento,
			&rec.IDParceiroDe, &rec.IDParceiroPara, &rec.IDVinculoDe, &rec.IDVinculoPara,
			&fp, &detRaw,
		); err != nil {
			continue
		}
		rec.CreatedAt = created.Format(time.RFC3339)
		if fp.Valid {
			rec.FPFaturaID = int(fp.Int64)
		}
		if strings.TrimSpace(detRaw) != "" {
			_ = json.Unmarshal([]byte(detRaw), &rec.Detalhe)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}
