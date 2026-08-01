package tipos

// MapaAmbiente — tabela Xano #104
type MapaAmbiente struct {
	Id             int    `json:"id"`
	IdCliente      string `json:"idCliente"`
	IdFranqueado   string `json:"idFranqueado"`
	NomeCliente    string `json:"nomeCliente"`
	NomeFranqueado string `json:"nomeFranqueado"`
	Descricao      string `json:"descricao"`
	ImagemUrl      string `json:"imagem_url"`
	Ordem          int    `json:"ordem"`
	Ativo          *bool  `json:"ativo"`
}

// MapaSetor — tabela Xano #105 (eixo Y: poxY no banco)
type MapaSetor struct {
	Id              int     `json:"id"`
	MapaAmbienteId  int     `json:"mapa_ambiente_id"`
	IdSetor         string  `json:"idSetor"`
	IdDispositivo   string  `json:"idDispositivo"`
	IdCliente       string  `json:"idCliente"`
	SetorNome       string  `json:"setornome"`
	DispositivoNome string  `json:"dispositivoNome"`
	ClienteNome     string  `json:"clienteNome"`
	PosX            float64 `json:"posX"`
	PoxY            float64 `json:"poxY"`
	Icone           string  `json:"icone"`
	Label           string  `json:"label"`
	Descricao       string  `json:"descricao"`
	TipoSetor       string  `json:"tipoSetor,omitempty"`
	Camera          string  `json:"camera,omitempty"`
	Numero          string  `json:"numero,omitempty"`
	Particao        string  `json:"particao,omitempty"`
	IdFranqueado    string  `json:"idFranqueado,omitempty"`
}
