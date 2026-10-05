package gerenciarSetoresAlarme

import (
	"confvision/src/conexao"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

const (
	setorCameraPadrao    = "S"
	setorTipoSetorPadrao = "CAMERA"
)

type respostaInserirSetor struct {
	Dados string `json:"dados"`
}

func aplicarDefaultsSetorConfVision(corpo []byte) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(corpo, &payload); err != nil {
		return corpo, err
	}

	payload["camera"] = setorCameraPadrao

	return json.Marshal(payload)
}

func extrairIdSetorInserido(corpo []byte) (string, error) {
	var resposta respostaInserirSetor
	if err := json.Unmarshal(corpo, &resposta); err != nil {
		return "", err
	}

	id := strings.TrimSpace(resposta.Dados)
	if id == "" {
		return "", errors.New("id do setor nao retornado pela api")
	}

	return id, nil
}

func extrairIdSetorPayload(corpo []byte) (string, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(corpo, &payload); err != nil {
		return "", err
	}

	valor, ok := payload["idSetor"]
	if !ok {
		return "", errors.New("id do setor nao informado")
	}

	id, ok := valor.(string)
	if !ok {
		return "", errors.New("id do setor invalido")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return "", errors.New("id do setor invalido")
	}

	return id, nil
}

func gravarTipoSetorConfVision(idSetor string) error {
	idSetor = strings.ToUpper(strings.TrimSpace(idSetor))
	if idSetor == "" {
		return errors.New("id do setor invalido")
	}

	db, erro := conexao.Conectar()
	if erro != nil {
		return fmt.Errorf("falha ao conectar no banco: %w", erro)
	}
	defer db.Close()

	resultado, erro := db.Exec(`
		UPDATE setorAlarme
		SET TipoSetor = ?
		WHERE ID_Setor = ?
	`, setorTipoSetorPadrao, idSetor)
	if erro != nil {
		return erro
	}

	linhas, erro := resultado.RowsAffected()
	if erro != nil {
		return erro
	}
	if linhas == 0 {
		return fmt.Errorf("setor %s nao encontrado para gravar TipoSetor", idSetor)
	}

	return nil
}

func aplicarTipoSetorConfVision(idSetor string) {
	if erro := gravarTipoSetorConfVision(idSetor); erro != nil {
		log.Printf("aviso confvision: TipoSetor nao atualizado para setor %s: %v", idSetor, erro)
	}
}

func obterTipoSetor(idSetor string) (string, error) {
	idSetor = strings.ToUpper(strings.TrimSpace(idSetor))
	if idSetor == "" {
		return "", errors.New("id do setor invalido")
	}

	db, erro := conexao.Conectar()
	if erro != nil {
		return "", fmt.Errorf("falha ao conectar no banco: %w", erro)
	}
	defer db.Close()

	var tipo sql.NullString
	erro = db.QueryRow(`
		SELECT TipoSetor
		FROM setorAlarme
		WHERE ID_Setor = ?
	`, idSetor).Scan(&tipo)
	if erro != nil {
		if errors.Is(erro, sql.ErrNoRows) {
			return "", fmt.Errorf("setor %s nao encontrado", idSetor)
		}
		return "", erro
	}

	return strings.TrimSpace(tipo.String), nil
}

func setorTipoSetorEhSensor(tipo string) bool {
	return strings.ToUpper(strings.TrimSpace(tipo)) == "SENSOR"
}

func setorAlarmeSensor(idSetor string) (bool, error) {
	tipo, erro := obterTipoSetor(idSetor)
	if erro != nil {
		return false, erro
	}
	return setorTipoSetorEhSensor(tipo), nil
}

func bloquearSetorSensor(idSetor string) error {
	sensor, erro := setorAlarmeSensor(idSetor)
	if erro != nil {
		return erro
	}
	if sensor {
		return errors.New("setor do tipo SENSOR nao pode ser alterado, excluido ou habilitado/desabilitado")
	}
	return nil
}

func enriquecerRespostaSetoresComTipoSetor(corpo []byte) ([]byte, error) {
	var resposta map[string]interface{}
	if erro := json.Unmarshal(corpo, &resposta); erro != nil {
		return corpo, erro
	}

	dadosRaw, ok := resposta["dados"]
	if !ok || dadosRaw == nil {
		return corpo, nil
	}

	switch dados := dadosRaw.(type) {
	case []interface{}:
		for _, item := range dados {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			idSetor, _ := m["idSetor"].(string)
			if tipo, erro := obterTipoSetor(idSetor); erro == nil {
				m["tipoSetor"] = tipo
			}
		}
	case map[string]interface{}:
		idSetor, _ := dados["idSetor"].(string)
		if tipo, erro := obterTipoSetor(idSetor); erro == nil {
			dados["tipoSetor"] = tipo
		}
	}

	return json.Marshal(resposta)
}
