package contrato

import (
	"apifunction/auth"
	"apifunction/config"
	"apifunction/db"
	"apifunction/xano"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

var xanoFin *xano.Client

func initXano() *xano.Client {
	if xanoFin == nil {
		xanoFin = xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	}
	return xanoFin
}

func contabilidadeObrigatoria() bool {
	return initXano().Enabled()
}

func espelharFaturaContabil(idContrato, idFatura int, nomeContrato string, sess auth.SessaoAdm) error {
	cli := initXano()
	if !cli.Enabled() {
		return nil
	}

	var idFra, idCen, idRep, tipo, ciclo, venc string
	var valor float64
	var idContabil sql.NullInt64
	err := db.Conn.QueryRow(`
		SELECT ID_Franqueado, ID_Central, COALESCE(ID_Representante,''), Tipo, CicloRef,
		       DATE_FORMAT(VencimentoEm,'%Y-%m-%d'), ValorTotal, ID_FaturaContabil
		FROM fp_fatura_contrato WHERE ID_Fatura = ?
	`, idFatura).Scan(&idFra, &idCen, &idRep, &tipo, &ciclo, &venc, &valor, &idContabil)
	if err != nil {
		return fmt.Errorf("carregar fatura contrato: %w", err)
	}
	if idContabil.Valid && idContabil.Int64 > 0 {
		return nil
	}

	itens, err := itensParaContabilidade(idContrato)
	if err != nil {
		return err
	}

	idXano, err := cli.SyncFaturaContrato(xano.SyncFaturaInput{
		IDContratoMySQL: idContrato,
		IDFaturaMySQL:   idFatura,
		IDFranqueado:    idFra,
		IDCentral:       idCen,
		IDRepresentante: idRep,
		TipoFatura:      tipo,
		ValorTotal:      valor,
		VencimentoEm:    venc,
		CicloRef:        ciclo,
		NomeContrato:    nomeContrato,
		AdminUsuario:    sess.IDUsuario,
		Itens:           itens,
	})
	if err != nil {
		return fmt.Errorf("sync contabilidade: %w", err)
	}
	if idXano <= 0 {
		return errors.New("sync contabilidade nao retornou id_fatura_contabil")
	}

	_, err = db.Conn.Exec(`UPDATE fp_fatura_contrato SET ID_FaturaContabil = ? WHERE ID_Fatura = ?`, idXano, idFatura)
	if err != nil {
		return fmt.Errorf("gravar id_fatura_contabil: %w", err)
	}
	return nil
}

func itensParaContabilidade(idContrato int) ([]xano.ItemSync, error) {
	rows, err := db.Conn.Query(`
		SELECT Descricao, Quantidade, ValorUnitario, ValorTotal, Chave
		FROM fp_contrato_item WHERE ID_Contrato = ? ORDER BY ID_ContratoItem
	`, idContrato)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []xano.ItemSync
	for rows.Next() {
		var desc, chave string
		var qtd int
		var vu, vt float64
		if err := rows.Scan(&desc, &qtd, &vu, &vt, &chave); err != nil {
			return nil, err
		}
		if strings.TrimSpace(desc) == "" {
			desc = chave
		}
		out = append(out, xano.ItemSync{
			Descricao:     desc,
			Quantidade:    qtd,
			ValorUnitario: vu,
			ValorTotal:    vt,
			RefChave:      chave,
		})
	}
	return out, rows.Err()
}

func cancelarFaturasContabilidade(faturas []faturaContabilRef, obs string, sess auth.SessaoAdm) {
	cli := initXano()
	if !cli.Enabled() {
		return
	}
	for _, f := range faturas {
		if err := cli.CancelarFaturaContrato(f.IDContrato, f.IDFatura, f.IDFaturaContabil, obs, sess.IDUsuario); err != nil {
			log.Printf("avisar: cancelar contabilidade fatura %d: %v", f.IDFatura, err)
		}
	}
}

func pagarFaturaContabil(idFatura int, valor float64, sess auth.SessaoAdm) error {
	cli := initXano()
	if !cli.Enabled() {
		return nil
	}

	var idContrato, idContabil int
	err := db.Conn.QueryRow(`
		SELECT ID_Contrato, COALESCE(ID_FaturaContabil, 0) FROM fp_fatura_contrato WHERE ID_Fatura = ?
	`, idFatura).Scan(&idContrato, &idContabil)
	if err != nil {
		return err
	}
	if idContabil <= 0 {
		var nome string
		_ = db.Conn.QueryRow(`SELECT Nome FROM fp_contrato WHERE ID_Contrato = ?`, idContrato).Scan(&nome)
		if err := espelharFaturaContabil(idContrato, idFatura, nome, sess); err != nil {
			return err
		}
		_ = db.Conn.QueryRow(`SELECT COALESCE(ID_FaturaContabil, 0) FROM fp_fatura_contrato WHERE ID_Fatura = ?`, idFatura).Scan(&idContabil)
	}
	if idContabil <= 0 {
		return errors.New("fatura sem espelho contabil")
	}
	return cli.PagarFaturaContrato(idContabil, idContrato, idFatura, valor, sess.IDUsuario)
}

type faturaContabilRef struct {
	IDContrato       int
	IDFatura         int
	IDFaturaContabil int
}

func listarFaturasAbertasContabil(idContrato int, tx *sql.Tx) ([]faturaContabilRef, error) {
	var rows *sql.Rows
	var err error
	q := `
		SELECT ID_Fatura, COALESCE(ID_FaturaContabil, 0)
		FROM fp_fatura_contrato
		WHERE ID_Contrato = ? AND Status IN ('aberta','vencida')
	`
	if tx != nil {
		rows, err = tx.Query(q, idContrato)
	} else {
		rows, err = db.Conn.Query(q, idContrato)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []faturaContabilRef
	for rows.Next() {
		var f faturaContabilRef
		f.IDContrato = idContrato
		if err := rows.Scan(&f.IDFatura, &f.IDFaturaContabil); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func erroContabilidade(err error) error {
	if err == nil || !contabilidadeObrigatoria() {
		return err
	}
	return fmt.Errorf("contabilidade: %w", err)
}

// avisarEstornoContabil estorna espelho Xano (caixa, assinatura, licenca CV).
func estornarFaturaContabil(idFaturaContabil int, motivo string, sess auth.SessaoAdm) error {
	cli := initXano()
	if !cli.Enabled() || idFaturaContabil <= 0 {
		return nil
	}
	if strings.TrimSpace(motivo) == "" {
		motivo = "Estorno — módulo Contrato"
	}
	return cli.EstornarFaturaContabil(idFaturaContabil, motivo, sess.IDUsuario)
}

func logContabilidadeOpcional(msg string, err error) {
	if err == nil {
		return
	}
	if !contabilidadeObrigatoria() {
		log.Printf("avisar: %s: %v", msg, err)
		return
	}
	log.Printf("contabilidade: %s: %v", msg, err)
}
