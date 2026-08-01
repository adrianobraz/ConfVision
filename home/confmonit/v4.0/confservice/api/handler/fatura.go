package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"confservice/config"
	"confservice/db"
	"confservice/internal/auth"

	"github.com/google/uuid"
)

// FaturaFechar — admin ConfMonit (X-Api-Key): fecha quinzena de TODOS os parceiros.
// Body opcional: { "periodoInicio":"2026-07-01", "periodoFim":"2026-07-15", "idParceiro":"..." }
func (h *Handler) FaturaFechar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}

	inicio, fim, idParc, err := parsePeriodoBody(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	resumo, err := fecharPeriodo(inicio, fim, idParc)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"periodoInicio": inicio.Format("2006-01-02"),
		"periodoFim":    fim.Format("2006-01-02"),
		"idParceiro":    idParc,
		"resumo":        resumo,
	})
}

// FaturaParceiroFechar — parceiro autenticado fecha só as faturas dele (seus vínculos).
func (h *Handler) FaturaParceiroFechar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	claims, ok := h.requireParceiro(w, r)
	if !ok {
		return
	}

	inicio, fim, _, err := parsePeriodoBody(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	resumo, err := fecharPeriodo(inicio, fim, claims.ParceiroID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"periodoInicio": inicio.Format("2006-01-02"),
		"periodoFim":    fim.Format("2006-01-02"),
		"idParceiro":    claims.ParceiroID,
		"resumo":        resumo,
	})
}

func parsePeriodoBody(r *http.Request) (inicio, fim time.Time, idParceiro string, err error) {
	inicio, fim = quinzenaAtual(time.Now())
	var in struct {
		PeriodoInicio string `json:"periodoInicio"`
		PeriodoFim    string `json:"periodoFim"`
		IDParceiro    string `json:"idParceiro"`
	}
	_ = decodeJSON(r, &in)
	idParceiro = strings.TrimSpace(in.IDParceiro)
	if in.PeriodoInicio != "" && in.PeriodoFim != "" {
		pi, err1 := time.Parse("2006-01-02", in.PeriodoInicio)
		pf, err2 := time.Parse("2006-01-02", in.PeriodoFim)
		if err1 != nil || err2 != nil {
			return inicio, fim, idParceiro, errors.New("periodo invalido (YYYY-MM-DD)")
		}
		inicio, fim = pi, pf
	}
	return inicio, fim, idParceiro, nil
}

