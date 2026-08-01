package dispositivoV4

import (
	connV4 "api/src/V4/conexao"
	contactidV4 "api/src/V4/modulos/contactid"
	emailEventoV4 "api/src/V4/modulos/emailEvento"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Dispositivo struct {
	ID_Franqueado  string `json:"idFranqueado"`
	NomeFranqueado string `json:"nomeFranqueado"`

	ID_Cliente  string `json:"idCliente"`
	NomeCliente string `json:"nomeCliente"`

	ID_Dispositivo        string `json:"idDispositivo"`
	ID_Fabricante         string `json:"idFabricante"`
	ID_Modelo             string `json:"idModelo"`
	Tipo                  string `json:"tipo"`
	Nome                  string `json:"nome"`
	Particao              string `json:"particao"`
	Conta                 string `json:"conta"`
	IdFisico1             string `json:"idFisico1"`
	IdFisico2             string `json:"idFisico2"`
	KeepAlive             string `json:"keepAlive"`
	Armado                string `json:"armado"`
	DataArmado            string `json:"dataArmado"`
	MsgAtendente          string `json:"msgAtendente"`
	DataUltimoEvento      string `json:"dataUltimoEvento"`
	CodgioUltimoEvento    string `json:"codigoUltimoEvento"`
	DescricaoUltimoEvento string `json:"descricaoUltimoEvento"`
	Manutencao            string `json:"manutencao"`
	Senha                 string `json:"senha"`
	SenhaVerbal           string `json:"senhaVerbal"`
	ContraSenhaVerbal     string `json:"contraSenhaVerbal"`
	Ativo                 string `json:"ativo"`
	DataCadastro          string `json:"dataCadastro"`
}

type SDispositivo struct {
	ID_Franqueado  sql.NullString
	NomeFranqueado sql.NullString

	ID_Cliente  sql.NullString
	NomeCliente sql.NullString

	ID_Dispositivo        sql.NullString
	ID_Fabricante         sql.NullString
	ID_Modelo             sql.NullString
	Tipo                  sql.NullString
	Nome                  sql.NullString
	Particao              sql.NullString
	Conta                 sql.NullString
	IdFisico1             sql.NullString
	IdFisico2             sql.NullString
	KeepAlive             sql.NullString
	Armado                sql.NullString
	DataArmado            sql.NullTime
	MsgAtendente          sql.NullString
	DataUltimoEvento      sql.NullTime
	CodgioUltimoEvento    sql.NullString
	DescricaoUltimoEvento sql.NullString
	Manutencao            sql.NullTime
	Senha                 sql.NullString
	SenhaVerbal           sql.NullString
	ContraSenhaVerbal     sql.NullString
	DataCadastro          sql.NullTime
}

func (d *Dispositivo) Insere() error {
	if d.ID_Dispositivo == "" {
		d.ID_Dispositivo = auxiliar.GeradorDeId()
	}

	if d.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	if d.Nome == "" {
		return errors.New("um nome para o dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO dispositivo(
			dispositivo.ID_Dispositivo, 
			dispositivo.ID_Cliente, 
			dispositivo.ID_Fabricante, 
			dispositivo.ID_Modelo, 
			dispositivo.Tipo, 
			dispositivo.Nome, 
			dispositivo.Particao, 
			dispositivo.Conta, 
			dispositivo.IdFisico1, 
			dispositivo.IdFisico2, 
			dispositivo.KeepAlive, 
			dispositivo.MsgAtendente, 		
			dispositivo.Senha,
			dispositivo.SenhaVerbal, 
			dispositivo.ContraSenhaVerbal 
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(d.ID_Dispositivo),
		strings.ToUpper(d.ID_Cliente),
		strings.ToUpper(d.ID_Fabricante),
		strings.ToUpper(d.ID_Modelo),
		strings.ToUpper(d.Tipo),
		strings.ToUpper(d.Nome),
		strings.ToUpper(d.Particao),
		strings.ToUpper(d.Conta),
		strings.ToUpper(d.IdFisico1),
		strings.ToUpper(d.IdFisico2),
		strings.ToUpper(d.KeepAlive),
		strings.ToUpper(d.MsgAtendente),
		strings.ToUpper(d.Senha),
		strings.ToUpper(d.SenhaVerbal),
		strings.ToUpper(d.ContraSenhaVerbal),
	); err != nil {

		return err
	}

	// Cria o setup de relatorio===============================================
	var ee emailEventoV4.EmailEvento

	ee.ID_Dispositivo = d.ID_Dispositivo
	ee.EmailAlarme = "N"
	ee.EmailArme = "N"
	ee.EmailDesarme = "N"
	ee.EmailEmergencia = "N"
	ee.EmailFalhas = "N"
	ee.EmailGeral = "N"
	ee.EmailMedico = "N"
	ee.EmailPanico = "N"
	ee.EmailRestaure = "N"
	ee.EmailSetup = "N"
	ee.EmailTeste = "N"
	ee.EmailSemComunicar = "N"

	if err := ee.Insere(); err != nil {
		return err
	}

	return nil
}

func (d *Dispositivo) GetDadosById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE dispositivo.ID_Dispositivo = '%s'", d.ID_Dispositivo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, d); err != nil {
			return err
		}

		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) GetDadosByIdFisico1() error {
	if d.IdFisico1 == "" {
		return errors.New("um id fisico 1 deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE dispositivo.IdFisico1 = '%s'", d.IdFisico1)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, d); err != nil {
			return err
		}

		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) GetDadosByIdFisico2() error {
	if d.IdFisico2 == "" {
		return errors.New("um id fisico 2 deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE dispositivo.IdFisico2 = '%s'", d.IdFisico2)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, d); err != nil {
			return err
		}
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

// Manipula o campo msgAtendente ==============================================
func (d *Dispositivo) GetMsgAtendenteById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT dispositivo.MsgAtendente
		FROM dispositivo
		WHERE dispositivo.ID_Dispositivo = ?
	`, d.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&d.MsgAtendente); err != nil {
			return err
		}
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) SetMsgAtendenteById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE dispositivo 
		SET dispositivo.MsgAtendente = ?
		WHERE dispositivo.ID_Dispositivo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		d.MsgAtendente,
		d.ID_Dispositivo,
	); err != nil {
		return err
	}
	return nil
}

// Manipula o campo arme ======================================================
func (d *Dispositivo) GetArmadoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			dispositivo.Armado
			dispositivo.DataArmado
		FROM dispositivo
		WHERE dispositivo.ID_Dispositivo = ?
	`, d.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&d.Armado, &d.DataArmado); err != nil {
			return err
		}
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) SetArmadoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE dispositivo 
		SET 
			dispositivo.Armado = ?
			dispositivo.DataArmado = ?
		WHERE dispositivo.ID_Dispositivo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		d.Armado,
		time.Now().Format("2006-01-02 15:04:05"),
		d.ID_Dispositivo,
	); err != nil {
		return err
	}
	return nil
}

