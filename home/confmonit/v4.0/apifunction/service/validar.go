package service

import (
	"apifunction/db"
	"database/sql"
	"fmt"
	"strings"
)

func validarCliente(idCliente string) ([]Bloqueio, Contagens, error) {
	var bloqueios []Bloqueio
	var c Contagens

	if bloqueado, msg, err := entidadeBloqueada(idCliente); err != nil {
		return nil, c, err
	} else if bloqueado {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "BLOQUEADO", Mensagem: msg, IDEntidade: idCliente})
	}

	var cancel sql.NullTime
	if err := db.Conn.QueryRow(`
		SELECT DataCancelamento FROM cliente WHERE ID_Cliente = ? LIMIT 1
	`, idCliente).Scan(&cancel); err == nil && cancel.Valid {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "CANCELADO", Mensagem: "cliente cancelado", IDEntidade: idCliente})
	}

	n, err := contarEventosCliente(idCliente)
	if err != nil {
		return nil, c, err
	}
	c.Eventos = n
	if n > MaxEventos {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "EVENTOS", Mensagem: fmt.Sprintf("cliente possui %d eventos (limite %d)", n, MaxEventos), Contagem: n, IDEntidade: idCliente,
		})
	}

	n, err = contarFaturasEntidade(idCliente)
	if err != nil {
		return nil, c, err
	}
	c.Faturas = n
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "FATURAS", Mensagem: fmt.Sprintf("cliente possui %d fatura(s) no MySQL", n), Contagem: n, IDEntidade: idCliente,
		})
	}

	n, err = contarProcessosAbertosCliente(idCliente)
	if err != nil {
		return nil, c, err
	}
	c.ProcessosAbertos = n
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "PROCESSO_ABERTO", Mensagem: fmt.Sprintf("%d processo(s) de atendimento aberto(s)", n), Contagem: n, IDEntidade: idCliente,
		})
	}

	n, err = contarTicketsAbertos(idCliente)
	if err != nil {
		return nil, c, err
	}
	c.TicketsAbertos = n
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "TICKETS", Mensagem: fmt.Sprintf("%d ticket(s) aberto(s)", n), Contagem: n, IDEntidade: idCliente,
		})
	}

	return bloqueios, c, nil
}

func validarFranqueado(idFranqueado string) ([]Bloqueio, Contagens, error) {
	var bloqueios []Bloqueio
	var c Contagens

	if bloqueado, msg, err := entidadeBloqueada(idFranqueado); err != nil {
		return nil, c, err
	} else if bloqueado {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "BLOQUEADO", Mensagem: msg, IDEntidade: idFranqueado})
	}

	var cancel sql.NullTime
	var usaCV sql.NullString
	if err := db.Conn.QueryRow(`
		SELECT DataCancelamento, COALESCE(UsaConfVision, 'N')
		FROM franqueado WHERE ID_Franqueado = ? LIMIT 1
	`, idFranqueado).Scan(&cancel, &usaCV); err != nil {
		return nil, c, err
	}
	if cancel.Valid {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "CANCELADO", Mensagem: "franqueado cancelado", IDEntidade: idFranqueado})
	}
	if strings.ToUpper(strings.TrimSpace(usaCV.String)) == "S" {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "CONFVISION", Mensagem: "franqueado com UsaConfVision = S", IDEntidade: idFranqueado})
	}

	if sms, err := temSmsListaEnvio(idFranqueado); err != nil {
		return nil, c, err
	} else if sms {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "WHATSAPP_SMS", Mensagem: "franqueado com SMS ativo em listaEnvio", IDEntidade: idFranqueado})
	}

	n, err := contarFaturasEntidade(idFranqueado)
	if err != nil {
		return nil, c, err
	}
	c.Faturas = n
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "FATURAS", Mensagem: fmt.Sprintf("franqueado possui %d fatura(s)", n), Contagem: n, IDEntidade: idFranqueado,
		})
	}

	n, err = contarTicketsAbertos(idFranqueado)
	if err != nil {
		return nil, c, err
	}
	c.TicketsAbertos = n
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "TICKETS", Mensagem: fmt.Sprintf("%d ticket(s) aberto(s)", n), Contagem: n, IDEntidade: idFranqueado})
	}

	if admFin, err := masterAdmFinanceiro(idFranqueado); err != nil {
		return nil, c, err
	} else if admFin {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "ADM_FINANCEIRO", Mensagem: "master do franqueado com AdmFinanceiro = S", IDEntidade: idFranqueado})
	}

	clientes, _ := idsClientesFranqueado(idFranqueado)
	for _, idCli := range clientes {
		ev, err := contarEventosCliente(idCli)
		if err != nil {
			return nil, c, err
		}
		c.Eventos += ev
		pa, err := contarProcessosAbertosCliente(idCli)
		if err != nil {
			return nil, c, err
		}
		c.ProcessosAbertos += pa
	}
	if c.Eventos > MaxEventos {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "EVENTOS", Mensagem: fmt.Sprintf("franqueado soma %d eventos nos clientes (limite %d)", c.Eventos, MaxEventos), Contagem: c.Eventos, IDEntidade: idFranqueado,
		})
	}
	if c.ProcessosAbertos > 0 {
		bloqueios = append(bloqueios, Bloqueio{
			Codigo: "PROCESSO_ABERTO", Mensagem: fmt.Sprintf("%d processo(s) aberto(s) nos clientes", c.ProcessosAbertos), Contagem: c.ProcessosAbertos, IDEntidade: idFranqueado,
		})
	}

	return bloqueios, c, nil
}

