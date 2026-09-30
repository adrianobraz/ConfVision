package contrato

import (
	"apifunction/db"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

func isCentralUUID(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && strings.ToUpper(s) != "CENTRAL" && len(s) >= 20
}

func SeedCatalogo(idCentral string) (int, error) {
	idCentral = strings.TrimSpace(idCentral)
	if !isCentralUUID(idCentral) {
		return 0, errors.New("id_central obrigatorio (IDCentralUUID)")
	}
	n := 0

	produtos := []struct {
		prod, plano, nome, grupo, sel string
		valor                           float64
		retencao                        int
		ordem                           int
	}{
		{"franqueadopro", "lite", "FranqueadoPro Lite", "franqueadopro", "radio", 99, 30, 1},
		{"franqueadopro", "pro", "FranqueadoPro Pro", "franqueadopro", "radio", 199, 90, 2},
		{"franqueadopro", "pro_plus", "FranqueadoPro Pro+", "franqueadopro", "radio", 299, 180, 3},
		{"webterminal", "lite", "WebTerminal Lite", "webterminal", "radio", 49, 30, 10},
		{"webterminal", "pro", "WebTerminal Pro", "webterminal", "radio", 79, 90, 11},
		{"webterminal", "pro_plus", "WebTerminal Pro+", "webterminal", "radio", 129, 180, 12},
		{"terminalmovel", "padrao", "Terminal Movel", "terminalmovel", "checkbox", 59, 30, 20},
		{"confvision", "padrao", "ConfVision - Plano (software)", "confvision", "checkbox", 79, 30, 21},
		{"webambiente", "pro", "webAmbiente Pro", "webambiente", "radio", 69, 90, 30},
		{"webambiente", "pro_plus", "webAmbiente Pro+", "webambiente", "radio", 99, 180, 31},
		{"dialyze", "padrao", "Dialyze", "dialyze", "checkbox", 99, 30, 40},
	}

	for _, p := range produtos {
		modJSON, _ := json.Marshal(modulosPorPlanoFP(p.plano))
		if p.prod != "franqueadopro" {
			modJSON = []byte("{}")
		}
		limJSON, _ := json.Marshal(limitesPorPlanoFP(p.plano))
		if p.prod != "franqueadopro" {
			limJSON = []byte("{}")
		}
		res, err := db.Conn.Exec(`
			INSERT IGNORE INTO fp_catalogo_produto
			(ID_Central, Produto, Plano, NomeExibicao, GrupoUI, TipoSelecao, ValorMensal, RetencaoDias, ModulosJSON, LimitesJSON, Ordem, Ativo)
			VALUES (?,?,?,?,?,?,?,?,?,?,?, 'S')
		`, idCentral, p.prod, p.plano, p.nome, p.grupo, p.sel, p.valor, p.retencao, modJSON, limJSON, p.ordem)
		if err != nil {
			return n, err
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			n++
		}
	}

	pacotes := []struct {
		nome string
		qtd  int
		val  float64
	}{
		{"Cota 50", 50, 50},
		{"Cota 200", 200, 100},
		{"Cota 800", 800, 200},
	}
	for _, p := range pacotes {
		var exists int
		_ = db.Conn.QueryRow(`
			SELECT COUNT(*) FROM fp_pacote_cota WHERE ID_Central = ? AND Quantidade = ?
		`, idCentral, p.qtd).Scan(&exists)
		if exists > 0 {
			continue
		}
		_, err := db.Conn.Exec(`
			INSERT INTO fp_pacote_cota (ID_Central, Nome, Quantidade, Valor, Ativo) VALUES (?,?,?,?, 'S')
		`, idCentral, p.nome, p.qtd, p.val)
		if err != nil {
			return n, err
		}
		n++
	}

	cvLic := []struct {
		plano, nome, unidade string
		valor                float64
		ordem                int
	}{
		{"online", "Camera online", "camera", 2.99, 1},
		{"sensor_foto", "Sensor foto", "camera", 7.99, 2},
		{"sensor_foto_video", "Sensor foto + video", "camera", 9.99, 3},
		{"analitico_armado_evento", "Analitico armado — so evento", "camera", 11.99, 4},
		{"analitico_armado_foto", "Analitico armado — foto", "camera", 13.99, 5},
		{"analitico_armado_foto_video", "Analitico armado — foto + video", "camera", 14.99, 6},
		{"analitico_24h_evento", "Analitico 24h — so evento", "camera", 16.99, 7},
		{"analitico_24h_foto", "Analitico 24h — foto", "camera", 18.99, 8},
		{"analitico_24h_foto_video", "Analitico 24h — foto + video", "camera", 19.99, 9},
		{"gravacao_7d", "Gravacao continua 7 dias", "gravacao", 12.99, 10},
		{"gravacao_15d", "Gravacao continua 15 dias", "gravacao", 17.99, 11},
		{"gravacao_30d", "Gravacao continua 30 dias", "gravacao", 24.99, 12},
		{"gravacao_movimento_7d", "Gravacao por movimento 7 dias", "gravacao", 9.99, 13},
		{"gravacao_movimento_15d", "Gravacao por movimento 15 dias", "gravacao", 13.99, 14},
		{"gravacao_movimento_30d", "Gravacao por movimento 30 dias", "gravacao", 19.99, 15},
		{"gravacao_timelapse_7d", "Gravacao timelapse 7 dias", "gravacao", 6.99, 16},
		{"gravacao_timelapse_15d", "Gravacao timelapse 15 dias", "gravacao", 9.99, 17},
		{"gravacao_timelapse_30d", "Gravacao timelapse 30 dias", "gravacao", 13.99, 18},
	}
	for _, c := range cvLic {
		res, err := db.Conn.Exec(`
			INSERT IGNORE INTO fp_catalogo_cv_licenca
			(ID_Central, Plano, NomeExibicao, Unidade, ValorMensal, Ordem, Ativo)
			VALUES (?,?,?,?,?,?,'S')
		`, idCentral, c.plano, c.nome, c.unidade, c.valor, c.ordem)
		if err != nil {
			return n, err
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			n++
		}
	}
	return n, nil
}

func ListarCatalogoProdutos(idCentral string) ([]CatalogoProduto, error) {
	rows, err := db.Conn.Query(`
		SELECT ID_CatalogoProduto, Produto, Plano, NomeExibicao, COALESCE(GrupoUI,''), TipoSelecao,
		       ValorMensal, COALESCE(RetencaoDias,30), Ordem
		FROM fp_catalogo_produto
		WHERE ID_Central = ? AND Ativo = 'S'
		ORDER BY Ordem, Produto, Plano
	`, idCentral)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CatalogoProduto
	for rows.Next() {
		var c CatalogoProduto
		if err := rows.Scan(&c.ID, &c.Produto, &c.Plano, &c.NomeExibicao, &c.GrupoUI, &c.TipoSelecao, &c.ValorMensal, &c.RetencaoDias, &c.Ordem); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func ListarPacotesCota(idCentral string) ([]PacoteCota, error) {
	rows, err := db.Conn.Query(`
		SELECT ID_PacoteCota, Nome, Quantidade, Valor, Ativo
		FROM fp_pacote_cota WHERE ID_Central = ? AND Ativo = 'S'
		ORDER BY Quantidade
	`, idCentral)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PacoteCota
	for rows.Next() {
		var p PacoteCota
		if err := rows.Scan(&p.ID, &p.Nome, &p.Quantidade, &p.Valor, &p.Ativo); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func ListarCVLicencas(idCentral string) ([]CVLicencaCat, error) {
	rows, err := db.Conn.Query(`
		SELECT ID_CatalogoCV, Plano, NomeExibicao, Unidade, ValorMensal, Ordem
		FROM fp_catalogo_cv_licenca WHERE ID_Central = ? AND Ativo = 'S'
		ORDER BY Ordem
	`, idCentral)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CVLicencaCat
	for rows.Next() {
		var c CVLicencaCat
		if err := rows.Scan(&c.ID, &c.Plano, &c.NomeExibicao, &c.Unidade, &c.ValorMensal, &c.Ordem); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func catalogoCount(idCentral, table string) (int, error) {
	var n int
	q := "SELECT COUNT(*) FROM " + table + " WHERE ID_Central = ?"
	err := db.Conn.QueryRow(q, idCentral).Scan(&n)
	return n, err
}

func EnsureCatalogo(idCentral string) error {
	n, err := catalogoCount(idCentral, "fp_catalogo_produto")
	if err != nil {
		return err
	}
	if n == 0 {
		_, err = SeedCatalogo(idCentral)
	}
	return err
}

func FranqueadoRepresentante(idFranqueado string) (idRep string, idCentral string, err error) {
	return franqueadoRepresentante(idFranqueado)
}

func franqueadoRepresentante(idFranqueado string) (idRep string, idCentral string, err error) {
	err = db.Conn.QueryRow(`
		SELECT COALESCE(f.ID_Representante,''),
			COALESCE(NULLIF(TRIM(c.IDCentralUUID), ''), NULLIF(TRIM(c.ID_Central), ''), '')
		FROM franqueado f
		LEFT JOIN representante r ON f.ID_Representante = r.ID_Representante
		LEFT JOIN central c ON r.IDCentralUUID = c.IDCentralUUID OR c.ID_Central = r.IDCentralUUID
		WHERE f.ID_Franqueado = ?
		LIMIT 1
	`, idFranqueado).Scan(&idRep, &idCentral)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return strings.TrimSpace(idRep), strings.TrimSpace(idCentral), err
}

func RepresentanteUsaAdm(idRep string) (string, error) {
	return representanteUsaAdm(idRep)
}

func representanteUsaAdm(idRep string) (string, error) {
	if idRep == "" {
		return "N", nil
	}
	var flag sql.NullString
	err := db.Conn.QueryRow(`
		SELECT COALESCE(UsaAdmConfmonit,'N') FROM representante WHERE ID_Representante = ?
	`, idRep).Scan(&flag)
	if err == sql.ErrNoRows {
		return "N", nil
	}
	f := strings.ToUpper(strings.TrimSpace(flag.String))
	if f != "S" {
		f = "N"
	}
	return f, err
}

func centralDoFranqueado(idFranqueado, fallback string) string {
	_, idCentral, err := franqueadoRepresentante(idFranqueado)
	if err == nil && isCentralUUID(idCentral) {
		return strings.TrimSpace(idCentral)
	}
	if isCentralUUID(fallback) {
		return strings.TrimSpace(fallback)
	}
	return ""
}