func (d *Dispositivo) InverteArmadoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if err := d.GetArmadoById(); err != nil {
		return err
	}

	if err := d.InverteArmadoById(); err != nil {
		return err
	}

	if err := d.SetArmadoById(); err != nil {
		return err
	}

	return nil
}

// Manipula o campo msgAtendente ==============================================
func (d *Dispositivo) GetAtivoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = d.ID_Dispositivo
	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	d.Ativo = lb.Ativo
	return nil
}

func (d *Dispositivo) SetAtivoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if d.Ativo == "" {
		return errors.New("um status de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = d.ID_Dispositivo
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = d.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (d *Dispositivo) InverteAtivoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = d.ID_Dispositivo
	lb.Descricao = "ALTERADO VIA WEB"

	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	d.Ativo = lb.Ativo
	return nil
}

func (d *Dispositivo) GravarUltimoEventoById() error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if d.DataUltimoEvento == "" {
		return errors.New("uma data de ultimo evento deve ser informada")
	}

	if d.CodgioUltimoEvento == "" {
		return errors.New("um codigo de ultimo evento deve se informado")
	}

	if d.DataUltimoEvento[2:3] == "/" {
		d.DataUltimoEvento = auxiliar.TimeBrToUs(d.DataUltimoEvento)
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE dispositivo SET 
			dispositivo.DataUltimoEvento = ?,
			dispositivo.CodgioUltimoEvento = ?
		WHERE dispositivo.ID_Dispositivo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		d.DataUltimoEvento,
		d.CodgioUltimoEvento,
		d.ID_Dispositivo,
	); err != nil {
		return err
	}
	return nil
}