func validarRepresentante(idRep string) ([]Bloqueio, error) {
	var bloqueios []Bloqueio

	if bloqueado, msg, err := entidadeBloqueada(idRep); err != nil {
		return nil, err
	} else if bloqueado {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "BLOQUEADO", Mensagem: msg, IDEntidade: idRep})
	}

	var cancel sql.NullTime
	var usaAdm sql.NullString
	if err := db.Conn.QueryRow(`
		SELECT DataCancelamento, COALESCE(UsaAdmConfmonit, 'N')
		FROM representante WHERE ID_Representante = ? LIMIT 1
	`, idRep).Scan(&cancel, &usaAdm); err != nil {
		return nil, err
	}
	if cancel.Valid {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "CANCELADO", Mensagem: "representante cancelado", IDEntidade: idRep})
	}
	if strings.ToUpper(strings.TrimSpace(usaAdm.String)) == "S" {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "USA_ADMCONFMONIT", Mensagem: "representante com UsaAdmConfmonit = S (historico financeiro)", IDEntidade: idRep})
	}

	n, err := contarFaturasEntidade(idRep)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "FATURAS", Mensagem: fmt.Sprintf("representante possui %d fatura(s)", n), Contagem: n, IDEntidade: idRep})
	}

	n, err = contarTicketsAbertos(idRep)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		bloqueios = append(bloqueios, Bloqueio{Codigo: "TICKETS", Mensagem: fmt.Sprintf("%d ticket(s) aberto(s)", n), Contagem: n, IDEntidade: idRep})
	}

	return bloqueios, nil
}

func validarFranqueadoDestino(idFra string) (*Bloqueio, error) {
	var cancel sql.NullTime
	err := db.Conn.QueryRow(`SELECT DataCancelamento FROM franqueado WHERE ID_Franqueado = ?`, idFra).Scan(&cancel)
	if err == sql.ErrNoRows {
		b := Bloqueio{Codigo: "DESTINO_INVALIDO", Mensagem: "franqueado destino nao encontrado"}
		return &b, nil
	}
	if err != nil {
		return nil, err
	}
	if cancel.Valid {
		b := Bloqueio{Codigo: "DESTINO_INVALIDO", Mensagem: "franqueado destino cancelado"}
		return &b, nil
	}
	if bloqueado, msg, err := entidadeBloqueada(idFra); err != nil {
		return nil, err
	} else if bloqueado {
		b := Bloqueio{Codigo: "DESTINO_INVALIDO", Mensagem: "franqueado destino bloqueado: " + msg}
		return &b, nil
	}
	return nil, nil
}

func validarRepresentanteDestino(idRep string) (*Bloqueio, error) {
	var cancel sql.NullTime
	err := db.Conn.QueryRow(`SELECT DataCancelamento FROM representante WHERE ID_Representante = ?`, idRep).Scan(&cancel)
	if err == sql.ErrNoRows {
		b := Bloqueio{Codigo: "DESTINO_INVALIDO", Mensagem: "representante destino nao encontrado"}
		return &b, nil
	}
	if err != nil {
		return nil, err
	}
	if cancel.Valid {
		b := Bloqueio{Codigo: "DESTINO_INVALIDO", Mensagem: "representante destino cancelado"}
		return &b, nil
	}
	if bloqueado, msg, err := entidadeBloqueada(idRep); err != nil {
		return nil, err
	} else if bloqueado {
		b := Bloqueio{Codigo: "DESTINO_INVALIDO", Mensagem: "representante destino bloqueado: " + msg}
		return &b, nil
	}
	return nil, nil
}

func entidadeBloqueada(idAlvo string) (bool, string, error) {
	var desc sql.NullString
	err := db.Conn.QueryRow(`SELECT Descricao FROM listaBloqueio WHERE ID_Alvo = ? LIMIT 1`, idAlvo).Scan(&desc)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, strings.TrimSpace(desc.String), nil
}

func contarEventosCliente(idCliente string) (int, error) {
	var n int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM evento e
		INNER JOIN processo p ON e.ID_Processo = p.ID_Processo
		INNER JOIN dispositivo d ON p.ID_Dispositivo = d.ID_Dispositivo
		WHERE d.ID_Cliente = ?
	`, idCliente).Scan(&n)
	return n, err
}

func contarFaturasEntidade(id string) (int, error) {
	var n int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM faturas
		WHERE ID_Origem = ? OR ID_Destino = ?
	`, id, id).Scan(&n)
	return n, err
}

func contarProcessosAbertosCliente(idCliente string) (int, error) {
	var n int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM processo p
		INNER JOIN dispositivo d ON p.ID_Dispositivo = d.ID_Dispositivo
		WHERE d.ID_Cliente = ? AND p.DataAtenFim IS NULL
	`, idCliente).Scan(&n)
	return n, err
}

func contarTicketsAbertos(idAlvo string) (int, error) {
	var n int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM ticket
		WHERE (ID_Master = ? OR ID_Slave = ?)
		  AND Status IN ('NOVO', 'AGUARDANDO RESPOSTA', 'RESPONDIDO')
	`, idAlvo, idAlvo).Scan(&n)
	return n, err
}

func temSmsListaEnvio(idAlvo string) (bool, error) {
	var sms sql.NullString
	err := db.Conn.QueryRow(`SELECT Sms FROM listaEnvio WHERE ID_Alvo = ? LIMIT 1`, idAlvo).Scan(&sms)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.ToUpper(strings.TrimSpace(sms.String)) == "S", nil
}

func masterAdmFinanceiro(idVinculo string) (bool, error) {
	var adm sql.NullString
	err := db.Conn.QueryRow(`
		SELECT COALESCE(AdmFinanceiro, 'N') FROM usuarios
		WHERE ID_Vinculo = ? AND Master = 'S'
		LIMIT 1
	`, idVinculo).Scan(&adm)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.ToUpper(strings.TrimSpace(adm.String)) == "S", nil
}
