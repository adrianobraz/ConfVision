package benuvemV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/V4/config"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var logado bool

type BeNuvem struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"access_token"`
	Tipo    string `json:"token_type"`
	Expira  int    `json:"expires_in"`
	Logado  bool
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
func Start() {
	var be BeNuvem

	fmt.Print("Iniciando serviço Benuvem")
	logado = false
	// Laço que aguarda o login no sistema benuvem
	for !logado {

		// Gera um token de acesso benuvem
		if err := be.GerarToken(&logado); err != nil {
			fmt.Println("mBenuvem ->", err)
		}

		if !logado {
			time.Sleep(time.Minute)
		}

	}
	fmt.Println(" -> CONECTADO")

	// Laço que fica atualizando o teken
	for logado {
		//time.Sleep(time.Second * 10)
		time.Sleep(time.Hour)

		fmt.Println("Atualizando token benuvem")

		if erro := be.AtualizarToken(&logado); erro != nil {
			if config.Benuven.ExibirLog {
				fmt.Println("mBenuvem ->", erro)
				logado = false
			}
		}

	}

	// Reinicia o serviço
	fmt.Println("Reiniciando modulo Benuvem")
	go Start()
}

// Loga no sistema benuvem
func (be *BeNuvem) GerarToken(logado *bool) error {
	*logado = false
	url := "https://app.benuvem.com.br/api/v1/auth/login"
	method := "POST"

	// Cria um buffer de bytes
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("email", config.Benuven.Email)
	_ = writer.WriteField("password", config.Benuven.Senha)
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
		println("não autorizado")
		return nil
	}

	body, erro := io.ReadAll(res.Body)
	if erro != nil {
		return erro
	}

	var b BeNuvem
	if erro := json.Unmarshal(body, &b); erro != nil {
		return erro
	}

	// Carrega as variaveis com o dados do token
	config.Benuven.Token = b.Token
	config.Benuven.TokenTipo = b.Tipo
	config.Benuven.TokenExpira = b.Expira - 600

	*logado = true
	return nil
}

// Atualiza o teken de acesso
func (be *BeNuvem) AtualizarToken(ok *bool) error {
	*ok = false

	url := "https://app.benuvem.com.br/api/v1/auth/refresh"
	method := "POST"

	client := &http.Client{}
	req, erro := http.NewRequest(method, url, nil)

	if erro != nil {
		return erro
	}
	req.Header.Add("Authorization", config.Benuven.TokenTipo+" "+config.Benuven.Token)
	res, erro := client.Do(req)
	if erro != nil {
		return erro
	}
	defer res.Body.Close()

	body, erro := io.ReadAll(res.Body)
	if erro != nil {
		return erro
	}

	if res.StatusCode == 200 {
		var be BeNuvem
		if erro := json.Unmarshal(body, &be); erro != nil {
			return erro
		}

		if be.Status == "Authorization Token not found" {
			return errors.New("não esta logado")
		}

		// Carrega as variaveis com o dados do token
		config.Benuven.Token = be.Token
		config.Benuven.TokenTipo = be.Tipo
		config.Benuven.TokenExpira = be.Expira - 600
		*ok = true
		return nil
	}
	return errors.New("erro ao requistar atualização")
}

// Efetua o logout no sistema benuvem
func (be *BeNuvem) LogOff(ok *bool) error {
	*ok = false
	// Caso logado seja false ele reinicia o modulo
	if !logado {
		return errors.New("sistema não logado no benuvem")
	}

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
	return nil
}

// Gera um evento de gravação no sistema benuvem
func (be *BeNuvem) GerarEvento(param BenuvemEvento) error {

	// Caso logado seja false ele reinicia o modulo
	if !logado {
		return errors.New("sistema não logado no benuvem")
	}

	url := "https://app.benuvem.com.br/api/v1/events/new"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("client_code", param.CodCliente)
	_ = writer.WriteField("company_code", param.CodFranqueado)
	_ = writer.WriteField("partition", param.Particao)

	for _, canal := range param.Canais {
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

	req.Header.Add("Authorization", config.Benuven.TokenTipo+" "+config.Benuven.Token)
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
}

// Finaliza um evento de gravação no sistema benuvem
func (be *BeNuvem) FinalizarEvento(param BenuvemEvento) error {

	// Caso logado seja false ele reinicia o modulo
	if !logado {
		return errors.New("sistema não logado no benuvem")
	}

	url := "https://app.benuvem.com.br/api/v1/events/close"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("client_code", param.CodCliente)
	_ = writer.WriteField("company_code", param.CodFranqueado)
	_ = writer.WriteField("partition", param.Particao)

	for _, canal := range param.Canais {
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

	req.Header.Add("Authorization", config.Benuven.TokenTipo+" "+config.Benuven.Token)
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
}

func (be *BeNuvem) GetRtpsImagens(codCliente, particao, data, canal string) (Imagens, error) {

	var img Imagens

	// Caso logado seja false ele reinicia o modulo
	if !logado {
		return img, errors.New("sistema não logado no benuvem")
	}

	url := "https://app.benuvem.com.br/api/v1/cameras/get-rtsp-images/"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("company_code", "1")
	_ = writer.WriteField("client_code", codCliente)
	_ = writer.WriteField("partition", particao)
	_ = writer.WriteField("date", data) //"2020-10-07 18:47:46")
	_ = writer.WriteField("channel", canal)
	erro := writer.Close()
	if erro != nil {
		return img, erro
	}

	client := &http.Client{}
	req, erro := http.NewRequest(method, url, payload)

	if erro != nil {
		return img, erro
	}

	req.Header.Add("Authorization", config.Benuven.TokenTipo+" "+config.Benuven.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return img, erro
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return img, erro
	}

	if erro := json.Unmarshal(body, &img); erro != nil {
		return img, erro
	}
	if img.Status == "true" {
		return img, nil
	}
	return img, errors.New("erro ao obter rtps ou imagens")
}

// Obtem as imagens ao vivo ou gravada
func (be *BeNuvem) GetUrlCamera(param BenuvemEvento, url *string, expira *int) (erro error) {

	// Caso logado seja false ele reinicia o modulo
	if !logado {

		return errors.New("sistema não logado no benuvem")
	}

	reqUrl := "https://app.benuvem.com.br/api/v1/cameras/show"
	method := "POST"
	var img Imagem

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("company_code", param.CodFranqueado)
	_ = writer.WriteField("client_code", param.CodCliente)
	_ = writer.WriteField("partition", param.Particao)
	if param.Data != "" {
		_ = writer.WriteField("date", param.Data)
	} else {
		// Para exibir ao vivo show_records_popup = false
		_ = writer.WriteField("show_records_popup", "false")
	}

	for _, canal := range param.Canais {
		_ = writer.WriteField("channels[]", canal)
	}

	erro = writer.Close()
	if erro != nil {
		return erro
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, reqUrl, payload)
	if err != nil {
		return erro
	}

	req.Header.Add("Authorization", config.Benuven.TokenTipo+" "+config.Benuven.Token)
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
}

// Verificar o modo de funcionamento
func (be *BeNuvem) GetStreamImagems(codCliente, particao, data, canal string) {

	// Caso logado seja false ele reinicia o modulo
	if !logado {
		return
	}

	url := "https://app.benuvem.com.br/api/v1/cameras/get-stream-images/"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("company_code", "1")
	_ = writer.WriteField("client_code", codCliente)
	_ = writer.WriteField("partition", particao)
	_ = writer.WriteField("date", "")
	_ = writer.WriteField("channel", canal)
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

func (be *BeNuvem) GetCodBenuvemByIdDispositivo(idDisp string, codBenuvem *string) error {
	if idDisp == "" {
		return errors.New("um id de dispositivo deve ser ")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT franqueado.CodBenuvem
		FROM dispositivo 

		LEF JOIN cliente
		ON franqueado.ID_Franqueado = cliente.ID_Franqueado
		
		LEF JOIN franqueado
		ON cliente.ID_Cliente = dispositivo.ID_Cliente

		WHERE dispositivo.ID_Dispositivo = ?
	
	`, idDisp)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&codBenuvem); err != nil {
			return err
		}
		return nil
	}
	return errors.New("codigo não encontrado na base de dados")
}