func (d *Dispositivo) AlteraById() error {
	if d.ID_Dispositivo == "" {
		d.ID_Dispositivo = auxiliar.GeradorDeId()
	}

	if d.Nome == "" {
		return errors.New("um nome para o dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE dispositivo SET 		 
		dispositivo.ID_Fabricante = ?, 
		dispositivo.ID_Modelo = ?, 
		dispositivo.Tipo = ?, 
		dispositivo.Nome = ?, 
		dispositivo.Particao = ?, 
		dispositivo.Conta = ?, 
		dispositivo.IdFisico1 = ?, 
		dispositivo.IdFisico2 = ?, 
		dispositivo.KeepAlive = ?, 
		dispositivo.MsgAtendente = ?, 		
		dispositivo.Senha = ?,
		dispositivo.SenhaVerbal = ?, 
		dispositivo.ContraSenhaVerbal = ?
		
		WHERE dispositivo.ID_Dispositivo = ? 
		
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		d.ID_Fabricante,
		d.ID_Modelo,
		d.Tipo,
		d.Nome,
		d.Particao,
		d.Conta,
		d.IdFisico1,
		d.IdFisico2,
		d.KeepAlive,
		d.MsgAtendente,
		d.Senha,
		d.SenhaVerbal,
		d.ContraSenhaVerbal,
		d.ID_Dispositivo,
	); err != nil {
		return nil
	}
	return nil
}

func (d *Dispositivo) DeletaById() error {
	if d.ID_Dispositivo == "" {
		d.ID_Dispositivo = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM dispositivo		
		WHERE dispositivo.ID_Dispositivo = ? 
		
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(d.ID_Dispositivo); err != nil {
		return nil
	}
	return nil
}

func (d *Dispositivo) GetNomeClienteById(nome *string) error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT cliente.Nome
		FROM dispositivo
		LEFT JOIN dispositivo.ID_Cliente = cliente.ID_Cliente
		WHERE dispositivo.ID_Dispositivo = ?
	`, d.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var sqlNome sql.NullString
		if err := tab.Scan(&sqlNome); err != nil {
			return err
		}
		*nome = sqlNome.String
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) GetIdClienteById(idCliente *string) error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT dispositivo.ID_Cliente
		FROM dispositivo
		WHERE dispositivo.ID_Dispositivo = ?
	`, d.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var sqlId sql.NullString
		if err := tab.Scan(&sqlId); err != nil {
			return err
		}
		*idCliente = sqlId.String
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) GetIdFranqueadoById(idFranqueado *string) error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT cliente.ID_Franqueado
		FROM dispositivo
		LEFT JOIN dispositivo.ID_Cliente = cliente.ID_Cliente
		WHERE dispositivo.ID_Dispositivo = ?
	`, d.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var sqlId sql.NullString
		if err := tab.Scan(&sqlId); err != nil {
			return err
		}
		*idFranqueado = sqlId.String
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) GetIdRepresentanteById(idRepresentante *string) error {
	if d.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT franqueado.ID_Representante
		FROM dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente
		
		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado
		
		WHERE dispositivo.ID_Dispositivo = ?
	`, d.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var sqlId sql.NullString
		if err := tab.Scan(&sqlId); err != nil {
			return err
		}
		*idRepresentante = sqlId.String
		return nil
	}
	return errors.New("dispositivo não encontrado na base de dados")
}

func (d *Dispositivo) ListarSemComunicacaoByIdFranqueado(lista *[]Dispositivo) error {

	// Valida campos obrigatorio
	if d.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria um filtro para consulta
	filtro := fmt.Sprintf(`
		WHERE dispositivo.ID_Cliente IN (
			SELECT cliente.ID_Cliente 
			FROM cliente
			WHERE cliente.ID_Franqueado = '%s'  
		) 
		AND dispositivo.ID_Dispositivo NOT IN (
			SELECT listaBloqueio.ID_Alvo 
			FROM listaBloqueio
		)
		AND dispositivo.DataUltimoEvento < '%s'
		ORDER BY dispositivo.DataUltimoEvento
	`, d.ID_Franqueado, time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05"))

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Dispositivo
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}

	return nil
}

