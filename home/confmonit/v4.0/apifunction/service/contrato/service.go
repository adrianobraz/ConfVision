package contrato

import (
	"apifunction/auth"
	"apifunction/db"
	"apifunction/pgcredito"
	"apifunction/pggovernanca"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

func Salvar(sess auth.SessaoAdm, in SalvarInput) (ContratoResumo, error) {
	idFra := strings.TrimSpace(in.IDFranqueado)
	if err := AssertPodeEditarContrato(sess, idFra); err != nil {
		return ContratoResumo{}, err
	}
	if strings.TrimSpace(in.Nome) == "" {
		return ContratoResumo{}, errors.New("nome do contrato obrigatorio")
	}
	idCentral := sess.IDCentralFinanceiro()
	if idCentral == "" {
		return ContratoResumo{}, errors.New("id_central obrigatorio (IDCentralUUID)")
	}
	idRep, _, _ := franqueadoRepresentante(idFra)
	idCentral = centralDoFranqueado(idFra, idCentral)
	if idCentral == "" {
		return ContratoResumo{}, errors.New("nao foi possivel resolver central do franqueado")
	}

	preview := CalcularPreview(in.Itens, in.DescontoTipo, in.DescontoValor)
	if err := ValidarDesconto(preview.ValorBase, in.DescontoTipo, in.DescontoValor); err != nil {
		return ContratoResumo{}, err
	}
	inicio, err := parseDate(in.InicioEm)
	if err != nil {
		return ContratoResumo{}, errors.New("inicio_em invalido")
	}
	period := strings.TrimSpace(in.Periodicidade)
	if period == "" {
		period = "mensal"
	}

	modJSON, _ := json.Marshal(preview.Modulos)
	limJSON, _ := json.Marshal(preview.Limites)

	tx, err := db.Conn.Begin()
	if err != nil {
		return ContratoResumo{}, err
	}
	defer tx.Rollback()

	var faturasCancelar []faturaContabilRef
	if in.IDContrato > 0 {
		var statusAtual string
		if err := tx.QueryRow(`SELECT Status FROM fp_contrato WHERE ID_Contrato = ?`, in.IDContrato).Scan(&statusAtual); err == nil {
			if strings.ToLower(strings.TrimSpace(statusAtual)) == "aguardando_pagamento" {
				faturasCancelar, _ = listarFaturasAbertasContabil(in.IDContrato, tx)
			}
		}
	}

	status := "rascunho"
	if in.GerarFatura {
		status = "aguardando_pagamento"
	}

	var idContrato int
	if in.IDContrato > 0 {
		idContrato, err = atualizarContratoTx(tx, sess, in, preview, idFra, idCentral, idRep, inicio, period, status, modJSON, limJSON)
	} else {
		if err := assertNovoContratoPermitido(tx, idFra); err != nil {
			return ContratoResumo{}, err
		}
		idContrato, err = inserirContratoTx(tx, sess, in, preview, idFra, idCentral, idRep, inicio, period, status, modJSON, limJSON)
	}
	if err != nil {
		return ContratoResumo{}, err
	}

	if err = inserirItensContratoTx(tx, idContrato, preview.Itens); err != nil {
		return ContratoResumo{}, err
	}

	var faturaID int
	if in.GerarFatura {
		venc := inicio.AddDate(0, 0, 7)
		if v := strings.TrimSpace(in.VencimentoEm); v != "" {
			if t, errV := parseDate(v); errV == nil {
				venc = t
			}
		}
		valorCobrar := ValorPorPeriodicidade(preview.ValorFinal, period)
		faturaID, err = gerarFaturaTx(tx, idContrato, idFra, idCentral, idRep, "contrato_inicial", valorCobrar, venc)
		if err != nil {
			return ContratoResumo{}, err
		}
		_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Fatura, ID_Franqueado, Acao, Detalhe, ID_Usuario) VALUES (?,?,?,?,?,?)`,
			idContrato, faturaID, idFra, "fatura_inicial", fmt.Sprintf("fatura_id=%d", faturaID), sess.IDUsuario)
	}

	acaoLog := "contrato_criado"
	if in.IDContrato > 0 {
		acaoLog = "contrato_atualizado"
	}
	_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Franqueado, Acao, Detalhe, ID_Usuario) VALUES (?,?,?,?,?)`,
		idContrato, idFra, acaoLog, in.Nome, sess.IDUsuario)

	if err = tx.Commit(); err != nil {
		return ContratoResumo{}, err
	}

	out := ContratoResumo{
		IDContrato: idContrato, IDFranqueado: idFra, Nome: in.Nome, Status: status,
		ValorFinal: preview.ValorFinal, Periodicidade: period,
		InicioEm: inicio.Format("2006-01-02"), Itens: preview.Itens,
	}

	if len(faturasCancelar) > 0 {
		cancelarFaturasContabilidade(faturasCancelar, "Cancelada — contrato reeditado", sess)
	}
	if faturaID > 0 {
		if err := espelharFaturaContabil(idContrato, faturaID, in.Nome, sess); err != nil {
			return out, erroContabilidade(err)
		}
	}

	return out, nil
}