func (h *Handler) FaturasFranqueado(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	idFranq := strings.TrimSpace(r.URL.Query().Get("idFranqueado"))
	if idFranq == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado obrigatorio")
		return
	}
	rows, err := db.Conn.Query(`
		SELECT id, periodo_inicio, periodo_fim, valor_bruto, status, created_at
		FROM cs_fatura_franqueado
		WHERE id_franqueado = ?
		ORDER BY periodo_inicio DESC
		LIMIT 24`, idFranq)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, status string
		var ini, fim, created time.Time
		var valor float64
		if err := rows.Scan(&id, &ini, &fim, &valor, &status, &created); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id":            id,
			"periodoInicio": ini.Format("2006-01-02"),
			"periodoFim":    fim.Format("2006-01-02"),
			"valorBruto":    valor,
			"status":        status,
			"createdAt":     created.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) FaturasParceiroMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	claims, ok := h.requireParceiro(w, r)
	if !ok {
		return
	}
	rows, err := db.Conn.Query(`
		SELECT id, periodo_inicio, periodo_fim, valor_bruto, comissao, valor_liquido, status, created_at
		FROM cs_fatura_parceiro
		WHERE id_parceiro = ?
		ORDER BY periodo_inicio DESC
		LIMIT 24`, claims.ParceiroID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, status string
		var ini, fim, created time.Time
		var bruto, comissao, liquido float64
		if err := rows.Scan(&id, &ini, &fim, &bruto, &comissao, &liquido, &status, &created); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id":            id,
			"periodoInicio": ini.Format("2006-01-02"),
			"periodoFim":    fim.Format("2006-01-02"),
			"valorBruto":    bruto,
			"comissao":      comissao,
			"valorLiquido":  liquido,
			"status":        status,
			"createdAt":     created.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) VinculosPorFranqueado(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	idFranq := strings.TrimSpace(r.URL.Query().Get("idFranqueado"))
	if idFranq == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado obrigatorio")
		return
	}
	rows, err := db.Conn.Query(`
		SELECT v.id, v.id_cliente, v.nome_cliente, v.id_parceiro, p.razao_social, p.nome_fantasia,
		       p.software, v.conta_externa, v.preco_congelado, v.inicio_em
		FROM cs_cliente_vinculo v
		JOIN cs_parceiro p ON p.id = v.id_parceiro
		WHERE v.id_franqueado = ? AND v.ativo = 1
		ORDER BY v.nome_cliente, v.id_cliente`, idFranq)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, idCli, idParc, razao, software string
		var nome, fantasia, conta sql.NullString
		var preco float64
		var inicio time.Time
		if err := rows.Scan(&id, &idCli, &nome, &idParc, &razao, &fantasia, &software, &conta, &preco, &inicio); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id":             id,
			"idCliente":      idCli,
			"nomeCliente":    nome.String,
			"idParceiro":     idParc,
			"parceiroNome":   firstNonEmpty(fantasia.String, razao),
			"software":       software,
			"contaExterna":   conta.String,
			"precoCongelado": preco,
			"inicioEm":       inicio.Format("2006-01-02"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

type faturaResumo struct {
	Itens              int     `json:"itens"`
	FaturasFranqueado  int     `json:"faturasFranqueado"`
	FaturasParceiro    int     `json:"faturasParceiro"`
	TotalBruto         float64 `json:"totalBruto"`
	TotalComissao      float64 `json:"totalComissao"`
	TotalLiquidoParceiro float64 `json:"totalLiquidoParceiro"`
}

func fecharPeriodo(inicio, fim time.Time, idParceiroFiltro string) (faturaResumo, error) {
	var res faturaResumo
	iniStr := inicio.Format("2006-01-02")
	fimStr := fim.Format("2006-01-02")

	q := `
		SELECT v.id, v.id_franqueado, v.id_cliente, v.id_parceiro, v.preco_congelado,
		       p.comissao_pct
		FROM cs_cliente_vinculo v
		JOIN cs_parceiro p ON p.id = v.id_parceiro
		WHERE v.ativo = 1
		  AND v.inicio_em <= ?
		  AND (v.fim_em IS NULL OR v.fim_em >= ?)`
	args := []any{fimStr, iniStr}
	if idParceiroFiltro != "" {
		q += ` AND v.id_parceiro = ?`
		args = append(args, idParceiroFiltro)
	}

	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	type item struct {
		vinculoID, idFranq, idCli, idParc string
		preco, comissao, liquido          float64
	}
	itens := []item{}

	for rows.Next() {
		var vinculoID, idFranq, idCli, idParc string
		var preco float64
		var comissaoPct sql.NullFloat64
		if err := rows.Scan(&vinculoID, &idFranq, &idCli, &idParc, &preco, &comissaoPct); err != nil {
			continue
		}
		pct := config.ComissaoPadraoPct
		if comissaoPct.Valid {
			pct = comissaoPct.Float64
		}
		comissao := round2(preco * pct / 100)
		liquido := round2(preco - comissao)
		itens = append(itens, item{vinculoID, idFranq, idCli, idParc, preco, comissao, liquido})
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// Limpa itens do periodo (global ou so do parceiro)
	if idParceiroFiltro == "" {
		_, _ = tx.Exec(`
			DELETE i FROM cs_fatura_item i
			INNER JOIN cs_fatura_franqueado f ON f.id = i.id_fatura_franqueado
			WHERE f.periodo_inicio = ? AND f.periodo_fim = ?`, iniStr, fimStr)
	} else {
		_, _ = tx.Exec(`
			DELETE i FROM cs_fatura_item i
			INNER JOIN cs_fatura_parceiro p ON p.id = i.id_fatura_parceiro
			WHERE p.periodo_inicio = ? AND p.periodo_fim = ? AND i.id_parceiro = ?`,
			iniStr, fimStr, idParceiroFiltro)
	}

	fatFranq := map[string]string{}
	fatParc := map[string]string{}
	brutoFranq := map[string]float64{}
	brutoParc := map[string]float64{}
	comParc := map[string]float64{}
	liqParc := map[string]float64{}

	for _, it := range itens {
		if _, ok := fatFranq[it.idFranq]; !ok {
			id, err := upsertFaturaFranqueado(tx, it.idFranq, iniStr, fimStr)
			if err != nil {
				return res, err
			}
			fatFranq[it.idFranq] = id
			res.FaturasFranqueado++
		}
		if _, ok := fatParc[it.idParc]; !ok {
			id, err := upsertFaturaParceiro(tx, it.idParc, iniStr, fimStr)
			if err != nil {
				return res, err
			}
			fatParc[it.idParc] = id
			res.FaturasParceiro++
		}

		itemID := uuid.NewString()
		_, err = tx.Exec(`
			INSERT INTO cs_fatura_item
			(id, id_fatura_franqueado, id_fatura_parceiro, id_vinculo, id_franqueado, id_cliente, id_parceiro, preco, comissao, liquido_parceiro)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			itemID, fatFranq[it.idFranq], fatParc[it.idParc], it.vinculoID,
			it.idFranq, it.idCli, it.idParc, it.preco, it.comissao, it.liquido,
		)
		if err != nil {
			return res, err
		}
		brutoFranq[it.idFranq] += it.preco
		brutoParc[it.idParc] += it.preco
		comParc[it.idParc] += it.comissao
		liqParc[it.idParc] += it.liquido
		res.Itens++
		res.TotalBruto += it.preco
		res.TotalComissao += it.comissao
		res.TotalLiquidoParceiro += it.liquido
	}

	for idParc, fatID := range fatParc {
		_, err = tx.Exec(`
			UPDATE cs_fatura_parceiro SET valor_bruto = ?, comissao = ?, valor_liquido = ? WHERE id = ?`,
			round2(brutoParc[idParc]), round2(comParc[idParc]), round2(liqParc[idParc]), fatID,
		)
		if err != nil {
			return res, err
		}
	}

	// Franqueado: se fechamento parcial (1 parceiro), total = soma de TODOS os itens do periodo
	if idParceiroFiltro != "" {
		for _, fatID := range fatFranq {
			var total float64
			err = tx.QueryRow(`
				SELECT COALESCE(SUM(preco), 0) FROM cs_fatura_item WHERE id_fatura_franqueado = ?`, fatID,
			).Scan(&total)
			if err != nil {
				return res, err
			}
			_, err = tx.Exec(`UPDATE cs_fatura_franqueado SET valor_bruto = ? WHERE id = ?`, round2(total), fatID)
			if err != nil {
				return res, err
			}
		}
	} else {
		for idFranq, fatID := range fatFranq {
			_, err = tx.Exec(`UPDATE cs_fatura_franqueado SET valor_bruto = ? WHERE id = ?`, round2(brutoFranq[idFranq]), fatID)
			if err != nil {
				return res, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return res, err
	}
	res.TotalBruto = round2(res.TotalBruto)
	res.TotalComissao = round2(res.TotalComissao)
	res.TotalLiquidoParceiro = round2(res.TotalLiquidoParceiro)
	return res, nil
}

func upsertFaturaFranqueado(tx *sql.Tx, idFranq, ini, fim string) (string, error) {
	var id string
	err := tx.QueryRow(`
		SELECT id FROM cs_fatura_franqueado
		WHERE id_franqueado = ? AND periodo_inicio = ? AND periodo_fim = ?`,
		idFranq, ini, fim,
	).Scan(&id)
	if err == nil {
		_, _ = tx.Exec(`UPDATE cs_fatura_franqueado SET valor_bruto = 0, status = 'ABERTA' WHERE id = ?`, id)
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	id = uuid.NewString()
	_, err = tx.Exec(`
		INSERT INTO cs_fatura_franqueado (id, id_franqueado, periodo_inicio, periodo_fim, valor_bruto, status)
		VALUES (?, ?, ?, ?, 0, 'ABERTA')`, id, idFranq, ini, fim)
	return id, err
}

func upsertFaturaParceiro(tx *sql.Tx, idParc, ini, fim string) (string, error) {
	var id string
	err := tx.QueryRow(`
		SELECT id FROM cs_fatura_parceiro
		WHERE id_parceiro = ? AND periodo_inicio = ? AND periodo_fim = ?`,
		idParc, ini, fim,
	).Scan(&id)
	if err == nil {
		_, _ = tx.Exec(`UPDATE cs_fatura_parceiro SET valor_bruto = 0, comissao = 0, valor_liquido = 0, status = 'ABERTA' WHERE id = ?`, id)
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	id = uuid.NewString()
	_, err = tx.Exec(`
		INSERT INTO cs_fatura_parceiro (id, id_parceiro, periodo_inicio, periodo_fim, valor_bruto, comissao, valor_liquido, status)
		VALUES (?, ?, ?, ?, 0, 0, 0, 'ABERTA')`, id, idParc, ini, fim)
	return id, err
}

func quinzenaAtual(now time.Time) (time.Time, time.Time) {
	y, m, d := now.Date()
	loc := now.Location()
	if d <= 15 {
		return time.Date(y, m, 1, 0, 0, 0, 0, loc), time.Date(y, m, 15, 0, 0, 0, 0, loc)
	}
	inicio := time.Date(y, m, 16, 0, 0, 0, 0, loc)
	fim := time.Date(y, m+1, 0, 0, 0, 0, 0, loc) // ultimo dia do mes
	return inicio, fim
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