func (d *Dispositivo) ListarByIdCliente(lista *[]Dispositivo) error {
	if d.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE dispositivo.ID_Cliente = '%s'", d.ID_Cliente)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Dispositivo
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

func (d *Dispositivo) ListarByIdFranqueado(lista *[]Dispositivo) error {
	if d.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE franqueado.ID_Franqueado = '%s'", d.ID_Franqueado)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Dispositivo
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

// Manipula o campo Conta =====================================================
func (d *Dispositivo) GerarContaByIdFranqueado() error {

	if d.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser fornecido")
	}

	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
			SELECT dispositivo.Conta

			FROM cliente
			
			LEFT JOIN dispositivo
			ON cliente.ID_Cliente = dispositivo.ID_Cliente

			WHERE cliente.ID_Franqueado = ? 

			ORDER BY dispositivo.Conta
		`, d.ID_Franqueado)
	if err != nil {

		return err
	}
	defer tab.Close()

	var contas []int

	for tab.Next() {
		// Pega um numero de conta no banco
		var sqlConta sql.NullString
		if err = tab.Scan(&sqlConta); err != nil {
			return err
		}
		if sqlConta.String != "" {
			// Coverte numero da conta em inteiro para comparação
			conta, erro := strconv.Atoi(sqlConta.String)
			if erro != nil {
				return err
			}
			contas = append(contas, conta)
		}
	}

	// fmt.Println(contas)
	conta := 1
	for conta <= 9999 {
		livre := true
		for _, c := range contas {
			if conta == c {
				livre = false
				break
			}
		}
		if livre {
			d.Conta = fmt.Sprintf("%04d", conta)
			break
		}
		conta++
	}
	if conta > 9999 {
		d.Conta = "Limite de conta exedido"
		return nil
	}

	return nil
}

func (d *Dispositivo) VericaContaByIdFranquado() error {
	if d.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	if d.Conta == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT cliente.Nome

		FROM cliente
		
		LEFT JOIN dispositivo
		ON cliente.ID_Cliente = dispositivo.ID_Cliente

		WHERE cliente.ID_Franqueado = ?
		AND dispositivo.Conta = ?
	`, d.ID_Franqueado, d.Conta)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&d.NomeCliente); err != nil {
			return err
		}
	} else {
		d.NomeCliente = "LIVRE"
	}

	return nil
}

