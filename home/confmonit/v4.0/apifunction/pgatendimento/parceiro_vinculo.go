package pgatendimento

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	mysqldb "apifunction/db"
)

type ParceiroVinculoResult struct {
	OK            bool   `json:"ok"`
	IDParceiro    string `json:"id_parceiro"`
	CodigoInterno string `json:"codigo_interno"`
	Origem        string `json:"origem"`
	NomeCliente   string `json:"nome_cliente,omitempty"`
}

type ParceiroEnvioLogItem struct {
	ID            int64          `json:"id"`
	CreatedAt     string         `json:"created_at"`
	IDCliente     string         `json:"id_cliente"`
	IDProcesso    string         `json:"id_processo"`
	AlarmEventsID *int64         `json:"alarm_events_id,omitempty"`
	IDParceiro    string         `json:"id_parceiro"`
	CtiGrupo      string         `json:"cti_grupo"`
	AcaoFinal     string         `json:"acao_final"`
	Detalhe       map[string]any `json:"detalhe,omitempty"`
}

func confServiceBase() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("CONFSERVICE_URL")), "/")
}

func confServiceKey() string {
	return strings.TrimSpace(os.Getenv("CONFSERVICE_API_KEY"))
}

func GetClienteCodigoInterno(idCliente string) (string, error) {
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" || mysqldb.Conn == nil {
		return "", nil
	}
	var codigo sql.NullString
	err := mysqldb.Conn.QueryRow(`SELECT CodigoInterno FROM cliente WHERE ID_Cliente = ? LIMIT 1`, idCliente).Scan(&codigo)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if codigo.Valid {
		return strings.ToUpper(strings.TrimSpace(codigo.String)), nil
	}
	return "", nil
}

func lookupExcecaoParceiro(ctx context.Context, idFranqueado, idCliente string) (idParceiro string, err error) {
	base := confServiceBase()
	if base == "" {
		return "", nil
	}
	q := url.Values{}
	q.Set("idFranqueado", idFranqueado)
	q.Set("idCliente", idCliente)
	rawURL := base + "/internal/vinculo/cliente?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	if key := confServiceKey(); key != "" {
		req.Header.Set("X-Api-Key", key)
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("confservice status=%d", resp.StatusCode)
	}
	var parsed struct {
		OK      bool `json:"ok"`
		Vinculo *struct {
			IDParceiro string `json:"idParceiro"`
		} `json:"vinculo"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.Vinculo != nil {
		return strings.TrimSpace(parsed.Vinculo.IDParceiro), nil
	}
	return "", nil
}

func ResolveParceiroVinculo(ctx context.Context, idFranqueado, idCliente, ctiGrupo string) (ParceiroVinculoResult, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	out := ParceiroVinculoResult{}
	if idFranqueado == "" || idCliente == "" {
		return out, nil
	}

	p, err := GetPolitica(ctx, idFranqueado)
	if err != nil {
		return out, err
	}
	ativo, err := ResolveComHorario(ctx, idFranqueado, idCliente, RecursoParceiro, ctiGrupo, &p)
	if err != nil {
		return out, err
	}
	if !ativo {
		return out, nil
	}

	idParceiro, err := lookupExcecaoParceiro(ctx, idFranqueado, idCliente)
	if err != nil {
		return out, err
	}
	origem := "excecao"
	if idParceiro == "" {
		idParceiro = strings.TrimSpace(p.IDParceiro)
		origem = "padrao"
	}
	if idParceiro == "" {
		return out, nil
	}

	codigo, err := GetClienteCodigoInterno(idCliente)
	if err != nil {
		return out, err
	}

	out.OK = true
	out.IDParceiro = idParceiro
	out.CodigoInterno = codigo
	out.Origem = origem
	return out, nil
}

func ListParceiroEnvioLog(ctx context.Context, idFranqueado string, limite int) ([]ParceiroEnvioLogItem, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, errors.New("id_franqueado obrigatorio")
	}
	if limite <= 0 || limite > 200 {
		limite = 50
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.QueryContext(ctx, `
SELECT id, created_at, id_cliente, COALESCE(id_processo,''), alarm_events_id,
       COALESCE(id_parceiro,''), COALESCE(cti_grupo,''), COALESCE(acao_final,''), COALESCE(detalhe::text,'{}')
FROM ops_parceiro_envio_log
WHERE id_franqueado = $1
ORDER BY created_at DESC
LIMIT $2`, idFranqueado, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ParceiroEnvioLogItem
	for rows.Next() {
		var it ParceiroEnvioLogItem
		var created time.Time
		var alarm sql.NullInt64
		var detalheRaw string
		if err := rows.Scan(&it.ID, &created, &it.IDCliente, &it.IDProcesso, &alarm,
			&it.IDParceiro, &it.CtiGrupo, &it.AcaoFinal, &detalheRaw); err != nil {
			continue
		}
		it.CreatedAt = created.Format(time.RFC3339)
		if alarm.Valid {
			v := alarm.Int64
			it.AlarmEventsID = &v
		}
		if strings.TrimSpace(detalheRaw) != "" {
			_ = json.Unmarshal([]byte(detalheRaw), &it.Detalhe)
		}
		list = append(list, it)
	}
	return list, rows.Err()
}
