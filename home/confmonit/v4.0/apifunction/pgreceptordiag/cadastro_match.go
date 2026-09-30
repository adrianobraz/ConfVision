package pgreceptordiag

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"apifunction/db"
)

type cadastroMatch struct {
	Tipo           string
	Distancia      int
	CadastroCodigo string
	Observacao     string
}

type SimularCadastroResult struct {
	Status          string           `json:"status"`
	Mensagem        string           `json:"mensagem"`
	CodigoInformado string           `json:"codigoInformado"`
	CodigoCadastro  string           `json:"codastroCodigo,omitempty"`
	Distancia       int              `json:"distancia,omitempty"`
	Dispositivo     *DispositivoInfo `json:"dispositivo,omitempty"`
}

const idFisicoWireMinLen = 4

func normalizarCodigoFisico(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func hammingIgualLen(a, b string) int {
	if len(a) != len(b) || a == "" {
		return -1
	}
	d := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			d++
		}
	}
	return d
}

func melhorSimilaridade(recebido, id1, id2 string) (melhor string, distancia int) {
	distancia = -1
	for _, raw := range []string{id1, id2} {
		c := normalizarCodigoFisico(raw)
		if c == "" {
			continue
		}
		if recebido == c {
			return c, 0
		}
		if strings.Contains(c, recebido) || strings.Contains(recebido, c) {
			if distancia < 0 {
				melhor = c
				distancia = 0
			}
			continue
		}
		d := hammingIgualLen(recebido, c)
		if d < 0 {
			n := len(recebido)
			if len(c) < n {
				n = len(c)
			}
			if n < idFisicoWireMinLen {
				n = idFisicoWireMinLen
			}
			if len(recebido) >= n && len(c) >= n {
				sufR := recebido[len(recebido)-n:]
				sufC := c[len(c)-n:]
				d = hammingIgualLen(sufR, sufC)
				if d >= 0 {
					c = sufC
				}
			}
		}
		if d >= 0 && d <= 2 && (distancia < 0 || d < distancia) {
			distancia = d
			melhor = c
		}
	}
	return melhor, distancia
}

