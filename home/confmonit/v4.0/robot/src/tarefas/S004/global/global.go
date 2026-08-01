package g004

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	"strconv"
	"strings"
	"time"
)

type Dados struct {
	NucId          sql.NullString
	NucNome        sql.NullString
	FraId          sql.NullString
	FraNome        sql.NullString
	CliId          sql.NullString
	CliNome        sql.NullString
	PacId          sql.NullString
	PacNome        sql.NullString
	PacValor       sql.NullString
	PacGrade       sql.NullString
	PacGradeValor  sql.NullString
	PacAteQtd      sql.NullString
	PacAteExedente sql.NullString
	PacEmaQtd      sql.NullString
	PacEmaExedente sql.NullString
	PacSmsQtd      sql.NullString
	PacSmsExedente sql.NullString
	PacLigQtd      sql.NullString
	PacLigExedente sql.NullString

	FaturaId string
	Itens    []Item

	// Variaveis auxiliares para calculo de exedentes
	QtdAtendimento int
	QtdLigacao     int
	QtdEmail       int
	QtdSms         int
}

type Item struct {
	Quantidade string
	Descricao  string
	Credito    float64
	Debito     float64
}

// converte string para float
func StrigToFloat(in string, out *float64) error {
	// Converte valor exedente pora float
	temp := strings.ReplaceAll(in, ".", "")
	temp = strings.ReplaceAll(temp, ",", ".")
	if temp == "0.00" {
		*out = 0.00
	} else {

		valor, erro := strconv.ParseFloat(temp, 64)
		if erro != nil {
			return erro
		}
		*out = valor
	}

	return nil
}

// converte float para string
func FloatToString(in float64) string {
	return strings.ReplaceAll(fmt.Sprintf("%.2f", in), ".", ",")
}

func BuscarClientes(db *sql.DB, listaFra []string, listaCli *[]string) error {
	data := time.Now().Add(-360 * time.Hour).Format("2006-01-02 15:04:05")
	for _, fra := range listaFra {
		// Busca todos os clientes do franqueado
		tab, err := db.Query(`
		SELECT cliente.ID_Cliente, cliente.Nome
		FROM cliente 
		WHERE cliente.ID_Franqueado = ?
		AND (
		 	cliente.DataCancelamento IS NULL
        	OR  cliente.DataCancelamento > ? 
		)
	 `, fra, data)
		if err != nil {
			return err
		}
		defer tab.Close()

		for tab.Next() {
			var id sql.NullString
			var nome sql.NullString
			if err := tab.Scan(&id, &nome); err != nil {
				return err
			}
			fmt.Println("    cliente ->", nome.String)
			*listaCli = append(*listaCli, id.String)
		}
	}
	return nil
}