func inserirContratoTx(tx *sql.Tx, sess auth.SessaoAdm, in SalvarInput, preview PreviewResult,
	idFra, idCentral, idRep string, inicio time.Time, period, status string, modJSON, limJSON []byte) (int, error) {
	res, err := tx.Exec(`
		INSERT INTO fp_contrato (
			ID_Franqueado, ID_Central, ID_Representante, Nome, Status,
			ValorBase, DescontoTipo, DescontoValor, ValorFinal,
			Periodicidade, InicioEm, ProximoReajusteEm, ProximaCobrancaEm, VencimentoDia,
			PermiteExcedente, ValorUnitarioCamera, Observacao, CriadoPor,
			ModulosJSON, LimitesJSON
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`, idFra, idCentral, nullStr(idRep), in.Nome, status,
		preview.ValorBase, nullStr(in.DescontoTipo), in.DescontoValor, preview.ValorFinal,
		period, inicio, nullDate(in.ProximoReajuste), nullTime(proximaCobranca(inicio, period)),
		nullInt(in.VencimentoDia), coalesceSN(in.PermiteExcedente, "S"), nullFloat(in.ValorUnitCamera),
		nullStr(in.Observacao), sess.IDUsuario, modJSON, limJSON)
	if err != nil {
		return 0, err
	}
	idContrato64, _ := res.LastInsertId()
	return int(idContrato64), nil
}

func atualizarContratoTx(tx *sql.Tx, sess auth.SessaoAdm, in SalvarInput, preview PreviewResult,
	idFra, idCentral, idRep string, inicio time.Time, period, status string, modJSON, limJSON []byte) (int, error) {
	var statusAtual string
	var idFraAtual string
	err := tx.QueryRow(`
		SELECT Status, ID_Franqueado FROM fp_contrato WHERE ID_Contrato = ? FOR UPDATE
	`, in.IDContrato).Scan(&statusAtual, &idFraAtual)
	if err == sql.ErrNoRows {
		return 0, errors.New("contrato nao encontrado")
	}
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(idFraAtual) != idFra {
		return 0, errors.New("franqueado do contrato nao confere")
	}
	if err := AssertPodeEditarContrato(sess, idFra); err != nil {
		return 0, err
	}

	stAtual := strings.ToLower(strings.TrimSpace(statusAtual))
	if err := assertContratoEditavel(stAtual, in.IDContrato, tx); err != nil {
		return 0, err
	}

	if stAtual == "aguardando_pagamento" {
		if _, err = cancelarFaturasAbertasTx(tx, in.IDContrato); err != nil {
			return 0, err
		}
	}

	_, err = tx.Exec(`
		UPDATE fp_contrato SET
			ID_Central = ?, ID_Representante = ?, Nome = ?, Status = ?,
			ValorBase = ?, DescontoTipo = ?, DescontoValor = ?, ValorFinal = ?,
			Periodicidade = ?, InicioEm = ?, ProximoReajusteEm = ?, ProximaCobrancaEm = ?,
			VencimentoDia = ?, PermiteExcedente = ?, ValorUnitarioCamera = ?, Observacao = ?,
			ModulosJSON = ?, LimitesJSON = ?, UpdatedAt = NOW()
		WHERE ID_Contrato = ?
	`, idCentral, nullStr(idRep), in.Nome, status,
		preview.ValorBase, nullStr(in.DescontoTipo), in.DescontoValor, preview.ValorFinal,
		period, inicio, nullDate(in.ProximoReajuste), nullTime(proximaCobranca(inicio, period)),
		nullInt(in.VencimentoDia), coalesceSN(in.PermiteExcedente, "S"), nullFloat(in.ValorUnitCamera),
		nullStr(in.Observacao), modJSON, limJSON, in.IDContrato)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`DELETE FROM fp_contrato_item WHERE ID_Contrato = ?`, in.IDContrato)
	if err != nil {
		return 0, err
	}
	return in.IDContrato, nil
}

