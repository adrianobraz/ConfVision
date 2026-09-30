package pgreceptordiag

import "time"

type FabricantePorta struct {
	IDFabricante string `json:"idFabricante"`
	Nome         string `json:"nome"`
	Modulo       string `json:"modulo"`
	Porta        string `json:"porta"`
}

type ErroConexaoRow struct {
	ID              string `json:"id"`
	Codigo          string `json:"codigo"`
	Serial          string `json:"serial"`
	Imei            string `json:"imei"`
	IPRemoto        string `json:"ipRemoto"`
	Modelo          string `json:"modelo"`
	Versao          string `json:"versao"`
	Protocolo       string `json:"protocolo"`
	Fabricante      string `json:"fabricante"`
	Conta           string `json:"conta"`
	IDFranqueado    string `json:"idFranqueado"`
	DataTentativa   string `json:"dataTentativa"`
	MatchTipo       string `json:"matchTipo,omitempty"`
	MatchDistancia  int    `json:"matchDistancia,omitempty"`
	CadastroCodigo  string `json:"cadastroCodigo,omitempty"`
	MatchObservacao string `json:"matchObservacao,omitempty"`
	CodigoMatch     string `json:"codigoMatch,omitempty"`
	ContaMatch      bool   `json:"contaMatch,omitempty"`
}

type FranqueadoRow struct {
	IDFranqueado string `json:"idFranqueado"`
	Nome         string `json:"nome"`
}

type DispositivoInfo struct {
	IDDispositivo  string `json:"idDispositivo"`
	Conta          string `json:"conta"`
	IdFisico1      string `json:"idFisico1"`
	IdFisico2      string `json:"idFisico2"`
	IDCliente      string `json:"idCliente"`
	NomeCliente    string `json:"nomeCliente"`
	IDFranqueado   string `json:"idFranqueado"`
	NomeFranqueado string `json:"nomeFranqueado"`
	DataUltimo     string `json:"dataUltimoEvento"`
	CodigoUltimo   string `json:"codigoUltimoEvento"`
}

type JournalEvent struct {
	DataHora string `json:"dataHora"`
	Driver   string `json:"driver"`
	Conta    string `json:"conta"`
	IMEI     string `json:"imei"`
	Evento   string `json:"evento"`
	Linha    string `json:"linha,omitempty"`
}

type Presenca struct {
	IDFranqueado  string     `json:"idFranqueado"`
	Fabricante    string     `json:"fabricante"`
	Modulo        string     `json:"modulo"`
	Conta         string     `json:"conta"`
	IDDispositivo string     `json:"idDispositivo"`
	IPRemoto      string     `json:"ipRemoto"`
	UltimoEvento  string     `json:"ultimoEvento"`
	UltimoSinal   *time.Time `json:"ultimoSinalAt,omitempty"`
}

type Check struct {
	ID      string `json:"id"`
	OK      bool   `json:"ok"`
	Alerta  bool   `json:"alerta"`
	Titulo  string `json:"titulo"`
	Detalhe string `json:"detalhe"`
	Comando string `json:"comando,omitempty"`
	Saida   string `json:"saida,omitempty"`
}

type ReqVerificar struct {
	IDFranqueado string `json:"id_franqueado"`
	Modulo       string `json:"modulo"`
	Fabricante   string `json:"fabricante"`
	Conta        string `json:"conta"`
	Codigo       string `json:"codigo"`
	IPTecnico    string `json:"ip_tecnico"`
	Lab          bool   `json:"lab"`
}

type Resultado struct {
	Status            string           `json:"status"`
	Conclusao         string           `json:"conclusao"`
	Acao              string           `json:"acao"`
	Modulo            string           `json:"modulo"`
	Fabricante        string           `json:"fabricante"`
	Porta             string           `json:"porta"`
	HostPainel        string           `json:"hostPainel"`
	EnderecoPainel    string           `json:"enderecoPainel"`
	IPTecnico         string           `json:"ipTecnico"`
	IPCentral         string           `json:"ipCentral"`
	MesmaRede         *bool            `json:"mesmaRede,omitempty"`
	Checks            []Check          `json:"checks"`
	JournalEventos    []JournalEvent   `json:"journalEventos,omitempty"`
	ErrosConexao      []ErroConexaoRow `json:"errosConexao"`
	Dispositivo       *DispositivoInfo   `json:"dispositivo,omitempty"`
	Dispositivos      []DispositivoInfo  `json:"dispositivos,omitempty"`
	Presenca          *Presenca          `json:"presenca,omitempty"`
	ModoReceptor      bool               `json:"modoReceptor,omitempty"`
	ModoDiag          string             `json:"modoDiag,omitempty"`
	ContaDigitada     string             `json:"contaDigitada,omitempty"`
	CodigoDigitado    string             `json:"codigoDigitado,omitempty"`
}

type AgentRunResult struct {
	ListenOK       bool     `json:"listenOk"`
	ListenCmd      string   `json:"listenCmd"`
	ListenOut      string   `json:"listenOut"`
	Established    []string `json:"established"`
	EstablishedCmd string   `json:"establishedCmd"`
	EstablishedOut string   `json:"establishedOut"`
	JournalLines   []string `json:"journalLines"`
	JournalCmd     string   `json:"journalCmd"`
	ProtocolOK     bool     `json:"protocolOk"`
	ProtocolCmd    string   `json:"protocolCmd"`
	ProtocolOut    string   `json:"protocolOut"`
}

type ReqPresenca struct {
	IDFranqueado  string `json:"id_franqueado"`
	Fabricante    string `json:"fabricante"`
	Modulo        string `json:"modulo"`
	Conta         string `json:"conta"`
	IDDispositivo string `json:"id_dispositivo"`
	IPRemoto      string `json:"ip_remoto"`
	UltimoEvento  string `json:"ultimo_evento"`
}
