package service

import (
	"apifunction/db"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

func Preview(req ReqTransferencia) (PreviewResult, error) {
	req.Tipo = strings.ToUpper(strings.TrimSpace(req.Tipo))
	req.IDOrigem = strings.TrimSpace(req.IDOrigem)
	req.IDDestino = strings.TrimSpace(req.IDDestino)

	switch req.Tipo {
	case TipoCLI:
		return previewCLI(req)
	case TipoFRA:
		return previewFRA(req)
	case TipoREP:
		return previewREP(req)
	default:
		return PreviewResult{}, errors.New("tipo invalido — use CLI, FRA ou REP")
	}
}

func Executar(req ReqTransferencia, idUsuario string) (ExecResult, error) {
	prev, err := Preview(req)
	if err != nil {
		return ExecResult{}, err
	}
	if !prev.OK {
		return ExecResult{}, errors.New("transferencia bloqueada — execute preview primeiro")
	}

	logID := strings.ReplaceAll(uuid.New().String(), "-", "")
	previewJSON, _ := json.Marshal(prev)

	tx, err := db.Conn.Begin()
	if err != nil {
		return ExecResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var execErr error
	switch strings.ToUpper(strings.TrimSpace(req.Tipo)) {
	case TipoCLI:
		execErr = executarCLI(tx, req)
	case TipoFRA:
		execErr = executarFRA(tx, req)
	case TipoREP:
		execErr = executarREP(tx, req)
	default:
		return ExecResult{}, errors.New("tipo invalido")
	}
	if execErr != nil {
		_ = insertLog(tx, logID, idUsuario, req, prev, "ERRO", execErr.Error(), previewJSON)
		_ = tx.Commit()
		return ExecResult{}, execErr
	}

	if err := insertLog(tx, logID, idUsuario, req, prev, "OK", "", previewJSON); err != nil {
		return ExecResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExecResult{}, err
	}
	return ExecResult{LogID: logID}, nil
}

func previewCLI(req ReqTransferencia) (PreviewResult, error) {
	out := PreviewResult{Contagens: Contagens{}}
	if req.IDOrigem == "" || req.IDDestino == "" {
		return out, errors.New("idOrigem e idDestino obrigatorios")
	}
	if req.IDOrigem == req.IDDestino {
		return out, errors.New("origem e destino iguais")
	}

	var idFraAtual, nome sql.NullString
	err := db.Conn.QueryRow(`
		SELECT cliente.ID_Franqueado, cliente.Nome
		FROM cliente
		WHERE cliente.ID_Cliente = ?
		LIMIT 1
	`, req.IDOrigem).Scan(&idFraAtual, &nome)
	if err == sql.ErrNoRows {
		return out, errors.New("cliente nao encontrado")
	}
	if err != nil {
		return out, err
	}
	out.DescricaoOrigem = strings.TrimSpace(nome.String)
	out.VinculoAntigo = strings.TrimSpace(idFraAtual.String)
	out.VinculoNovo = req.IDDestino

	if strings.TrimSpace(idFraAtual.String) == req.IDDestino {
		out.Bloqueios = append(out.Bloqueios, Bloqueio{Codigo: "MESMO_VINCULO", Mensagem: "cliente ja pertence ao franqueado destino"})
		out.OK = false
		return out, nil
	}

	bloqueios, contagens, err := validarCliente(req.IDOrigem)
	if err != nil {
		return out, err
	}
	out.Contagens = contagens
	out.Bloqueios = append(out.Bloqueios, bloqueios...)

	if bloqueio, err := validarFranqueadoDestino(req.IDDestino); err != nil {
		return out, err
	} else if bloqueio != nil {
		out.Bloqueios = append(out.Bloqueios, *bloqueio)
	}

	nomeFra, _ := nomeFranqueado(req.IDDestino)
	out.DescricaoDestino = nomeFra
	out.OK = len(out.Bloqueios) == 0
	return out, nil
}

func previewFRA(req ReqTransferencia) (PreviewResult, error) {
	out := PreviewResult{Contagens: Contagens{}}
	if req.IDOrigem == "" || req.IDDestino == "" {
		return out, errors.New("idOrigem e idDestino obrigatorios")
	}

	var idRepAtual, razao sql.NullString
	err := db.Conn.QueryRow(`
		SELECT franqueado.ID_Representante, franqueado.RazaoSocial
		FROM franqueado
		WHERE franqueado.ID_Franqueado = ?
		LIMIT 1
	`, req.IDOrigem).Scan(&idRepAtual, &razao)
	if err == sql.ErrNoRows {
		return out, errors.New("franqueado nao encontrado")
	}
	if err != nil {
		return out, err
	}
	out.DescricaoOrigem = strings.TrimSpace(razao.String)
	out.VinculoAntigo = strings.TrimSpace(idRepAtual.String)
	out.VinculoNovo = req.IDDestino

	if strings.TrimSpace(idRepAtual.String) == req.IDDestino {
		out.Bloqueios = append(out.Bloqueios, Bloqueio{Codigo: "MESMO_VINCULO", Mensagem: "franqueado ja pertence ao representante destino"})
		out.OK = false
		return out, nil
	}

	bloqueios, contagens, err := validarFranqueado(req.IDOrigem)
	if err != nil {
		return out, err
	}
	out.Contagens = contagens
	out.Bloqueios = append(out.Bloqueios, bloqueios...)

	clientes, err := idsClientesFranqueado(req.IDOrigem)
	if err != nil {
		return out, err
	}
	out.Contagens.Clientes = len(clientes)
	for _, idCli := range clientes {
		bs, c, err := validarCliente(idCli)
		if err != nil {
			return out, err
		}
		out.Contagens.Eventos += c.Eventos
		out.Contagens.Faturas += c.Faturas
		out.Contagens.ProcessosAbertos += c.ProcessosAbertos
		out.Contagens.TicketsAbertos += c.TicketsAbertos
		for _, b := range bs {
			b.Entidade = "cliente"
			b.IDEntidade = idCli
			out.Bloqueios = append(out.Bloqueios, b)
		}
	}

	if bloqueio, err := validarRepresentanteDestino(req.IDDestino); err != nil {
		return out, err
	} else if bloqueio != nil {
		out.Bloqueios = append(out.Bloqueios, *bloqueio)
	}

	nomeRep, _ := nomeRepresentante(req.IDDestino)
	out.DescricaoDestino = nomeRep
	out.OK = len(out.Bloqueios) == 0
	return out, nil
}

func previewREP(req ReqTransferencia) (PreviewResult, error) {
	out := PreviewResult{Contagens: Contagens{}}
	if req.IDOrigem == "" || req.IDDestino == "" {
		return out, errors.New("idOrigem e idDestino obrigatorios")
	}

	var idCentralAtual, razao sql.NullString
	err := db.Conn.QueryRow(`
		SELECT COALESCE(representante.IDCentralUUID, ''), representante.RazaoSocial
		FROM representante
		WHERE representante.ID_Representante = ?
		LIMIT 1
	`, req.IDOrigem).Scan(&idCentralAtual, &razao)
	if err == sql.ErrNoRows {
		return out, errors.New("representante nao encontrado")
	}
	if err != nil {
		return out, err
	}
	out.DescricaoOrigem = strings.TrimSpace(razao.String)
	out.VinculoAntigo = strings.TrimSpace(idCentralAtual.String)

	centralUUID, centralNome, err := resolverCentralDestino(req.IDDestino)
	if err != nil {
		return out, err
	}
	out.VinculoNovo = centralUUID
	out.DescricaoDestino = centralNome

	if centralUUID == strings.TrimSpace(idCentralAtual.String) {
		out.Bloqueios = append(out.Bloqueios, Bloqueio{Codigo: "MESMO_VINCULO", Mensagem: "representante ja pertence a central destino"})
		out.OK = false
		return out, nil
	}

	bloqueios, err := validarRepresentante(req.IDOrigem)
	if err != nil {
		return out, err
	}
	out.Bloqueios = append(out.Bloqueios, bloqueios...)

	franqueados, err := idsFranqueadosRepresentante(req.IDOrigem)
	if err != nil {
		return out, err
	}
	out.Contagens.Franqueados = len(franqueados)
	for _, idFra := range franqueados {
		bs, c, err := validarFranqueado(idFra)
		if err != nil {
			return out, err
		}
		out.Contagens.Eventos += c.Eventos
		out.Contagens.Faturas += c.Faturas
		out.Contagens.ProcessosAbertos += c.ProcessosAbertos
		out.Contagens.TicketsAbertos += c.TicketsAbertos
		for _, b := range bs {
			b.Entidade = "franqueado"
			b.IDEntidade = idFra
			out.Bloqueios = append(out.Bloqueios, b)
		}
		clientes, _ := idsClientesFranqueado(idFra)
		out.Contagens.Clientes += len(clientes)
		for _, idCli := range clientes {
			bs, c, err := validarCliente(idCli)
			if err != nil {
				return out, err
			}
			out.Contagens.Eventos += c.Eventos
			out.Contagens.Faturas += c.Faturas
			out.Contagens.ProcessosAbertos += c.ProcessosAbertos
			out.Contagens.TicketsAbertos += c.TicketsAbertos
			for _, b := range bs {
				b.Entidade = "cliente"
				b.IDEntidade = idCli
				out.Bloqueios = append(out.Bloqueios, b)
			}
		}
	}

	out.OK = len(out.Bloqueios) == 0
	return out, nil
}

func executarCLI(tx *sql.Tx, req ReqTransferencia) error {
	centralUUID, err := centralUUIDPorFranqueado(req.IDDestino)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE cliente SET ID_Franqueado = ? WHERE ID_Cliente = ?`, req.IDDestino, req.IDOrigem)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("cliente nao atualizado")
	}
	_, err = tx.Exec(`
		UPDATE usuarios SET IDCentralUUID = ?
		WHERE ID_Vinculo = ? AND (IDCentralUUID IS NULL OR IDCentralUUID <> ?)
	`, centralUUID, req.IDOrigem, centralUUID)
	return err
}

func executarFRA(tx *sql.Tx, req ReqTransferencia) error {
	centralUUID, err := centralUUIDPorRepresentante(req.IDDestino)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE franqueado SET ID_Representante = ? WHERE ID_Franqueado = ?`, req.IDDestino, req.IDOrigem)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("franqueado nao atualizado")
	}
	_, err = tx.Exec(`
		UPDATE usuarios SET IDCentralUUID = ?
		WHERE ID_Vinculo = ?
	`, centralUUID, req.IDOrigem)
	return err
}

