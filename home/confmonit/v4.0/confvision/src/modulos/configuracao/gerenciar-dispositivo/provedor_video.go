package gerenciarDispositivo

import (
	"confvision/src/conexao"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const provedorVideoPadrao = "CONFVISION"

type respostaInserirDispositivo struct {
	Dados string `json:"dados"`
}

func normalizarProvedorVideo(valor string) string {
	valor = strings.ToUpper(strings.TrimSpace(valor))
	if valor == "" {
		return provedorVideoPadrao
	}
	return valor
}

func extrairProvedorVideo(corpo []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(corpo, &payload); err != nil {
		return provedorVideoPadrao
	}

	valor, ok := payload["provedorVideo"]
	if !ok {
		return provedorVideoPadrao
	}

	texto, ok := valor.(string)
	if !ok {
		return provedorVideoPadrao
	}

	return normalizarProvedorVideo(texto)
}

func removerProvedorVideo(corpo []byte) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(corpo, &payload); err != nil {
		return corpo, err
	}

	delete(payload, "provedorVideo")
	return json.Marshal(payload)
}

func extrairIdDispositivoInserido(corpo []byte) (string, error) {
	var resposta respostaInserirDispositivo
	if err := json.Unmarshal(corpo, &resposta); err != nil {
		return "", err
	}

	id := strings.TrimSpace(resposta.Dados)
	if id == "" {
		return "", errors.New("id do dispositivo nao retornado pela api")
	}

	return id, nil
}

func gravarProvedorVideo(idDispositivo, provedorVideo string) error {
	idDispositivo = strings.TrimSpace(idDispositivo)
	if idDispositivo == "" {
		return errors.New("id do dispositivo invalido")
	}

	db, erro := conexao.Conectar()
	if erro != nil {
		return fmt.Errorf("falha ao conectar no banco: %w", erro)
	}
	defer db.Close()

	resultado, erro := db.Exec(`
		UPDATE dispositivo
		SET ProvedorVideo = ?
		WHERE ID_Dispositivo = ?
	`, normalizarProvedorVideo(provedorVideo), idDispositivo)
	if erro != nil {
		return erro
	}

	linhas, erro := resultado.RowsAffected()
	if erro != nil {
		return erro
	}
	if linhas == 0 {
		return errors.New("dispositivo nao encontrado para gravar ProvedorVideo")
	}

	return nil
}
