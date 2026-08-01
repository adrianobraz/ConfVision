package evento

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const contactIdConfVisionAnalitico = "CV01"

// ResultadoConfVision IDs gerados para amarrar o vis_evento no Xano/worker.
type ResultadoConfVision struct {
	IDEvento      string `json:"idEvento"`
	IDProcesso    string `json:"idProcesso"`
	IDDispositivo string `json:"idDispositivo"`
	Conta         string `json:"conta"`
	Codigo        string `json:"codigo"`
	Particao      string `json:"particao"`
	ZonaUser      string `json:"zonaUser"`
}

// RecebeEventoConfVision injeta ALARME CV01 no pipeline do terminal com midia ConfVision.
func RecebeEventoConfVision(conn *sql.DB, idFranqueado, conta, particao, zonaUser string, visEventoID int64) (*ResultadoConfVision, error) {
	if conn == nil {
		return nil, errors.New("conexao invalida")
	}
	if strings.TrimSpace(idFranqueado) == "" {
		return nil, errors.New("idFranqueado obrigatorio")
	}
	if visEventoID <= 0 {
		return nil, errors.New("vis_evento_id invalido")
	}

	conta = padLeft(digitsOnly(conta), 4, '0')
	particao = padLeft(digitsOnly(particao), 2, '0')
	zonaUser = strings.TrimSpace(zonaUser)
	if zonaUser == "" {
		zonaUser = "001"
	} else if onlyDigits(zonaUser) {
		zonaUser = padLeft(zonaUser, 3, '0')
	}

	if conta == "" || conta == "0000" {
		return nil, errors.New("conta invalida")
	}

	eventoIn := conta + contactIdConfVisionAnalitico + particao + zonaUser

	var objEvt Evento
	objEvt.conn = conn
	objEvt.email = setupMail
	objEvt.idFranqueado = idFranqueado

	var periodico bool
	if err := testePeriodico(&objEvt, eventoIn, &periodico); err != nil {
		return nil, err
	}
	if periodico {
		return nil, errors.New("evento periodico ignorado")
	}

	if err := decodeEvento(eventoIn, &objEvt); err != nil {
		return nil, err
	}

	if objEvt.ativo != "S" {
		return nil, fmt.Errorf("dispositivo bloqueado ou inativo: %s", objEvt.msg)
	}

	if err := getIdProcesso(&objEvt); err != nil {
		return nil, err
	}

	if err := armarAlarme(&objEvt); err != nil {
		return nil, err
	}

	if err := verificaGrade(&objEvt); err != nil {
		return nil, err
	}

	if err := atualizaNivelById(&objEvt); err != nil {
		return nil, err
	}

	// Nao chama Benuvem: midia vem do ConfVision (foto/video/ao vivo pela licenca).
	objEvt.img = fmt.Sprintf(`{"confvision":true,"vis_evento_id":%d}`, visEventoID)

	if err := insere(objEvt); err != nil {
		return nil, err
	}

	if err := gravarUltimoEventoById(&objEvt); err != nil {
		return nil, err
	}

	if err := enviaEmail(&objEvt); err != nil {
		return nil, err
	}

	fmt.Printf("%s -> ConfVision CV01 evento=%s processo=%s vis=%d\n",
		objEvt.msg, objEvt.idEvento, objEvt.idProcesso, visEventoID)

	return &ResultadoConfVision{
		IDEvento:      objEvt.idEvento,
		IDProcesso:    objEvt.idProcesso,
		IDDispositivo: objEvt.idDispositivo,
		Conta:         objEvt.conta,
		Codigo:        objEvt.codigo,
		Particao:      objEvt.particao,
		ZonaUser:      objEvt.zonaUser,
	}, nil
}

func padLeft(s string, size int, pad rune) string {
	s = strings.TrimSpace(s)
	if len(s) >= size {
		if len(s) > size {
			return s[len(s)-size:]
		}
		return s
	}
	return strings.Repeat(string(pad), size-len(s)) + s
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func onlyDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
