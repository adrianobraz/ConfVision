package compatec

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"receptorWeb/src/auxiliar"
	"receptorWeb/src/config"
	evento "receptorWeb/src/eventos"
	"strings"
	"time"
)

type sessao struct {
	imeiSufixo  string
	conta       string
	idFranquead string
}

func Start() {
	if !config.Compatec.Ligar {
		logCompatec("servico desabilitado por configuracao")
		return
	}

	endereco := fmt.Sprintf(":%s", config.Compatec.Porta)
	ln, err := net.Listen("tcp", endereco)
	if err != nil {
		log.Printf("compatec: erro ao iniciar listener em %s: %v", endereco, err)
		return
	}
	logCompatec("listener iniciado em %s", endereco)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("compatec: erro ao aceitar conexao: %v", err)
			continue
		}
		go processaConexao(conn)
	}
}

func processaConexao(conn net.Conn) {
	defer conn.Close()
	logCompatec("cliente conectado: %s", conn.RemoteAddr().String())

	estado := &sessao{}
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 512)
	timeout := time.Duration(config.Compatec.ReadTimeoutSec) * time.Second

	for {
		if timeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(timeout))
		}

		n, err := conn.Read(tmp)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logCompatec("cliente desconectado: %s", conn.RemoteAddr().String())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logCompatec("timeout de leitura, fechando conexao: %s", conn.RemoteAddr().String())
				return
			}
			log.Printf("compatec: erro de leitura %s: %v", conn.RemoteAddr().String(), err)
			return
		}
		if n <= 0 {
			continue
		}

		buf = append(buf, tmp[:n]...)
		for {
			msg, consumido, ok := extrairMensagem(buf)
			if !ok {
				break
			}
			buf = buf[consumido:]
			processaMensagem(conn, estado, msg)
		}
	}
}

func processaMensagem(conn net.Conn, estado *sessao, msg []byte) {
	if len(msg) == 0 {
		return
	}

	switch msg[0] {
	case '*':
		estado.imeiSufixo = strings.ToUpper(strings.TrimSpace(string(msg[1:])))
		logCompatec("imei recebido: %s", estado.imeiSufixo)
		_ = enviaAck(conn, '+')
	case '#':
		if len(msg) < 5 {
			return
		}
		estado.conta = strings.TrimSpace(string(msg[1:5]))
		idFra, err := resolveIdFranqueado(estado.conta, estado.imeiSufixo)
		if err != nil {
			log.Printf("compatec: erro ao resolver idFranqueado conta=%s imeiSufixo=%s: %v", estado.conta, estado.imeiSufixo, err)
		} else {
			estado.idFranquead = idFra
			logCompatec("conta recebida: %s | idFranqueado=%s", estado.conta, estado.idFranquead)
		}
		_ = enviaAck(conn, '@')
	case '@':
		_ = enviaAck(conn, '@')
	case '$':
		cid, err := extrairCID(msg)
		if err != nil {
			log.Printf("compatec: CID invalido: %v", err)
			return
		}

		contaCID := cid[0:4]
		idFra := estado.idFranquead
		if idFra == "" || estado.conta != contaCID {
			idFra, err = resolveIdFranqueado(contaCID, estado.imeiSufixo)
			if err != nil {
				log.Printf("compatec: falha ao resolver idFranqueado para evento conta=%s: %v", contaCID, err)
				return
			}
			estado.conta = contaCID
			estado.idFranquead = idFra
		}

		db, err := auxiliar.Conectar()
		if err != nil {
			log.Printf("compatec: erro conexao banco: %v", err)
			return
		}
		defer db.Close()

		if err := evento.RecebeEvento(db, idFra, cid); err != nil {
			log.Printf("compatec: erro ao processar evento idFranqueado=%s cid=%s: %v", idFra, cid, err)
			return
		}

		logCompatec("evento processado idFranqueado=%s cid=%s", idFra, cid)
		_ = enviaAck(conn, '@')
	}
}

func extrairMensagem(buf []byte) ([]byte, int, bool) {
	if len(buf) == 0 {
		return nil, 0, false
	}

	switch buf[0] {
	case '*':
		if len(buf) < 7 {
			return nil, 0, false
		}
		return buf[:7], 7, true
	case '#':
		if len(buf) < 5 {
			return nil, 0, false
		}
		return buf[:5], 5, true
	case '@':
		return buf[:1], 1, true
	case '$':
		for i := 1; i < len(buf); i++ {
			if buf[i] == 0xB6 {
				return buf[:i+1], i + 1, true
			}
		}
		return nil, 0, false
	default:
		return buf[:1], 1, true
	}
}

func extrairCID(msg []byte) (string, error) {
	if len(msg) < 15 || msg[0] != '$' {
		return "", errors.New("mensagem CID incompleta")
	}
	cid := strings.TrimSpace(string(msg[1:14]))
	if len(cid) != 13 {
		return "", errors.New("formato CID invalido")
	}
	return cid, nil
}

func enviaAck(conn net.Conn, ack byte) error {
	_, err := conn.Write([]byte{ack})
	return err
}

func resolveIdFranqueado(conta, imeiSufixo string) (string, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	if strings.TrimSpace(conta) == "" {
		return "", errors.New("conta nao informada")
	}

	if strings.TrimSpace(imeiSufixo) != "" {
		if idFra, err := buscaIdFranqueadoPorContaESufixo(db, conta, strings.ToUpper(imeiSufixo)); err == nil {
			return idFra, nil
		}
	}

	return buscaIdFranqueadoPorConta(db, conta)
}

func buscaIdFranqueadoPorContaESufixo(db *sql.DB, conta, sufixo string) (string, error) {
	tab, err := db.Query(`
		SELECT franqueado.ID_Franqueado
		FROM dispositivo
		INNER JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente
		INNER JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado
		WHERE dispositivo.Conta = ?
		AND (
			UPPER(dispositivo.IdFisico1) LIKE ?
			OR UPPER(dispositivo.IdFisico2) LIKE ?
		)
		LIMIT 1
	`, conta, "%"+sufixo, "%"+sufixo)
	if err != nil {
		return "", err
	}
	defer tab.Close()

	if tab.Next() {
		var idFra sql.NullString
		if err := tab.Scan(&idFra); err != nil {
			return "", err
		}
		if idFra.Valid {
			return idFra.String, nil
		}
	}
	return "", errors.New("franqueado nao encontrado por conta+sufixo")
}

func buscaIdFranqueadoPorConta(db *sql.DB, conta string) (string, error) {
	tab, err := db.Query(`
		SELECT franqueado.ID_Franqueado
		FROM dispositivo
		INNER JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente
		INNER JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado
		WHERE dispositivo.Conta = ?
		LIMIT 1
	`, conta)
	if err != nil {
		return "", err
	}
	defer tab.Close()

	if tab.Next() {
		var idFra sql.NullString
		if err := tab.Scan(&idFra); err != nil {
			return "", err
		}
		if idFra.Valid {
			return idFra.String, nil
		}
	}
	return "", errors.New("franqueado nao encontrado por conta")
}

func logCompatec(formato string, args ...any) {
	if !config.ExibirLog {
		return
	}
	log.Printf("compatec: "+formato, args...)
}