func BuscarDispositivoExato(idFra, fabricante, conta, codigo string) (*DispositivoInfo, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	conta = normalizarConta(conta)
	codigo = normalizarCodigoFisico(codigo)
	if conta == "" || codigo == "" {
		return nil, nil
	}
	fab := strings.ToUpper(strings.TrimSpace(fabricante))
	n1 := `UPPER(REPLACE(REPLACE(REPLACE(IFNULL(d.IdFisico1,''), ':', ''), '-', ''), '.', ''))`
	n2 := `UPPER(REPLACE(REPLACE(REPLACE(IFNULL(d.IdFisico2,''), ':', ''), '-', ''), '.', ''))`
	wl := len(codigo)
	q := `
SELECT d.ID_Dispositivo, d.Conta, d.IdFisico1, d.IdFisico2,
       c.ID_Cliente, c.Nome,
       c.ID_Franqueado, COALESCE(NULLIF(fr.NomeFantasia,''), NULLIF(fr.RazaoSocial,''), ''),
       d.DataUltimoEvento, d.CodgioUltimoEvento
FROM dispositivo d
JOIN cliente c ON d.ID_Cliente = c.ID_Cliente
JOIN franqueado fr ON c.ID_Franqueado = fr.ID_Franqueado
JOIN fabricantes f ON d.ID_Fabricante = f.ID_Fabricante
WHERE d.Conta = ?
AND (
  ` + n1 + ` = ?
  OR RIGHT(` + n1 + `, ?) = ?
  OR ` + n2 + ` = ?
  OR RIGHT(` + n2 + `, ?) = ?
)`
	args := []any{conta, codigo, wl, codigo, codigo, wl, codigo}
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

func avaliarMatchCadastro(idFra, fabricante, conta, codigo string) cadastroMatch {
	recv := normalizarCodigoFisico(codigo)
	conta = normalizarConta(conta)
	if recv == "" || conta == "" {
		return cadastroMatch{Tipo: "none"}
	}

	if disp, _ := BuscarDispositivoExato(idFra, fabricante, conta, recv); disp != nil {
		return cadastroMatch{
			Tipo:           "exact",
			CadastroCodigo: pickIdFisico(disp.IdFisico1, disp.IdFisico2),
			Observacao:     "Confere com o cadastro",
		}
	}

	dispConta, _ := BuscarDispositivo(idFra, fabricante, conta)
	if dispConta != nil {
		melhor, dist := melhorSimilaridade(recv,
			dispConta.IdFisico1, dispConta.IdFisico2)
		if dist > 0 && dist <= 2 {
			return cadastroMatch{
				Tipo:           "similar",
				Distancia:      dist,
				CadastroCodigo: melhor,
				Observacao: fmt.Sprintf(
					"Possivel typo no cadastro: central enviou %s, cadastrado %s",
					recv, melhor,
				),
			}
		}
		cad := pickIdFisico(dispConta.IdFisico1, dispConta.IdFisico2)
		if cad != "" {
			return cadastroMatch{
				Tipo:           "mismatch",
				CadastroCodigo: cad,
				Observacao: fmt.Sprintf(
					"Conta ok, cadastro tem %s, central enviou %s",
					cad, recv,
				),
			}
		}
	}

	if dispCod, _ := BuscarDispositivoPorImei(idFra, fabricante, recv); dispCod != nil {
		return cadastroMatch{
			Tipo:           "outra_conta",
			CadastroCodigo: pickIdFisico(dispCod.IdFisico1, dispCod.IdFisico2),
			Observacao: fmt.Sprintf(
				"ID cadastrado na conta %s, nao em %s",
				dispCod.Conta, conta,
			),
		}
	}

	return cadastroMatch{Tipo: "none", Observacao: "Sem cadastro compativel"}
}

func EnrichErroConexaoMatch(idFra, fabricante, contaDigitada, codigoDigitado string, rows []ErroConexaoRow) []ErroConexaoRow {
	out := make([]ErroConexaoRow, len(rows))
	copy(out, rows)
	contaDig := normalizarConta(contaDigitada)
	codDig := normalizarCodigoFisico(codigoDigitado)
	for i := range out {
		if contaDig != "" && normalizarConta(out[i].Conta) == contaDig {
			out[i].ContaMatch = true
		}
		recv := normalizarCodigoFisico(out[i].Codigo)
		if codDig != "" && recv != "" {
			out[i].CodigoMatch = classificarCodigoMatch(recv, codDig)
		}
		m := avaliarMatchCadastro(idFra, fabricante, out[i].Conta, out[i].Codigo)
		if out[i].MatchObservacao == "" {
			out[i].MatchTipo = m.Tipo
			out[i].MatchDistancia = m.Distancia
			out[i].CadastroCodigo = m.CadastroCodigo
			out[i].MatchObservacao = m.Observacao
		}
		if codDig != "" && recv != "" && out[i].CodigoMatch == "similar" {
			_, dist := melhorSimilaridade(recv, codDig, codDig)
			if dist > 0 {
				out[i].MatchTipo = "similar"
				out[i].MatchDistancia = dist
				out[i].MatchObservacao = fmt.Sprintf(
					"MAC real da central: %s — voce digitou %s (%d caractere(s) diferente(s))",
					recv, codDig, dist,
				)
			}
		}
	}
	return out
}

func classificarCodigoMatch(recebido, digitado string) string {
	if recebido == digitado {
		return "exact"
	}
	if d := hammingIgualLen(recebido, digitado); d >= 0 && d <= 2 {
		return "similar"
	}
	n := len(recebido)
	if len(digitado) < n {
		n = len(digitado)
	}
	if n < idFisicoWireMinLen {
		n = idFisicoWireMinLen
	}
	if len(recebido) >= n && len(digitado) >= n {
		if recebido[len(recebido)-n:] == digitado[len(digitado)-n:] {
			return "exact"
		}
		if d := hammingIgualLen(recebido[len(recebido)-n:], digitado[len(digitado)-n:]); d >= 0 && d <= 2 {
			return "similar"
		}
	}
	return "different"
}

// EnrichErroConexaoMatch legacy wrapper
func enrichErroConexaoRowsLegacy(idFra, fabricante string, rows []ErroConexaoRow) []ErroConexaoRow {
	return EnrichErroConexaoMatch(idFra, fabricante, "", "", rows)
}

func SimularCadastro(idFra, fabricante, conta, codigo string) (*SimularCadastroResult, error) {
	conta = normalizarConta(conta)
	codigo = normalizarCodigoFisico(codigo)
	if conta == "" {
		return nil, fmt.Errorf("conta obrigatoria")
	}
	if codigo == "" {
		return nil, fmt.Errorf("codigo ou MAC obrigatorio")
	}
	if len(codigo) < 4 {
		return nil, fmt.Errorf("codigo muito curto (minimo 4 caracteres)")
	}

	res := &SimularCadastroResult{CodigoInformado: codigo}

	if disp, err := BuscarDispositivoExato(idFra, fabricante, conta, codigo); err != nil {
		return nil, err
	} else if disp != nil {
		res.Status = "ok"
		res.Mensagem = "Cadastro confere — conta e ID fisico batem com a consulta do receptor"
		res.Dispositivo = disp
		res.CodigoCadastro = pickIdFisico(disp.IdFisico1, disp.IdFisico2)
		return res, nil
	}

	dispConta, err := BuscarDispositivo(idFra, fabricante, conta)
	if err != nil {
		return nil, err
	}
	if dispConta != nil {
		melhor, dist := melhorSimilaridade(codigo, dispConta.IdFisico1, dispConta.IdFisico2)
		if dist > 0 && dist <= 2 {
			res.Status = "similar"
			res.Distancia = dist
			res.CodigoCadastro = melhor
			res.Dispositivo = dispConta
			res.Mensagem = fmt.Sprintf(
				"Possivel erro de digitacao — cadastrado %s, informado %s (%d caractere(s) diferente(s))",
				melhor, codigo, dist,
			)
			return res, nil
		}
		cad := pickIdFisico(dispConta.IdFisico1, dispConta.IdFisico2)
		res.Status = "erro"
		res.CodigoCadastro = cad
		res.Dispositivo = dispConta
		res.Mensagem = fmt.Sprintf(
			"Conta %s existe com ID %s, mas voce informou %s — o receptor rejeitaria a conexao",
			conta, cad, codigo,
		)
		return res, nil
	}

	if dispCod, _ := BuscarDispositivoPorImei(idFra, fabricante, codigo); dispCod != nil {
		res.Status = "erro"
		res.Dispositivo = dispCod
		res.CodigoCadastro = pickIdFisico(dispCod.IdFisico1, dispCod.IdFisico2)
		res.Mensagem = fmt.Sprintf(
			"ID %s esta cadastrado na conta %s, nao na conta %s informada",
			codigo, dispCod.Conta, conta,
		)
		return res, nil
	}

	res.Status = "erro"
	res.Mensagem = "Nenhum dispositivo encontrado com esta conta e ID fisico no cadastro"
	return res, nil
}

var _ = sql.ErrNoRows