func executarREP(tx *sql.Tx, req ReqTransferencia) error {
	centralUUID, _, err := resolverCentralDestino(req.IDDestino)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`
		UPDATE representante SET IDCentralUUID = ?
		WHERE ID_Representante = ?
	`, centralUUID, req.IDOrigem)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("representante nao atualizado")
	}
	_, err = tx.Exec(`
		UPDATE usuarios SET IDCentralUUID = ?
		WHERE ID_Vinculo = ?
	`, centralUUID, req.IDOrigem)
	if err != nil {
		return err
	}
	franqueados, err := idsFranqueadosRepresentanteTx(tx, req.IDOrigem)
	if err != nil {
		return err
	}
	for _, idFra := range franqueados {
		if _, err := tx.Exec(`
			UPDATE usuarios SET IDCentralUUID = ?
			WHERE ID_Vinculo = ?
		`, centralUUID, idFra); err != nil {
			return err
		}
	}
	return nil
}

func insertLog(tx *sql.Tx, logID, idUsuario string, req ReqTransferencia, prev PreviewResult, status, errMsg string, previewJSON []byte) error {
	_, err := tx.Exec(`
		INSERT INTO transferencia_vinculo_log (
			ID_Log, DataOperacao, ID_Usuario, Tipo, ID_Origem, ID_Destino,
			VinculoAntigo, VinculoNovo, Motivo, PreviewJSON, Status, ErroMsg
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		logID,
		time.Now().Format("2006-01-02 15:04:05"),
		idUsuario,
		strings.ToUpper(req.Tipo),
		req.IDOrigem,
		req.IDDestino,
		prev.VinculoAntigo,
		prev.VinculoNovo,
		strings.TrimSpace(req.Motivo),
		string(previewJSON),
		status,
		errMsg,
	)
	return err
}

func resolverCentralDestino(idCentralRef string) (uuid, nome string, err error) {
	idCentralRef = strings.TrimSpace(idCentralRef)
	var idCentral, idCentralUUID, razao sql.NullString
	err = db.Conn.QueryRow(`
		SELECT ID_Central, IDCentralUUID, COALESCE(NomeFantasia, RazaoSocial, ID_Central)
		FROM central
		WHERE ID_Central = ? OR IDCentralUUID = ?
		LIMIT 1
	`, idCentralRef, idCentralRef).Scan(&idCentral, &idCentralUUID, &razao)
	if err == sql.ErrNoRows {
		return "", "", errors.New("central destino nao encontrada")
	}
	if err != nil {
		return "", "", err
	}
	uuid = strings.TrimSpace(idCentralUUID.String)
	if uuid == "" {
		uuid = strings.TrimSpace(idCentral.String)
	}
	if uuid == "" {
		return "", "", errors.New("central destino sem identificador")
	}
	return uuid, strings.TrimSpace(razao.String), nil
}

func centralUUIDPorFranqueado(idFranqueado string) (string, error) {
	var idRep sql.NullString
	err := db.Conn.QueryRow(`
		SELECT ID_Representante FROM franqueado WHERE ID_Franqueado = ? LIMIT 1
	`, idFranqueado).Scan(&idRep)
	if err != nil {
		return "", err
	}
	return centralUUIDPorRepresentante(strings.TrimSpace(idRep.String))
}

func centralUUIDPorRepresentante(idRep string) (string, error) {
	var uuid sql.NullString
	err := db.Conn.QueryRow(`
		SELECT COALESCE(IDCentralUUID, '') FROM representante WHERE ID_Representante = ? LIMIT 1
	`, idRep).Scan(&uuid)
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(uuid.String)
	if out == "" {
		return "", errors.New("representante sem IDCentralUUID")
	}
	return out, nil
}

func nomeFranqueado(id string) (string, error) {
	var n sql.NullString
	err := db.Conn.QueryRow(`SELECT RazaoSocial FROM franqueado WHERE ID_Franqueado = ?`, id).Scan(&n)
	return strings.TrimSpace(n.String), err
}

func nomeRepresentante(id string) (string, error) {
	var n sql.NullString
	err := db.Conn.QueryRow(`SELECT RazaoSocial FROM representante WHERE ID_Representante = ?`, id).Scan(&n)
	return strings.TrimSpace(n.String), err
}

func idsClientesFranqueado(idFra string) ([]string, error) {
	rows, err := db.Conn.Query(`SELECT ID_Cliente FROM cliente WHERE ID_Franqueado = ?`, idFra)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func idsFranqueadosRepresentante(idRep string) ([]string, error) {
	return idsFranqueadosRepresentanteTx(db.Conn, idRep)
}

func idsFranqueadosRepresentanteTx(q querier, idRep string) ([]string, error) {
	rows, err := q.Query(`SELECT ID_Franqueado FROM franqueado WHERE ID_Representante = ?`, idRep)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

type querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}
