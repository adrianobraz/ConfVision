package financeiro

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"apifunction/auth"
	"apifunction/config"
	"apifunction/db"
	"apifunction/pgfinmirror"
	"apifunction/xano"
)

type ListarFiltro struct {
	Status          string
	Tipo            string
	IDFranqueado    string
	IDRepresentante string
	AdminToken      string
}

type FaturaItem struct {
	ID            int64   `json:"id"`
	FPFaturaID    int64   `json:"fp_fatura_id"`
	Descricao     string  `json:"descricao"`
	Quantidade    int     `json:"quantidade"`
	ValorUnitario float64 `json:"valor_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	RefTipo       string  `json:"ref_tipo"`
	RefID         string  `json:"ref_id"`
	ValorPiso     float64 `json:"valor_piso,omitempty"`
	MargemCentral float64 `json:"margem_central,omitempty"`
	MargemRep     float64 `json:"margem_rep,omitempty"`
}

type Fatura struct {
	ID                  int64        `json:"id"`
	CreatedAt           *time.Time   `json:"created_at,omitempty"`
	IDFranqueado        string       `json:"id_franqueado"`
	IDRepresentante     string       `json:"id_representante"`
	IDCentral           string       `json:"id_central"`
	Referencia          string       `json:"referencia"`
	Status              string       `json:"status"`
	Tipo                string       `json:"tipo"`
	ValorTotal          float64      `json:"valor_total"`
	ValorPisoCentral    float64      `json:"valor_piso_central,omitempty"`
	ValorPisoBreakglass float64      `json:"valor_piso_breakglass,omitempty"`
	MargemCentral       float64      `json:"margem_central,omitempty"`
	MargemRep           float64      `json:"margem_rep,omitempty"`
	FaturaOrigemID      int64        `json:"fatura_origem_id,omitempty"`
	VencimentoEm        *time.Time   `json:"vencimento_em,omitempty"`
	PagoEm              *time.Time   `json:"pago_em,omitempty"`
	CicloRef            string       `json:"ciclo_ref"`
	Observacao          string       `json:"observacao"`
	Itens               []FaturaItem `json:"itens,omitempty"`
}

func ListarFaturas(ctx context.Context, sess auth.SessaoAdm, f ListarFiltro) ([]Fatura, string, error) {
	if idCentralSessao(sess) == "" {
		return nil, "", fmt.Errorf("sessao sem idCentral")
	}

	fonte := strings.ToLower(strings.TrimSpace(config.FaturaListarFonte))
	if fonte == "" {
		fonte = "auto"
	}

	if fonte == "postgres" || fonte == "auto" {
		ready, motivo, _ := pgfinmirror.IsReady(ctx)
		if ready {
			out, err := listarFaturasPostgres(ctx, sess, f)
			if err == nil {
				if out == nil {
					out = []Fatura{}
				}
				return out, "postgres_mirror", nil
			}
			if fonte == "postgres" {
				return nil, "postgres_mirror", err
			}
			_ = motivo
		} else if fonte == "postgres" {
			return nil, "postgres_mirror", fmt.Errorf("espelho nao pronto: %s", motivo)
		}
	}

	if fonte == "meta" || fonte == "auto" {
		if config.XanoMetaAccessToken != "" {
			out, err := listarFaturasMeta(ctx, sess, f)
			if err != nil {
				return nil, "xano_meta", err
			}
			if out == nil {
				out = []Fatura{}
			}
			return out, "xano_meta", nil
		}

		if config.XanoAPIFinanceiro != "" && strings.TrimSpace(f.AdminToken) != "" {
			out, err := listarFaturasXanoHTTP(f)
			if err != nil {
				return nil, "xano_http", err
			}
			if len(out) > 0 || sess.UserTipo == "REP" {
				if out == nil {
					out = []Fatura{}
				}
				return out, "xano_http", nil
			}
		}

		if fonte == "meta" {
			return nil, "", fmt.Errorf("configure XANO_META_ACCESS_TOKEN no apifunction")
		}
	}

	return nil, "", fmt.Errorf("fatura listar indisponivel — configure POSTGRES_URL ou XANO_META_ACCESS_TOKEN")
}

// idCentralSessao — preferir IDCentralUUID (usuarios/representante); fallback catálogo.
func idCentralSessao(sess auth.SessaoAdm) string {
	return IDCentralSessaoExport(sess)
}

// IDCentralSessaoExport expõe id central da sessão para outros pacotes (finresumo).
func IDCentralSessaoExport(sess auth.SessaoAdm) string {
	if u := strings.TrimSpace(sess.IDCentralUUID); u != "" {
		return u
	}
	return strings.TrimSpace(sess.IDCentralCatalogo)
}

// ChavesCentral expande ID_Central e IDCentralUUID equivalentes na tabela central.
func ChavesCentral(ctx context.Context, idCentral string) (map[string]struct{}, error) {
	return chavesCentral(ctx, idCentral)
}

// chavesCentral expande ID_Central e IDCentralUUID equivalentes na tabela central.
func chavesCentral(ctx context.Context, idCentral string) (map[string]struct{}, error) {
	idCentral = strings.TrimSpace(idCentral)
	out := map[string]struct{}{}
	if idCentral == "" {
		return out, nil
	}
	out[idCentral] = struct{}{}

	rows, err := db.Conn.QueryContext(ctx, `
		SELECT ID_Central, IDCentralUUID
		FROM central
		WHERE TRIM(COALESCE(ID_Central, '')) = ?
		   OR TRIM(COALESCE(IDCentralUUID, '')) = ?
	`, idCentral, idCentral)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var idCat, idUUID sql.NullString
		if err := rows.Scan(&idCat, &idUUID); err != nil {
			return nil, err
		}
		if v := strings.TrimSpace(idCat.String); v != "" {
			out[v] = struct{}{}
		}
		if v := strings.TrimSpace(idUUID.String); v != "" {
			out[v] = struct{}{}
		}
	}
	return out, rows.Err()
}

func repsDaCentral(ctx context.Context, idCentral string) ([]string, error) {
	keys, err := chavesCentral(ctx, idCentral)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}

	vals := make([]string, 0, len(keys))
	for k := range keys {
		vals = append(vals, k)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(vals)), ",")
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}

	q := fmt.Sprintf(`
		SELECT ID_Representante
		FROM representante
		WHERE TRIM(COALESCE(ID_Representante, '')) != ''
		  AND TRIM(COALESCE(IDCentralUUID, '')) IN (%s)
	`, placeholders)

	rows, err := db.Conn.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reps []string
	seen := map[string]struct{}{}
	for rows.Next() {
		var id sql.NullString
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		r := strings.TrimSpace(id.String)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		reps = append(reps, r)
	}
	return reps, rows.Err()
}

func franqueadosDaCentral(ctx context.Context, reps []string) ([]string, error) {
	if len(reps) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(reps)), ",")
	args := make([]any, len(reps))
	for i, r := range reps {
		args[i] = r
	}
	q := fmt.Sprintf(`
		SELECT ID_Franqueado
		FROM franqueado
		WHERE TRIM(COALESCE(ID_Franqueado, '')) != ''
		  AND ID_Representante IN (%s)
	`, placeholders)
	rows, err := db.Conn.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	seen := map[string]struct{}{}
	for rows.Next() {
		var id sql.NullString
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		f := strings.TrimSpace(id.String)
		if f == "" {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		ids = append(ids, f)
	}
	return ids, rows.Err()
}

func listarFaturasXanoHTTP(f ListarFiltro) ([]Fatura, error) {
	cli := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	if cli == nil || !cli.Enabled() {
		return nil, fmt.Errorf("XANO_API_FINANCEIRO nao configurado")
	}
	payload := map[string]any{"admin_token": strings.TrimSpace(f.AdminToken)}
	if st := strings.TrimSpace(f.Status); st != "" {
		payload["status"] = st
	}
	if tp := strings.TrimSpace(f.Tipo); tp != "" {
		payload["tipo"] = tp
	}
	if idf := strings.TrimSpace(f.IDFranqueado); idf != "" {
		payload["id_franqueado"] = idf
	}
	if idr := strings.TrimSpace(f.IDRepresentante); idr != "" {
		payload["id_representante"] = idr
	}
	raw, err := cli.Post("/fp_fatura_listar", payload)
	if err != nil {
		return nil, err
	}
	return parseFaturasResposta(raw)
}

func parseFaturasResposta(raw map[string]any) ([]Fatura, error) {
	dados, ok := raw["dados"]
	if !ok || dados == nil {
		return nil, nil
	}
	b, err := json.Marshal(dados)
	if err != nil {
		return nil, err
	}
	var out []Fatura
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