// Funções internas ===========================================================
func sDispositivoToDispositivo(disp SDispositivo) (d Dispositivo) {
	d.ID_Franqueado = disp.ID_Franqueado.String
	d.NomeFranqueado = disp.NomeFranqueado.String

	d.ID_Cliente = disp.ID_Cliente.String
	d.NomeCliente = disp.NomeCliente.String

	d.ID_Modelo = disp.ID_Modelo.String
	d.Tipo = disp.Tipo.String
	d.Nome = disp.Nome.String
	d.Particao = disp.Particao.String
	d.Conta = disp.Conta.String
	d.IdFisico1 = disp.IdFisico1.String
	d.IdFisico2 = disp.IdFisico2.String
	d.KeepAlive = disp.KeepAlive.String
	d.Armado = disp.Armado.String
	d.DataArmado = disp.DataArmado.Time.Format("02/01/2006 15:04:05")
	d.MsgAtendente = disp.MsgAtendente.String

	if disp.DataUltimoEvento.Valid {
		d.DataUltimoEvento = disp.DataUltimoEvento.Time.Format("02/01/2006 15:04:05")
	} else {
		d.DataUltimoEvento = ""
	}

	d.CodgioUltimoEvento = disp.CodgioUltimoEvento.String
	d.DescricaoUltimoEvento = disp.DescricaoUltimoEvento.String

	if disp.Manutencao.Valid {
		d.Manutencao = disp.Manutencao.Time.Format("02/01/2006 15:04:05")
	} else {
		d.Manutencao = ""
	}

	d.Senha = disp.Senha.String
	d.SenhaVerbal = disp.SenhaVerbal.String
	d.ContraSenhaVerbal = disp.ContraSenhaVerbal.String

	if disp.DataCadastro.Valid {
		d.DataCadastro = disp.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		d.DataCadastro = ""
	}

	d.ID_Dispositivo = disp.ID_Dispositivo.String
	d.ID_Fabricante = disp.ID_Fabricante.String
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			cliente.ID_Franqueado,
			franqueado.RazaoSocial,

			dispositivo.ID_Cliente, 
			cliente.Nome,
			
			dispositivo.ID_Dispositivo, 			
			dispositivo.Tipo, 
			dispositivo.Nome, 
			dispositivo.Particao, 
			dispositivo.Conta, 
			dispositivo.IdFisico1, 
			dispositivo.IdFisico2, 
			dispositivo.KeepAlive, 
			dispositivo.Armado, 
			dispositivo.DataArmado, 
			dispositivo.MsgAtendente, 
			dispositivo.DataUltimoEvento, 
			dispositivo.CodgioUltimoEvento,			 
			dispositivo.Manutencao, 
			dispositivo.Senha, 
			dispositivo.SenhaVerbal, 
			dispositivo.ContraSenhaVerbal, 			 
			dispositivo.DataCadastro,
			
			dispositivo.ID_Fabricante, 
			dispositivo.ID_Modelo,

			bloqDispositivo.ID_Alvo,
			
			bloqCliente.ID_Alvo,

			bloqFranqueado.ID_Alvo,

			bloqRepresentante.ID_Alvo
		
		FROM dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente		

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN listaBloqueio AS bloqDispositivo
		ON bloqDispositivo.ID_Alvo = dispositivo.ID_Dispositivo

		LEFT JOIN listaBloqueio AS bloqCliente
		ON bloqCliente.ID_Alvo = cliente.ID_Cliente

		LEFT JOIN listaBloqueio AS bloqFranqueado
		ON bloqFranqueado.ID_Alvo =  franqueado.ID_Franqueado

		LEFT JOIN listaBloqueio AS bloqRepresentante
		ON bloqRepresentante.ID_Alvo =  franqueado.ID_Representante
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, d *Dispositivo) error {

	var (
		tmp     SDispositivo
		bloqDis sql.NullString
		bloqCli sql.NullString
		bloqFra sql.NullString
		bloqRep sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Franqueado,
		&tmp.NomeFranqueado,

		&tmp.ID_Cliente,
		&tmp.NomeCliente,

		&tmp.ID_Dispositivo,
		&tmp.Tipo,
		&tmp.Nome,
		&tmp.Particao,
		&tmp.Conta,
		&tmp.IdFisico1,
		&tmp.IdFisico2,
		&tmp.KeepAlive,
		&tmp.Armado,
		&tmp.DataArmado,
		&tmp.MsgAtendente,
		&tmp.DataUltimoEvento,
		&tmp.CodgioUltimoEvento,
		&tmp.Manutencao,
		&tmp.Senha,
		&tmp.SenhaVerbal,
		&tmp.ContraSenhaVerbal,

		&tmp.DataCadastro,

		&tmp.ID_Fabricante,
		&tmp.ID_Modelo,

		&bloqDis,
		&bloqCli,
		&bloqFra,
		&bloqRep,
	); err != nil {
		return err
	}

	var cti contactidV4.ContactId
	cti.Codigo = tmp.CodgioUltimoEvento.String
	cti.ID_Viculo = d.ID_Franqueado
	if err := cti.GetDadosByCodigo(); err != nil {
		tmp.DescricaoUltimoEvento.String = "Não Cadastrado"
	} else {
		tmp.DescricaoUltimoEvento.String = cti.Descricao
	}

	disp := sDispositivoToDispositivo(tmp)

	if bloqDis.Valid {
		disp.Ativo = "N"
	} else if bloqCli.Valid {
		disp.Ativo = "N"
	} else if bloqFra.Valid {
		disp.Ativo = "N"
	} else if bloqRep.Valid {
		disp.Ativo = "N"
	} else {
		disp.Ativo = "S"
	}

	*d = disp
	return nil
}