func GravaIdFatura(db *sql.DB, faturaId, tarifaId string) error {

	stm, err := db.Prepare(`
		UPDATE tarifacao 
		SET tarifacao.ID_FaturaNuc	= ? 
		WHERE tarifacao.ID_Tarifacao = ?	
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(faturaId, tarifaId); err != nil {
		return err
	}

	return nil
}

// Cria uma nova fatura
func CriaFatura(db *sql.DB, faturaId, origemId, destinoId string) error {
	idCentralUUID := resolveIDCentralUUID(db, origemId, destinoId)

	stm, err := db.Prepare(`
		INSERT INTO faturas(
			ID_Fatura, 
			ID_Origem,
			IDCentralUUID,
			ID_Destino, 
			Status, 
			Valor, 
			DataCadastro, 
			DataUltimoStatus, 
			DataVencimento
		) VALUES (?,?,?,?,?,?,?,?,?)
	`)

	if err != nil {
		return err
	}
	defer stm.Close()

	var idCentral interface{}
	if idCentralUUID != "" {
		idCentral = idCentralUUID
	}

	if _, err := stm.Exec(
		faturaId,
		origemId,
		idCentral,
		destinoId,
		"AGUARDANDO PAGAMENTO",
		"0,00",
		time.Now().Format("2006-01-02 15:04:04"),
		time.Now().Format("2006-01-02 15:04:04"),
		time.Now().Add(360*time.Hour).Format("2006-01-02 15:04:04"),
	); err != nil {
		return err
	}
	return nil
}

func resolveIDCentralUUID(db *sql.DB, origemId, destinoId string) string {
	var uuid sql.NullString
	// Preferencia: representante destino (fatura Central->Rep) ou origem (Rep->Fra)
	_ = db.QueryRow(`
		SELECT IDCentralUUID FROM representante
		WHERE ID_Representante IN (?, ?)
		AND IDCentralUUID IS NOT NULL AND IDCentralUUID <> ''
		LIMIT 1
	`, destinoId, origemId).Scan(&uuid)
	if uuid.Valid && uuid.String != "" {
		return uuid.String
	}
	_ = db.QueryRow(`
		SELECT IDCentralUUID FROM central
		WHERE ID_Central = 'CENTRAL'
		LIMIT 1
	`).Scan(&uuid)
	if uuid.Valid {
		return uuid.String
	}
	return ""
}

func GravaTotal(db *sql.DB, idFatura string, credito, debito float64) error {

	total := FloatToString(debito - credito)

	stm, err := db.Prepare(`
		UPDATE faturas 
		SET Valor = ? 
		WHERE faturas.ID_Fatura = ?	
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(total, idFatura); err != nil {
		return err
	}

	return nil
}

func CarregaGrade(db *sql.DB, listaCli []string, valorGrade string, itens *[]Item) error {
	txtSql := fmt.Sprintf(`
		SELECT COUNT(grade.ID_Grade) 
		FROM grade 
		LEFT JOIN dispositivo 
		ON grade.ID_Dispositivo = dispositivo.ID_Dispositivo 
		WHERE dispositivo.ID_Cliente IN(%s) 
		AND grade.ID_Grade NOT IN (
			SELECT listaBloqueio.ID_Alvo FROM listaBloqueio
		)	
	`, strings.Join(listaCli, ","))

	tab, err := db.Query(txtSql)

	if err != nil {
		return err
	}
	defer tab.Close()
	if tab.Next() {
		var qtd int
		if err := tab.Scan(&qtd); err != nil {
			return err
		}

		// grade
		if qtd > 0 {

			var valor float64
			if err := StrigToFloat(valorGrade, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i Item
			i.Quantidade = strconv.Itoa(qtd)
			i.Descricao = "GRADE ATIVA"
			i.Credito = 0.00
			i.Debito = float64(qtd) * valor

			// Adiciona o item na lista de itens da fatura
			*itens = append(*itens, i)

		}
	}
	return nil
}

// Adiciona os itens a fatura criada
func AdicionaItem(db *sql.DB, idFatura string, itens []Item) error {
	var credito, debito float64

	stm, err := db.Prepare(`
		INSERT INTO faturasItem(
			ID_FaturaItem, 
			ID_Fatura, 
			Quantidade, 
			Descricao, 
			Credito, 
			Debito, 
			Tipo, 
			DataCadastro
		) VALUES (?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	for _, i := range itens {
		if _, err := stm.Exec(
			auxiliar.GeradorDeId(),
			idFatura,
			i.Quantidade,
			i.Descricao,
			FloatToString(i.Credito),
			FloatToString(i.Debito),
			"",
			time.Now().Format("2006-01-02 15:04:04"),
		); err != nil {
			return err
		}

		credito = credito + i.Credito
		debito = debito + i.Debito
	}

	// Grava o totais dos itens na fatura
	GravaTotal(db, idFatura, credito, debito)
	return nil
}

func InserePacote(valorPacote, nomePacote string, itens *[]Item) error {

	if valorPacote == "" {
		valorPacote = "0,00"
	}
	var i Item
	i.Quantidade = "1"
	i.Descricao = "PACOTE" + nomePacote
	i.Credito = 0.00
	if err := StrigToFloat(valorPacote, &i.Debito); err != nil {
		return err
	}

	*itens = append(*itens, i)
	return nil
}

func InsereAdicionais(dadoOperacao, credito, debito string, itens *[]Item) error {

	var i Item
	i.Quantidade = "1"
	i.Descricao = dadoOperacao

	if err := StrigToFloat(credito, &i.Credito); err != nil {
		return err
	}
	if err := StrigToFloat(debito, &i.Debito); err != nil {
		return err
	}
	*itens = append(*itens, i)
	return nil
}
