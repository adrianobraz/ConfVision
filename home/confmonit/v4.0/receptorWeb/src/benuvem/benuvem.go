package benuvem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var (
// logado bool
)

type BeNuvem struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"access_token"`
	Tipo    string `json:"token_type"`
	Expira  int    `json:"expires_in"`
	Evento  BenuvemEvento
	// parametros internos
	logado      bool
	tokenExpira int
	atualiza    time.Duration
	exibirLog   bool
}

type BenuvemEvento struct {
	CodFranqueado string   `json:"codFranqueado"`
	CodCliente    string   `json:"codCliente"`
	Particao      string   `json:"particao"`
	Canais        []string `json:"canais"`
	Data          string   `json:"data"`
}

type Imagem struct {
	Url      string `json:"url"`
	UrlThumb string `json:"url_thumb"`
	Status   bool   `json:"success"`
	CriadoEm string `json:"created_at"`
	Exepira  int    `json:"expiration"`
}

type Imagens struct {
	Status        string   `json:"success"`
	RtspPrimary   string   `json:"rtsp_primary"`
	RtspSecondary string   `json:"rtsp_secondary"`
	Imagens       []Imagem `json:"images"`
}

// Inicia o serviço benuvem
func Start(
	email string,
	senha string,
	atuliza time.Duration,
	exbirlog bool,
) {
	var be BeNuvem
	be.logado = false       // inicializa logado em falso
	be.atualiza = atuliza   // carraga a propriedade de tempo de atualização
	be.exibirLog = exbirlog // carrega a propriedade
	// Laço que aguarda o login no sistema benuvem
	for !be.logado {

		// Gera um token de acesso benuvem
		if err := be.GerarToken(email, senha); err != nil {
			fmt.Println("Inicializando serviço Benuvem ->", err)
		}

		// Caso de erro no login ele aguarda tempo determinado para
		// tentar logar novamente
		if !be.logado {
			time.Sleep(time.Minute)
		}
	}

	fmt.Println("Inicializando serviço Benuvem -> CONECTADO")

	// Laço que fica atualizando o teken
	for be.logado {
		// time.Sleep(time.Second * 10)
		time.Sleep(time.Minute * be.atualiza)

		if erro := be.AtualizarToken(); erro != nil {
			if be.exibirLog {
				fmt.Println("Atualizando token benuvem ->", erro)
				be.logado = false
			}
		}

		fmt.Println("Atualizando token benuvem -> OK")
	}

	// Reinicia o serviço
	fmt.Println("Reiniciando modulo Benuvem")
	go Start(email, senha, atuliza, exbirlog)
}

// Loga no sistema benuvem
func (be *BeNuvem) GerarToken(email, senha string) error {
	be.logado = false
	url := "https://app.benuvem.com.br/api/v1/auth/login"
	method := "POST"

	// Cria um buffer de bytes
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("email", email)
	_ = writer.WriteField("password", senha)
	if erro := writer.Close(); erro != nil {
		return erro
	}

	client := &http.Client{}
	req, erro := http.NewRequest(method, url, payload)
	if erro != nil {
		return erro
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, erro := client.Do(req)
	if erro != nil {
		return erro
	}
	defer res.Body.Close()

	// Verifica se nao retornou um codigo de erro
	if res.StatusCode >= 400 {
		println("benuvem não autorizado")
		return nil
	}

	body, erro := io.ReadAll(res.Body)
	if erro != nil {
		return erro
	}

	if erro := json.Unmarshal(body, &be); erro != nil {
		return erro
	}

	// Carrega as variaveis com o dados do token
	be.tokenExpira = be.Expira - 600

	be.logado = true
	return nil
}

// Atualiza o teken de acesso
func (be *BeNuvem) AtualizarToken() error {

	url := "https://app.benuvem.com.br/api/v1/auth/refresh"
	method := "POST"

	client := &http.Client{}
	req, erro := http.NewRequest(method, url, nil)
	if erro != nil {
		be.logado = false
		return erro
	}

	req.Header.Add("Authorization", be.Tipo+" "+be.Token)
	res, erro := client.Do(req)
	if erro != nil {
		be.logado = false
		return erro
	}
	defer res.Body.Close()

	body, erro := io.ReadAll(res.Body)
	if erro != nil {
		be.logado = false
		return erro
	}

	if res.StatusCode == 200 {
		if erro := json.Unmarshal(body, &be); erro != nil {
			be.logado = false
			return erro
		}

		if be.Status == "Authorization Token not found" {
			return errors.New("não esta logado")
		}

		// Carrega as variaveis com o dados do token
		be.tokenExpira = be.Expira - 600

		be.logado = true
		return nil
	}
	return errors.New("erro ao requistar atualização")
}

// Gera um evento de gravação no sistema benuvem
func (be *BeNuvem) GerarEvento() error {
	if be.logado {
		url := "https://app.benuvem.com.br/api/v1/events/new"
		method := "POST"

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("client_code", be.Evento.CodCliente)
		_ = writer.WriteField("company_code", be.Evento.CodFranqueado)
		_ = writer.WriteField("partition", be.Evento.Particao)

		for _, canal := range be.Evento.Canais {
			_ = writer.WriteField("channels[]", canal)
		}

		erro := writer.Close()
		if erro != nil {
			return erro
		}

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			return erro
		}

		req.Header.Add("Authorization", be.Tipo+" "+be.Token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, err := client.Do(req)
		if err != nil {
			return erro
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return erro
		}

		var ret struct {
			Status bool   `json:"success"`
			Msg    string `json:"message"`
		}
		if erro := json.Unmarshal(body, &ret); erro != nil {
			return erro
		}

		if ret.Status {
			return nil
		} else {
			return errors.New(ret.Msg)
		}
	} else {
		return errors.New("sistema não logado no benuvem")
	}
}

//=============================================================================

// Efetua o logout no sistema benuvem
func (be *BeNuvem) LogOff(ok *bool) error {
	*ok = false
	// Caso logado seja false ele reinicia o modulo
	if be.logado {
		url := "https://app.benuvem.com.br/api/v1/auth/logout"
		method := "POST"

		client := &http.Client{}
		req, erro := http.NewRequest(method, url, nil)

		if erro != nil {
			return erro
		}
		res, erro := client.Do(req)
		if erro != nil {
			return erro
		}
		defer res.Body.Close()

		// body, erro := io.ReadAll(res.Body)
		// if erro != nil {
		// 	return erro
		// }
		*ok = true
	} else {
		return errors.New("sistema não logado no benuvem")
	}
	return nil
}

// Finaliza um evento de gravação no sistema benuvem
func (be *BeNuvem) FinalizarEvento() error {

	if be.logado {
		url := "https://app.benuvem.com.br/api/v1/events/close"
		method := "POST"

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("client_code", be.Evento.CodCliente)
		_ = writer.WriteField("company_code", be.Evento.CodFranqueado)
		_ = writer.WriteField("partition", be.Evento.Particao)

		for _, canal := range be.Evento.Canais {
			_ = writer.WriteField("channels[]", canal)
		}

		erro := writer.Close()
		if erro != nil {
			return erro
		}

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			return erro
		}

		req.Header.Add("Authorization", be.Tipo+" "+be.Token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, err := client.Do(req)
		if err != nil {
			return erro
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return erro
		}

		var ret struct {
			Status bool   `json:"success"`
			Msg    string `json:"message"`
		}
		if erro := json.Unmarshal(body, &ret); erro != nil {
			return erro
		}

		if ret.Status {
			return nil
		} else {
			return errors.New(ret.Msg)
		}
	} else {
		return errors.New("sistema não logado no benuvem")
	}
}

// codCliente, particao, data, canal string
func (be *BeNuvem) GetRtpsImagens(img *Imagem) error {

	// Caso logado seja false ele reinicia o modulo
	if be.logado {
		url := "https://app.benuvem.com.br/api/v1/cameras/get-rtsp-images/"
		method := "POST"

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("company_code", be.Evento.CodFranqueado)
		_ = writer.WriteField("client_code", be.Evento.CodCliente)
		_ = writer.WriteField("partition", be.Evento.Particao)
		_ = writer.WriteField("date", be.Evento.Data) //"2020-10-07 18:47:46")
		_ = writer.WriteField("channel", be.Evento.Canais[0])
		erro := writer.Close()
		if erro != nil {
			return erro
		}

		client := &http.Client{}
		req, erro := http.NewRequest(method, url, payload)

		if erro != nil {
			return erro
		}

		req.Header.Add("Authorization", be.Tipo+" "+be.Token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, err := client.Do(req)
		if err != nil {
			return erro
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return erro
		}

		if erro := json.Unmarshal(body, &img); erro != nil {
			return erro
		}
		if img.Status {
			return nil
		}
		return errors.New("erro ao obter rtps ou imagens")
	} else {
		return errors.New("sistema não logado no benuvem")
	}
}

// Obtem as imagens ao vivo ou gravada
func (be *BeNuvem) GetUrlCamera(url *string, expira *int) error {

	// Caso logado seja false ele reinicia o modulo
	if be.logado {

		reqUrl := "https://app.benuvem.com.br/api/v1/cameras/show"
		method := "POST"
		var img Imagem

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("company_code", be.Evento.CodFranqueado)
		_ = writer.WriteField("client_code", be.Evento.CodCliente)
		_ = writer.WriteField("partition", be.Evento.Particao)
		if be.Evento.Data != "" {
			_ = writer.WriteField("date", be.Evento.Data)
		} else {
			// Para exibir ao vivo show_records_popup = false
			_ = writer.WriteField("show_records_popup", "false")
		}

		for _, canal := range be.Evento.Canais {
			_ = writer.WriteField("channels[]", canal)
		}

		erro := writer.Close()
		if erro != nil {
			return erro
		}

		client := &http.Client{}
		req, err := http.NewRequest(method, reqUrl, payload)
		if err != nil {
			return erro
		}

		req.Header.Add("Authorization", be.Tipo+" "+be.Token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, err := client.Do(req)
		if err != nil {
			return erro
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return erro
		}

		if erro := json.Unmarshal(body, &img); erro != nil {
			return erro
		}

		if img.Status {
			*url = img.Url
			*expira = img.Exepira
			return nil
		} else {
			return nil
		}
	} else {
		return errors.New("sistema não logado no benuvem")
	}
}

// Verificar o modo de funcionamento
func (be *BeNuvem) GetStreamImagems() {

	// Caso logado seja false ele reinicia o modulo
	if be.logado {

		url := "https://app.benuvem.com.br/api/v1/cameras/get-stream-images/"
		method := "POST"

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("company_code", be.Evento.CodFranqueado) //"1"
		_ = writer.WriteField("client_code", be.Evento.CodCliente)
		_ = writer.WriteField("partition", be.Evento.Particao)
		_ = writer.WriteField("date", "")
		_ = writer.WriteField("channel", be.Evento.Canais[0])
		err := writer.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			fmt.Println(err)
			return
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(body))
	}
}
