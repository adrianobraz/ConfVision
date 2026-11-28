// Pacote model é o contrato JSON do monitor.
// Origem separa número medido, conta feita em memória, projeção e ausência.
package model

import "time"

const (
	OrigemMedido       = "medido"
	OrigemCalculado    = "calculado"
	OrigemProjetado    = "projetado"
	OrigemIndisponivel = "indisponivel"
)

// Reading é um valor com a origem explícita.
// Valor fica ausente quando a origem é indisponivel. Zero não substitui ausência.
type Reading struct {
	Nome    string   `json:"nome"`
	Origem  string   `json:"origem"`
	Valor   *float64 `json:"valor,omitempty"`
	Unidade string   `json:"unidade,omitempty"`
	Texto   string   `json:"texto,omitempty"`
	Motivo  string   `json:"motivo,omitempty"`
}

func Measured(nome, unidade string, valor float64) Reading {
	v := valor
	return Reading{Nome: nome, Origem: OrigemMedido, Valor: &v, Unidade: unidade}
}

func Calculated(nome, unidade string, valor float64) Reading {
	v := valor
	return Reading{Nome: nome, Origem: OrigemCalculado, Valor: &v, Unidade: unidade}
}

func Projected(nome, unidade string, valor float64) Reading {
	v := valor
	return Reading{Nome: nome, Origem: OrigemProjetado, Valor: &v, Unidade: unidade}
}

func Unavailable(nome, motivo string) Reading {
	return Reading{Nome: nome, Origem: OrigemIndisponivel, Motivo: motivo}
}

func Text(nome, origem, texto string) Reading {
	return Reading{Nome: nome, Origem: origem, Texto: texto}
}

type Server struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

type Container struct {
	ID           string  `json:"id,omitempty"`
	Nome         string  `json:"nome"`
	Estado       Reading `json:"estado"`
	CPU          Reading `json:"cpu"`
	Memoria      Reading `json:"memoria"`
	LimiteCPU    Reading `json:"limite_cpu"`
	LimiteMem    Reading `json:"limite_memoria"`
	Uptime       Reading `json:"uptime"`
	Reinicios    Reading `json:"reinicios"`
	RedeRX       Reading `json:"rede_rx"`
	RedeTX       Reading `json:"rede_tx"`
	Papel        string  `json:"papel,omitempty"`
	IncluiFilhos bool    `json:"inclui_processos_filhos"`
}

type Process struct {
	PID       int     `json:"pid,omitempty"`
	Nome      string  `json:"nome"`
	Comando   string  `json:"comando,omitempty"`
	Container string  `json:"container,omitempty"`
	CPU       Reading `json:"cpu"`
	Memoria   Reading `json:"memoria"`
	// IncluidoNoContainer evita somar de novo o que o cgroup do container já contém.
	IncluidoNoContainer bool `json:"incluido_no_container"`
}

type Camera struct {
	VisCameraID int                `json:"vis_camera_id"`
	Campos      map[string]Reading `json:"campos"`
}

type Alert struct {
	Chave   string    `json:"chave"`
	Estado  string    `json:"estado"`
	Nome    string    `json:"nome"`
	Origem  string    `json:"origem"`
	Detalhe string    `json:"detalhe,omitempty"`
	Desde   time.Time `json:"desde"`
}

type Capacity struct {
	CPUMedida      Reading `json:"cpu_medida"`
	CPUConfigurada Reading `json:"cpu_configurada"`
	CPUProjetada   Reading `json:"cpu_projetada"`
	RAMMedida      Reading `json:"ram_medida"`
	RAMConfigurada Reading `json:"ram_configurada"`
	RAMProjetada   Reading `json:"ram_projetada"`
	MargemCPU      Reading `json:"margem_cpu"`
	MargemRAM      Reading `json:"margem_ram"`
}

// Snapshot é a coleta atual. Não é histórico persistido.
type Snapshot struct {
	ColetadoEm          time.Time   `json:"coletado_em"`
	Servidor            Server      `json:"servidor"`
	Host                []Reading   `json:"host"`
	Containers          []Container `json:"containers"`
	Servicos            []Container `json:"servicos"`
	Processos           []Process   `json:"processos"`
	Cameras             []Camera    `json:"cameras"`
	Worker              []Reading   `json:"worker"`
	MediaMTX            []Reading   `json:"mediamtx"`
	Redis               []Reading   `json:"redis"`
	Alertas             []Alert     `json:"alertas"`
	Capacidade          Capacity    `json:"capacidade"`
	Coleta              []Reading   `json:"coleta"`
	HistoricoPersistido bool        `json:"historico_persistido"`
}

type HistoryResponse struct {
	Persistido   bool       `json:"persistido"`
	Motivo       string     `json:"motivo"`
	AmostraAtual *Snapshot  `json:"amostra_atual,omitempty"`
	Series       []struct{} `json:"series"`
}
