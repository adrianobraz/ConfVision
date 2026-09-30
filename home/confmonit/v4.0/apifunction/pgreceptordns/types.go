package pgreceptordns

import "time"

type Registro struct {
	IDFranqueado string    `json:"id_franqueado"`
	Subdominio   string    `json:"subdominio"`
	FQDN         string    `json:"fqdn"`
	Zona         string    `json:"zona"`
	IPDestino    string    `json:"ip_destino"`
	CFRecordID   string    `json:"cf_record_id,omitempty"`
	Status       string    `json:"status"`
	Erro         string    `json:"erro,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type EndpointFabricante struct {
	IDFabricante string `json:"id_fabricante"`
	Nome         string `json:"nome"`
	Modulo       string `json:"modulo"`
	Host         string `json:"host"`
	Porta        string `json:"porta"`
	Endereco     string `json:"endereco"`
}

type Disponibilidade struct {
	Subdominio  string `json:"subdominio"`
	FQDN        string `json:"fqdn"`
	Disponivel  bool   `json:"disponivel"`
	Motivo      string `json:"motivo,omitempty"`
}
