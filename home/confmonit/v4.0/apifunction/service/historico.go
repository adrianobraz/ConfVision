package service

import (
	"apifunction/db"
	"database/sql"
	"encoding/json"
)

type LogItem struct {
	IDLog         string          `json:"idLog"`
	DataOperacao  string          `json:"dataOperacao"`
	IDUsuario     string          `json:"idUsuario"`
	Tipo          string          `json:"tipo"`
	IDOrigem      string          `json:"idOrigem"`
	IDDestino     string          `json:"idDestino"`
	VinculoAntigo string          `json:"vinculoAntigo"`
	VinculoNovo   string          `json:"vinculoNovo"`
	Motivo        string          `json:"motivo"`
	Status        string          `json:"status"`
	ErroMsg       string          `json:"erroMsg,omitempty"`
	Preview       json.RawMessage `json:"preview,omitempty"`
}

func ListarHistorico(limite int) ([]LogItem, error) {
	if limite <= 0 || limite > 200 {
		limite = 50
	}
	rows, err := db.Conn.Query(`
		SELECT ID_Log, DataOperacao, ID_Usuario, Tipo, ID_Origem, ID_Destino,
		       VinculoAntigo, VinculoNovo, Motivo, Status, COALESCE(ErroMsg, ''), PreviewJSON
		FROM transferencia_vinculo_log
		ORDER BY DataOperacao DESC
		LIMIT ?
	`, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LogItem
	for rows.Next() {
		var item LogItem
		var preview sql.NullString
		if err := rows.Scan(
			&item.IDLog, &item.DataOperacao, &item.IDUsuario, &item.Tipo,
			&item.IDOrigem, &item.IDDestino, &item.VinculoAntigo, &item.VinculoNovo,
			&item.Motivo, &item.Status, &item.ErroMsg, &preview,
		); err != nil {
			return nil, err
		}
		if preview.Valid && preview.String != "" {
			item.Preview = json.RawMessage(preview.String)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
