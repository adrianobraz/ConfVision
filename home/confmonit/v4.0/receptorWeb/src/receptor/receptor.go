package receptor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"receptorWeb/src/auxiliar"
	"receptorWeb/src/config"
	evento "receptorWeb/src/eventos"
	"receptorWeb/src/respApp"
)

type Receptor struct {
	IdDispositivo  string `json:"idDispositivo"`
	IdFranqueado   string `json:"idFranqueado"` // Recebe
	Evento         string `json:"evento"`       // Recebe
	SenhaRecSerial string `json:"senha"`        // Recebe
	// IdCentral      string
	// IdCliente      string
	// Conta          string
	// Benuvem        string
	// NomeCliente    string
	// Msg            string
	// Ativo          string
}

// Funcao que trata o evento
func ReceberEvento(w http.ResponseWriter, r *http.Request) {

	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		fmt.Println("Informando Erro Receptor Web =>", erro)
		return
	}

	var d Receptor

	if erro := json.Unmarshal(corpo, &d); erro != nil {
		fmt.Println("Informando Erro Receptor Web =>", erro)
		return
	}

	if d.SenhaRecSerial != config.SenhaWeb {
		fmt.Println("Informando Erro Receptor Web =>", "Senha invalida")
		return
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		fmt.Println("Informando Erro Receptor Web =>", "Senha invalida")
		return
	}

	if erro := evento.RecebeEvento(db, d.IdFranqueado, d.Evento); erro != nil {
		fmt.Println("Informando Erro Receptor Web =>", erro.Error())
		return
	}

	respApp.RespOK(w)
}