func inserirItensContratoTx(tx *sql.Tx, idContrato int, itens []ItemCalculado) error {
	for _, it := range itens {
		_, err := tx.Exec(`
			INSERT INTO fp_contrato_item
			(ID_Contrato, Tipo, Chave, Descricao, Quantidade, QtdPorPacote, ValorUnitario, ValorTotal, RefID)
			VALUES (?,?,?,?,?,?,?,?,?)
		`, idContrato, it.Tipo, it.Chave, it.Descricao, it.Quantidade, nullInt(it.QtdPorPacote),
			it.ValorUnitario, it.ValorTotal, nullInt(it.RefID))
		if err != nil {
			return err
		}
	}
	return nil
}

func Obter(sess auth.SessaoAdm, idContrato int) (ContratoDetalhe, error) {
	if idContrato <= 0 {
		return ContratoDetalhe{}, errors.New("id_contrato invalido")
	}

	var d ContratoDetalhe
	var prox, reaj sql.NullString
	var descTipo sql.NullString
	var descVal sql.NullFloat64
	var idRep sql.NullString
	var nomeFra, nomeRep sql.NullString

	err := db.Conn.QueryRow(`
		SELECT
			c.ID_Contrato, c.ID_Franqueado, c.Nome, c.Status, c.ValorFinal, c.Periodicidade,
			DATE_FORMAT(c.InicioEm,'%Y-%m-%d'), DATE_FORMAT(c.ProximaCobrancaEm,'%Y-%m-%d'),
			c.DescontoTipo, COALESCE(c.DescontoValor, 0),
			DATE_FORMAT(c.ProximoReajusteEm,'%Y-%m-%d'),
			c.ID_Representante,
			COALESCE(NULLIF(f.NomeFantasia,''), NULLIF(f.RazaoSocial,''), ''),
			COALESCE(NULLIF(r.NomeFantasia,''), NULLIF(r.RazaoSocial,''), '')
		FROM fp_contrato c
		LEFT JOIN franqueado f ON f.ID_Franqueado = c.ID_Franqueado
		LEFT JOIN representante r ON r.ID_Representante = c.ID_Representante
		WHERE c.ID_Contrato = ?
		LIMIT 1
	`, idContrato).Scan(
		&d.IDContrato, &d.IDFranqueado, &d.Nome, &d.Status, &d.ValorFinal, &d.Periodicidade,
		&d.InicioEm, &prox, &descTipo, &descVal, &reaj, &idRep, &nomeFra, &nomeRep,
	)
	if err == sql.ErrNoRows {
		return ContratoDetalhe{}, errors.New("contrato nao encontrado")
	}
	if err != nil {
		return ContratoDetalhe{}, err
	}

	if err := AssertPodeEditarContrato(sess, d.IDFranqueado); err != nil {
		return ContratoDetalhe{}, err
	}

	if prox.Valid {
		d.ProximaCobranca = prox.String
	}
	if descTipo.Valid {
		d.DescontoTipo = strings.TrimSpace(descTipo.String)
	}
	d.DescontoValor = descVal.Float64
	if reaj.Valid {
		d.ProximoReajusteEm = reaj.String
	}
	if idRep.Valid {
		d.IDRepresentante = strings.TrimSpace(idRep.String)
	}
	if nomeFra.Valid {
		d.NomeFranqueado = strings.TrimSpace(nomeFra.String)
	}
	if nomeRep.Valid {
		d.NomeRepresentante = strings.TrimSpace(nomeRep.String)
	}

	rows, err := db.Conn.Query(`
		SELECT Tipo, Chave, Descricao, Quantidade, QtdPorPacote, ValorUnitario, ValorTotal, RefID
		FROM fp_contrato_item WHERE ID_Contrato = ? ORDER BY ID_ContratoItem
	`, idContrato)
	if err != nil {
		return ContratoDetalhe{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var it ItemCalculado
		var qtdPac sql.NullInt64
		var refID sql.NullInt64
		if err := rows.Scan(&it.Tipo, &it.Chave, &it.Descricao, &it.Quantidade, &qtdPac, &it.ValorUnitario, &it.ValorTotal, &refID); err != nil {
			return ContratoDetalhe{}, err
		}
		if qtdPac.Valid {
			it.QtdPorPacote = int(qtdPac.Int64)
		}
		if refID.Valid {
			it.RefID = int(refID.Int64)
		}
		d.Itens = append(d.Itens, it)
	}
	editavel, err := ContratoEditavel(d.Status, idContrato, db.Conn)
	if err != nil {
		return ContratoDetalhe{}, err
	}
	d.Editavel = editavel
	return d, rows.Err()
}

func gerarFaturaTx(tx *sql.Tx, idContrato int, idFra, idCentral, idRep, tipo string, valor float64, vencimento time.Time) (int, error) {
	ref := time.Now().Format("2006-01")
	ciclo := fmt.Sprintf("%s-C%d", ref, idContrato)
	res, err := tx.Exec(`
		INSERT INTO fp_fatura_contrato
		(ID_Contrato, ID_Franqueado, ID_Central, ID_Representante, Tipo, Referencia, Status, ValorTotal, VencimentoEm, CicloRef)
		VALUES (?,?,?,?,?,?,'aberta',?,?,?)
	`, idContrato, idFra, idCentral, nullStr(idRep), tipo, ref, valor, vencimento, ciclo)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func ConfirmarPagamento(sess auth.SessaoAdm, idFatura int) error {
	return confirmarPagamentoInterno(sess, idFatura, false)
}

// ConfirmarPagamentoContabil ativa contrato/fatura MySQL quando o recebimento
// já foi registrado no Financeiro (fp_fatura Xano paga).
func ConfirmarPagamentoContabil(sess auth.SessaoAdm, idFaturaContabil int) error {
	if idFaturaContabil <= 0 {
		return errors.New("id_fatura_contabil obrigatorio")
	}
	var idFatura int
	err := db.Conn.QueryRow(`
		SELECT ID_Fatura FROM fp_fatura_contrato WHERE ID_FaturaContabil = ?
	`, idFaturaContabil).Scan(&idFatura)
	if err == sql.ErrNoRows {
		return errors.New("fatura contrato nao vinculada ao espelho contabil")
	}
	if err != nil {
		return err
	}
	return confirmarPagamentoInterno(sess, idFatura, true)
}

func confirmarPagamentoInterno(sess auth.SessaoAdm, idFatura int, contabilJaPaga bool) error {
	var idContrato int
	var idFra string
	var status string
	err := db.Conn.QueryRow(`
		SELECT ID_Contrato, ID_Franqueado, Status FROM fp_fatura_contrato WHERE ID_Fatura = ?
	`, idFatura).Scan(&idContrato, &idFra, &status)
	if err != nil {
		return errors.New("fatura nao encontrada")
	}
	if err := AssertPodeEditarContrato(sess, idFra); err != nil {
		return err
	}
	if status == "paga" {
		var contratoStatus string
		_ = db.Conn.QueryRow(`SELECT Status FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&contratoStatus)
		if contratoStatus == "ativo" {
			return errors.New("fatura ja paga")
		}
		return ativarContratoPosPagamento(sess, idContrato, idFra, idFatura)
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.Exec(`UPDATE fp_fatura_contrato SET Status = 'paga', PagoEm = ? WHERE ID_Fatura = ?`, now, idFatura)
	if err != nil {
		return err
	}

	var periodicidade string
	var inicio sql.NullTime
	_ = tx.QueryRow(`SELECT Periodicidade, InicioEm FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&periodicidade, &inicio)

	proxCob := proximaCobranca(now, periodicidade)
	_, err = tx.Exec(`
		UPDATE fp_contrato SET Status = 'ativo', ProximaCobrancaEm = ? WHERE ID_Contrato = ?
	`, proxCob, idContrato)
	if err != nil {
		return err
	}

	// Suspende outros contratos ativos do mesmo franqueado
	_, _ = tx.Exec(`
		UPDATE fp_contrato SET Status = 'cancelado'
		WHERE ID_Franqueado = ? AND ID_Contrato != ? AND Status = 'ativo'
	`, idFra, idContrato)

	if err = aplicarLiberacaoTx(tx, idContrato, idFra); err != nil {
		return err
	}

	_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Fatura, ID_Franqueado, Acao, ID_Usuario) VALUES (?,?,?,?,?)`,
		idContrato, idFatura, idFra, "pagamento_confirmado", sess.IDUsuario)

	if err = tx.Commit(); err != nil {
		return err
	}

	if contabilJaPaga {
		return nil
	}

	var valor float64
	_ = db.Conn.QueryRow(`SELECT ValorTotal FROM fp_fatura_contrato WHERE ID_Fatura = ?`, idFatura).Scan(&valor)
	return erroContabilidade(pagarFaturaContabil(idFatura, valor, sess))
}

func ativarContratoPosPagamento(sess auth.SessaoAdm, idContrato int, idFra string, idFatura int) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	var periodicidade string
	_ = tx.QueryRow(`SELECT Periodicidade FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&periodicidade)
	proxCob := proximaCobranca(now, periodicidade)
	_, err = tx.Exec(`
		UPDATE fp_contrato SET Status = 'ativo', ProximaCobrancaEm = ? WHERE ID_Contrato = ?
	`, proxCob, idContrato)
	if err != nil {
		return err
	}
	_, _ = tx.Exec(`
		UPDATE fp_contrato SET Status = 'cancelado'
		WHERE ID_Franqueado = ? AND ID_Contrato != ? AND Status = 'ativo'
	`, idFra, idContrato)
	if err = aplicarLiberacaoTx(tx, idContrato, idFra); err != nil {
		return err
	}
	_, _ = tx.Exec(`INSERT INTO fp_contrato_log (ID_Contrato, ID_Fatura, ID_Franqueado, Acao, Detalhe, ID_Usuario) VALUES (?,?,?,?,?,?)`,
		idContrato, idFatura, idFra, "contrato_reativado_pos_contabil", "Pagamento ja registrado no Financeiro", sess.IDUsuario)
	return tx.Commit()
}

func aplicarLiberacaoTx(tx *sql.Tx, idContrato int, idFra string) error {
	var modRaw, limRaw []byte
	err := tx.QueryRow(`SELECT ModulosJSON, LimitesJSON FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&modRaw, &limRaw)
	if err != nil {
		return err
	}

	// ConfVision flag
	temCV := false
	rows, err := tx.Query(`SELECT Chave FROM fp_contrato_item WHERE ID_Contrato = ? AND Tipo IN ('produto','cv_licenca')`, idContrato)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ch string
		_ = rows.Scan(&ch)
		if strings.HasPrefix(ch, "confvision") || strings.Contains(ch, "confvision") {
			temCV = true
		}
	}
	usa := "N"
	if temCV {
		usa = "S"
	}
	_, err = tx.Exec(`UPDATE franqueado SET UsaConfVision = ? WHERE ID_Franqueado = ?`, usa, idFra)
	return err
}

func Efetivo(idFranqueado, produto string) EfetivoResult {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if produto == "" {
		produto = "franqueadopro"
	}
	if bloqueado, err := franqueadoBloqueadoGovernanca(idFranqueado); err == nil && bloqueado {
		return EfetivoResult{
			Liberado: false, Motivo: pggovernanca.MotivoPublico, Produto: produto,
			Status: "suspenso", ModulosJSON: map[string]bool{}, LimitesJSON: map[string]int{}, RetencaoDias: 30,
		}
	}
	def := EfetivoResult{
		Liberado: false, Motivo: "sem_contrato", Produto: produto,
		ModulosJSON: map[string]bool{}, LimitesJSON: map[string]int{}, RetencaoDias: 30,
	}

	var idContrato int
	var status string
	var modRaw, limRaw []byte
	var valorFinal sql.NullFloat64
	err := db.Conn.QueryRow(`
		SELECT ID_Contrato, Status, ModulosJSON, LimitesJSON, ValorFinal
		FROM fp_contrato
		WHERE ID_Franqueado = ? AND Status IN ('ativo','aguardando_pagamento','suspenso')
		ORDER BY FIELD(Status,'ativo','aguardando_pagamento','suspenso'), ID_Contrato DESC
		LIMIT 1
	`, idFranqueado).Scan(&idContrato, &status, &modRaw, &limRaw, &valorFinal)
	if err == sql.ErrNoRows {
		return def
	}
	if err != nil {
		def.Motivo = "erro_contrato"
		return def
	}

	def.ContratoID = idContrato
	def.Status = status

	if status == "aguardando_pagamento" {
		def.Motivo = "pendente"
		return def
	}
	if status == "suspenso" {
		def.Motivo = "suspensa"
		return def
	}

	// Verifica fatura vencida aberta
	var faturaVencida int
	_ = db.Conn.QueryRow(`
		SELECT COUNT(*) FROM fp_fatura_contrato
		WHERE ID_Contrato = ? AND Status = 'aberta' AND VencimentoEm < CURDATE()
	`, idContrato).Scan(&faturaVencida)
	if faturaVencida > 0 {
		def.Motivo = "vencida"
		return def
	}

	def.Liberado = true
	def.Motivo = "ok"

	if len(modRaw) > 0 {
		_ = json.Unmarshal(modRaw, &def.ModulosJSON)
	}
	if len(limRaw) > 0 {
		_ = json.Unmarshal(limRaw, &def.LimitesJSON)
	}

	// Plano FP
	var chave string
	_ = db.Conn.QueryRow(`
		SELECT Chave FROM fp_contrato_item
		WHERE ID_Contrato = ? AND Tipo = 'produto' AND Chave LIKE 'franqueadopro|%'
		LIMIT 1
	`, idContrato).Scan(&chave)
	if parts := strings.SplitN(chave, "|", 2); len(parts) == 2 {
		def.Plano = parts[1]
		def.RetencaoDias = retencaoPorPlanoFP(parts[1])
	}

	return def
}

func ListarContratos(sess auth.SessaoAdm, idFranqueado string) ([]ContratoResumo, error) {
	idCentral := sess.IDCentralFinanceiro()
	if idCentral == "" {
		return nil, errors.New("id_central obrigatorio (IDCentralUUID)")
	}
	if idFranqueado != "" {
		idCentral = centralDoFranqueado(idFranqueado, idCentral)
	}
	q := `
		SELECT c.ID_Contrato, c.ID_Franqueado, c.Nome, c.Status, c.ValorFinal, c.Periodicidade,
		       DATE_FORMAT(c.InicioEm,'%Y-%m-%d'), DATE_FORMAT(c.ProximaCobrancaEm,'%Y-%m-%d')
		FROM fp_contrato c
		WHERE c.ID_Central = ?
	`
	args := []any{idCentral}
	if idFranqueado != "" {
		if sess.UserTipo == "REP" {
			if err := AssertPodeEditarContrato(sess, idFranqueado); err != nil {
				return nil, err
			}
		}
		q += ` AND c.ID_Franqueado = ?`
		args = append(args, idFranqueado)
	} else if sess.UserTipo == "REP" {
		q += ` AND c.ID_Representante = ?`
		args = append(args, sess.IDRepresentante)
	}
	q += ` AND c.Status != 'cancelado'`
	q += ` ORDER BY c.ID_Contrato DESC LIMIT 200`

	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ContratoResumo
	for rows.Next() {
		var c ContratoResumo
		var prox sql.NullString
		if err := rows.Scan(&c.IDContrato, &c.IDFranqueado, &c.Nome, &c.Status, &c.ValorFinal, &c.Periodicidade, &c.InicioEm, &prox); err != nil {
			return nil, err
		}
		if prox.Valid {
			c.ProximaCobranca = prox.String
		}
		out = append(out, c)
	}
	if err := anexarItensContratos(out); err != nil {
		return nil, err
	}
	return out, rows.Err()
}

func ListarFaturas(sess auth.SessaoAdm, idFranqueado string) ([]FaturaResumo, error) {
	idCentral := sess.IDCentralFinanceiro()
	if idCentral == "" {
		return nil, errors.New("id_central obrigatorio (IDCentralUUID)")
	}
	if idFranqueado != "" {
		idCentral = centralDoFranqueado(idFranqueado, idCentral)
	}
	q := `
		SELECT f.ID_Fatura, f.ID_Contrato, f.Tipo, f.Referencia, f.Status, f.ValorTotal,
		       DATE_FORMAT(f.VencimentoEm,'%Y-%m-%d'), DATE_FORMAT(f.PagoEm,'%Y-%m-%d %H:%i:%s'),
		       COALESCE(f.ID_FaturaContabil, 0)
		FROM fp_fatura_contrato f
		WHERE f.ID_Central = ?
	`
	args := []any{idCentral}
	if idFranqueado != "" {
		q += ` AND f.ID_Franqueado = ?`
		args = append(args, idFranqueado)
	} else if sess.UserTipo == "REP" {
		q += ` AND f.ID_Representante = ?`
		args = append(args, sess.IDRepresentante)
	}
	q += ` ORDER BY f.ID_Fatura DESC LIMIT 200`

	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FaturaResumo
	for rows.Next() {
		var f FaturaResumo
		var pago sql.NullString
		if err := rows.Scan(&f.IDFatura, &f.IDContrato, &f.Tipo, &f.Referencia, &f.Status, &f.ValorTotal, &f.VencimentoEm, &pago, &f.IDFaturaContabil); err != nil {
			return nil, err
		}
		if pago.Valid {
			f.PagoEm = pago.String
		}
		out = append(out, f)
	}
	if err := anexarItensFaturas(out); err != nil {
		return nil, err
	}
	return out, rows.Err()
}

func WorkerSuspenderVencidas() (int, error) {
	res, err := db.Conn.Exec(`
		UPDATE fp_fatura_contrato SET Status = 'vencida'
		WHERE Status = 'aberta' AND VencimentoEm < CURDATE()
	`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()

	_, err = db.Conn.Exec(`
		UPDATE fp_contrato c
		INNER JOIN fp_fatura_contrato f ON f.ID_Contrato = c.ID_Contrato
		SET c.Status = 'suspenso'
		WHERE c.Status = 'ativo' AND f.Status = 'vencida'
	`)
	return int(n), err
}

func WorkerGerarRenovacoes(diasAntecedencia int) (int, error) {
	rows, err := db.Conn.Query(`
		SELECT ID_Contrato, ID_Franqueado, ID_Central, COALESCE(ID_Representante,''), ValorFinal, Periodicidade, ProximaCobrancaEm
		FROM fp_contrato
		WHERE Status = 'ativo' AND ProximaCobrancaEm IS NOT NULL
		  AND ProximaCobrancaEm <= DATE_ADD(CURDATE(), INTERVAL ? DAY)
	`, diasAntecedencia)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	geradas := 0
	for rows.Next() {
		var idC int
		var idFra, idCen, idRep, period string
		var valor float64
		var prox sql.NullTime
		if err := rows.Scan(&idC, &idFra, &idCen, &idRep, &valor, &period, &prox); err != nil {
			continue
		}
		var existe int
		_ = db.Conn.QueryRow(`
			SELECT COUNT(*) FROM fp_fatura_contrato
			WHERE ID_Contrato = ? AND Tipo = 'renovacao' AND Status = 'aberta'
		`, idC).Scan(&existe)
		if existe > 0 {
			continue
		}
		venc := time.Now().AddDate(0, 0, 7)
		if prox.Valid {
			venc = prox.Time.AddDate(0, 0, 7)
		}
		valorCobrar := ValorPorPeriodicidade(valor, period)
		tx, _ := db.Conn.Begin()
		fatID, err := gerarFaturaTx(tx, idC, idFra, idCen, idRep, "renovacao", valorCobrar, venc)
		if err != nil {
			tx.Rollback()
			continue
		}
		if prox.Valid {
			_, _ = tx.Exec(`UPDATE fp_contrato SET ProximaCobrancaEm = ? WHERE ID_Contrato = ?`,
				proximaCobranca(prox.Time, period), idC)
		}
		if tx.Commit() == nil {
			geradas++
			var nome string
			_ = db.Conn.QueryRow(`SELECT Nome FROM fp_contrato WHERE ID_Contrato = ?`, idC).Scan(&nome)
			sess := auth.SessaoAdm{UsuarioAdm: auth.UsuarioAdm{IDUsuario: "worker"}}
			if err := espelharFaturaContabil(idC, fatID, nome, sess); err != nil {
				log.Printf("avisar: contabilidade renovacao contrato %d fatura %d: %v", idC, fatID, err)
			}
		}
	}
	return geradas, nil
}

func nullStr(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt(n int) sql.NullInt64 {
	if n == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(n), Valid: true}
}

func nullFloat(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func nullDate(s string) sql.NullTime {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func coalesceSN(s, def string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "S" || s == "N" {
		return s
	}
	return def
}

func franqueadoBloqueadoGovernanca(idFranqueado string) (bool, error) {
	if !pgcredito.Configurado() {
		return false, nil
	}
	return pggovernanca.FranqueadoBloqueadoCascata(idFranqueado)
}
