package auxiliar

import (
	"errors"
	"strings"
)

// IDCentralPadrao e o ID_Central da empresa legada.
// A coluna representante.IDCentralUUID (e similares) guarda o ID_Central da central,
// nao o campo central.IDCentralUUID.
const IDCentralPadrao = "CENTRAL"

// IDCentralUUIDPadrao retorna o identificador padrao da central legada (ID_Central).
func IDCentralUUIDPadrao() (string, error) {
	return IDCentralPadrao, nil
}

// GarantirIDCentralUUID preenche se vazio com a central legada (ID_Central = CENTRAL).
// O valor retornado/armazenado e sempre o ID_Central da tabela central.
func GarantirIDCentralUUID(atual string) string {
	atual = strings.TrimSpace(atual)
	if atual != "" {
		return atual
	}
	return IDCentralPadrao
}

// ExigirFiltroCentralUUID valida isolamento multi-tenant nas listagens.
// listarTodas=true (break-glass) dispensa o id; caso contrario e obrigatorio.
// O parametro "uuid" na pratica e o ID_Central da central.
func ExigirFiltroCentralUUID(uuid string, listarTodas bool) error {
	if listarTodas {
		return nil
	}
	if strings.TrimSpace(uuid) == "" {
		return errors.New("idCentralUUID obrigatorio")
	}
	return nil
}
