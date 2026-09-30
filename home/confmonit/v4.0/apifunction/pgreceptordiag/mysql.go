package pgreceptordiag

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"apifunction/db"
)

func ModuloForFabricante(nome string) string {
	n := strings.ToUpper(strings.TrimSpace(nome))
	switch {
	case strings.Contains(n, "JFL"):
		return "JFL"
	case strings.Contains(n, "INTELBRAS"):
		return "INTELBRAS"
	case strings.Contains(n, "VETTI"):
		return "VETTI"
	case strings.Contains(n, "COMPATEC"), strings.Contains(n, "CONTINENTE"):
		return "COMPATEC"
	default:
		return n
	}
}

func ListarFabricantesPortas() ([]FabricantePorta, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	portas, err := listarPortasMaster()
	if err != nil {
		return nil, err
	}
	portaPorModulo := map[string]string{}
	for _, p := range portas {
		mod := strings.ToUpper(strings.TrimSpace(p.Modulo))
		if mod != "" && p.Porta != "" {
			if _, ok := portaPorModulo[mod]; !ok {
				portaPorModulo[mod] = p.Porta
			}
		}
	}
	rows, err := db.Conn.Query(`
SELECT fabricantes.ID_Fabricante, fabricantes.Nome
FROM fabricantes
LEFT JOIN listaBloqueio ON fabricantes.ID_Fabricante = listaBloqueio.ID_Alvo
WHERE listaBloqueio.ID_Alvo IS NULL
ORDER BY fabricantes.Nome`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FabricantePorta
	for rows.Next() {
		var f FabricantePorta
		if err := rows.Scan(&f.IDFabricante, &f.Nome); err != nil {
			return nil, err
		}
		mod := ModuloForFabricante(f.Nome)
		porta := portaPorModulo[mod]
		if porta == "" {
			continue
		}
		f.Modulo = mod
		f.Porta = porta
		out = append(out, f)
	}
	return out, rows.Err()
}

type portaRow struct {
	Modulo string
	Porta  string
}

func listarPortasMaster() ([]portaRow, error) {
	rows, err := db.Conn.Query(`
SELECT receptorEvento.Modulo, receptorEvento.Porta
FROM receptorEvento
WHERE receptorEvento.Ativo = 'S'
  AND receptorEvento.Producao = 'S'
  AND receptorEvento.Master = 'S'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []portaRow
	for rows.Next() {
		var p portaRow
		if err := rows.Scan(&p.Modulo, &p.Porta); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func PortaModulo(modulo string) (string, error) {
	portas, err := listarPortasMaster()
	if err != nil {
		return "", err
	}
	mod := strings.ToUpper(strings.TrimSpace(modulo))
	for _, p := range portas {
		if strings.ToUpper(strings.TrimSpace(p.Modulo)) == mod {
			return strings.TrimSpace(p.Porta), nil
		}
	}
	return "", fmt.Errorf("porta nao encontrada para modulo %s", modulo)
}

const (
	diagErroConexaoLimite   = 25  // worker / verificar (1a pagina)
	diagErroConexaoPageSize = 50  // lab scroll
	diagErroConexaoMaxPage  = 200 // teto por requisicao
)

func fabricanteSQLFilter(fab string) (string, []any) {
	fab = strings.ToUpper(strings.TrimSpace(fab))
	if fab == "COMPATEC" {
		return ` AND (UPPER(f.Nome) LIKE '%COMPATEC%' OR UPPER(f.Nome) LIKE '%CONTINENTE%')`, nil
	}
	return ` AND UPPER(f.Nome) LIKE ?`, []any{"%" + fab + "%"}
}

func scanDispositivo(row *sql.Row) (*DispositivoInfo, error) {
	var d DispositivoInfo
	var idF1, idF2, cod sql.NullString
	var dt sql.NullTime
	err := row.Scan(
		&d.IDDispositivo, &d.Conta, &idF1, &idF2,
		&d.IDCliente, &d.NomeCliente,
		&d.IDFranqueado, &d.NomeFranqueado,
		&dt, &cod,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.IdFisico1 = idF1.String
	d.IdFisico2 = idF2.String
	if dt.Valid {
		d.DataUltimo = dt.Time.Format("02/01/2006 15:04:05")
	}
	d.CodigoUltimo = cod.String
	return &d, nil
}

func BuscarDispositivo(idFra, fabricante, conta string) (*DispositivoInfo, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	conta = strings.TrimSpace(conta)
	if conta == "" {
		return nil, nil
	}
	fab := strings.ToUpper(strings.TrimSpace(fabricante))
	q := `
SELECT d.ID_Dispositivo, d.Conta, d.IdFisico1, d.IdFisico2,
       c.ID_Cliente, c.Nome,
       c.ID_Franqueado, COALESCE(NULLIF(fr.NomeFantasia,''), NULLIF(fr.RazaoSocial,''), ''),
       d.DataUltimoEvento, d.CodgioUltimoEvento
FROM dispositivo d
JOIN cliente c ON d.ID_Cliente = c.ID_Cliente
JOIN franqueado fr ON c.ID_Franqueado = fr.ID_Franqueado
JOIN fabricantes f ON d.ID_Fabricante = f.ID_Fabricante
WHERE d.Conta = ?`
	args := []any{conta}
	frag, fargs := fabricanteSQLFilter(fab)
	q += frag
	args = append(args, fargs...)
	if strings.TrimSpace(idFra) != "" {
		q += ` AND c.ID_Franqueado = ?`
		args = append(args, strings.TrimSpace(idFra))
	}
	q += ` LIMIT 1`
	return scanDispositivo(db.Conn.QueryRow(q, args...))
}

func BuscarDispositivoPorImei(idFra, fabricante, codigo string) (*DispositivoInfo, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	codigo = strings.TrimSpace(codigo)
	if codigo == "" {
		return nil, nil
	}
	fab := strings.ToUpper(strings.TrimSpace(fabricante))
	q := `
SELECT d.ID_Dispositivo, d.Conta, d.IdFisico1, d.IdFisico2,
       c.ID_Cliente, c.Nome,
       c.ID_Franqueado, COALESCE(NULLIF(fr.NomeFantasia,''), NULLIF(fr.RazaoSocial,''), ''),
       d.DataUltimoEvento, d.CodgioUltimoEvento
FROM dispositivo d
JOIN cliente c ON d.ID_Cliente = c.ID_Cliente
JOIN franqueado fr ON c.ID_Franqueado = fr.ID_Franqueado
JOIN fabricantes f ON d.ID_Fabricante = f.ID_Fabricante
WHERE (d.IdFisico1 = ? OR d.IdFisico2 = ? OR d.IdFisico1 LIKE ? OR d.IdFisico2 LIKE ?)`
	args := []any{codigo, codigo, "%" + codigo + "%", "%" + codigo + "%"}
	frag, fargs := fabricanteSQLFilter(fab)
	q += frag
	args = append(args, fargs...)
	if strings.TrimSpace(idFra) != "" {
		q += ` AND c.ID_Franqueado = ?`
		args = append(args, strings.TrimSpace(idFra))
	}
	q += ` ORDER BY d.DataUltimoEvento DESC LIMIT 1`
	return scanDispositivo(db.Conn.QueryRow(q, args...))
}

func ListarFranqueados(limite int) ([]FranqueadoRow, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	if limite <= 0 || limite > 500 {
		limite = 200
	}
	rows, err := db.Conn.Query(`
SELECT f.ID_Franqueado, COALESCE(NULLIF(f.NomeFantasia,''), NULLIF(f.RazaoSocial,''), f.ID_Franqueado)
FROM franqueado f
WHERE (f.DataCancelamento IS NULL OR f.DataCancelamento = '0000-00-00 00:00:00')
ORDER BY 2
LIMIT ?`, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FranqueadoRow
	for rows.Next() {
		var f FranqueadoRow
		if err := rows.Scan(&f.IDFranqueado, &f.Nome); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func normalizarConta(conta string) string {
	conta = strings.TrimSpace(conta)
	if conta == "" {
		return ""
	}
	for len(conta) < 4 {
		conta = "0" + conta
	}
	if len(conta) > 4 {
		conta = conta[len(conta)-4:]
	}
	return conta
}

func fabricanteErroConexaoFilter(fab string) (string, []any) {
	fab = ModuloForFabricante(fab)
	if fab == "" {
		return "", nil
	}
	if fab == "COMPATEC" {
		return ` AND (UPPER(e.Fabricante) LIKE '%COMPATEC%' OR UPPER(e.Fabricante) LIKE '%CONTINENTE%')`, nil
	}
	if fab == "JFL" {
		return ` AND UPPER(e.Fabricante) LIKE '%JFL%'`, nil
	}
	return ` AND UPPER(e.Fabricante) LIKE ?`, []any{"%" + fab + "%"}
}

func ListarDispositivosPorConta(idFra, fabricante, conta string, limite int) ([]DispositivoInfo, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	conta = normalizarConta(conta)
	if conta == "" {
		return nil, nil
	}
	if limite <= 0 || limite > 20 {
		limite = 10
	}
	fab := strings.ToUpper(strings.TrimSpace(fabricante))
	q := `
SELECT d.ID_Dispositivo, d.Conta, d.IdFisico1, d.IdFisico2,
       c.ID_Cliente, c.Nome,
       c.ID_Franqueado, COALESCE(NULLIF(fr.NomeFantasia,''), NULLIF(fr.RazaoSocial,''), ''),
       d.DataUltimoEvento, d.CodgioUltimoEvento
FROM dispositivo d
JOIN cliente c ON d.ID_Cliente = c.ID_Cliente
JOIN franqueado fr ON c.ID_Franqueado = fr.ID_Franqueado
JOIN fabricantes f ON d.ID_Fabricante = f.ID_Fabricante
WHERE d.Conta = ?`
	args := []any{conta}
	frag, fargs := fabricanteSQLFilter(fab)
	q += frag
	args = append(args, fargs...)
	if strings.TrimSpace(idFra) != "" {
		q += ` AND c.ID_Franqueado = ?`
		args = append(args, strings.TrimSpace(idFra))
	}
	q += ` ORDER BY d.DataUltimoEvento DESC LIMIT ?`
	args = append(args, limite)

	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DispositivoInfo
	for rows.Next() {
		var idF1, idF2 sql.NullString
		var d DispositivoInfo
		var dt sql.NullTime
		var cod sql.NullString
		if err := rows.Scan(
			&d.IDDispositivo, &d.Conta, &idF1, &idF2,
			&d.IDCliente, &d.NomeCliente,
			&d.IDFranqueado, &d.NomeFranqueado,
			&dt, &cod,
		); err != nil {
			return nil, err
		}
		d.IdFisico1 = idF1.String
		d.IdFisico2 = idF2.String
		if dt.Valid {
			d.DataUltimo = dt.Time.Format("02/01/2006 15:04:05")
		}
		d.CodigoUltimo = cod.String
		out = append(out, d)
	}
	return out, rows.Err()
}

func ListarErroConexao(idFra, fabricante, conta, codigo string, limite, offset int) ([]ErroConexaoRow, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	if limite <= 0 {
		limite = diagErroConexaoPageSize
	}
	if limite > diagErroConexaoMaxPage {
		limite = diagErroConexaoMaxPage
	}
	if offset < 0 {
		offset = 0
	}
	codigo = normalizarCodigoFisico(codigo)
	fabricante = ModuloForFabricante(fabricante)
	_ = conta  // highlight via EnrichErroConexaoMatch, nao filtra SQL
	_ = codigo // highlight via EnrichErroConexaoMatch, nao filtra SQL

	q := `
SELECT e.ID_ErroConexao, e.Codigo, e.Serial, e.Imei, e.IP_Remoto,
       e.Modelo, e.Versao, e.Protocolo, e.Fabricante, e.Conta, e.ID_Franqueado, e.DataTentativa
FROM erroConexao e
WHERE 1=1`
	args := []any{}

	frag, fargs := fabricanteErroConexaoFilter(fabricante)
	q += frag
	args = append(args, fargs...)

	// conta e MAC: nao filtra na SQL — lista recentes e destaca via EnrichErroConexaoMatch
	if strings.TrimSpace(idFra) != "" {
		q += ` AND (
			e.ID_Franqueado = ?
			OR (
				COALESCE(e.ID_Franqueado, '') = ''
				AND e.Conta IN (
					SELECT d.Conta FROM dispositivo d
					JOIN cliente c ON d.ID_Cliente = c.ID_Cliente
					WHERE c.ID_Franqueado = ?
				)
			)
		)`
		args = append(args, strings.TrimSpace(idFra), strings.TrimSpace(idFra))
	}
	q += ` AND DATE(e.DataTentativa) = CURDATE()`
	q += ` ORDER BY e.DataTentativa DESC LIMIT ? OFFSET ?`
	args = append(args, limite, offset)

	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ErroConexaoRow
	for rows.Next() {
		var e ErroConexaoRow
		var dt sql.NullTime
		if err := rows.Scan(
			&e.ID, &e.Codigo, &e.Serial, &e.Imei, &e.IPRemoto,
			&e.Modelo, &e.Versao, &e.Protocolo, &e.Fabricante, &e.Conta, &e.IDFranqueado, &dt,
		); err != nil {
			return nil, err
		}
		if dt.Valid {
			e.DataTentativa = dt.Time.Format("02/01/2006 15:04:05")
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func ListarErroConexaoLab(fabricante string, limite, offset int) ([]ErroConexaoRow, error) {
	return ListarErroConexao("", fabricante, "", "", limite, offset)
}

// ContagemErroConexaoHoje retorna (total hoje, total do fabricante hoje) para diagnostico lab.
func ContagemErroConexaoHoje(fabricante string) (int, int, error) {
	if db.Conn == nil {
		return 0, 0, errors.New("mysql indisponivel")
	}
	var total int
	if err := db.Conn.QueryRow(`
SELECT COUNT(*) FROM erroConexao e WHERE DATE(e.DataTentativa) = CURDATE()`).Scan(&total); err != nil {
		return 0, 0, err
	}
	fabricante = ModuloForFabricante(fabricante)
	if fabricante == "" {
		return total, 0, nil
	}
	q := `SELECT COUNT(*) FROM erroConexao e WHERE DATE(e.DataTentativa) = CURDATE()`
	args := []any{}
	frag, fargs := fabricanteErroConexaoFilter(fabricante)
	q += frag
	args = append(args, fargs...)
	var porFab int
	if err := db.Conn.QueryRow(q, args...).Scan(&porFab); err != nil {
		return total, 0, err
	}
	return total, porFab, nil
}
