package tarifa

import (
	"context"

	"apifunction/auth"
	"apifunction/db"
	"apifunction/ops"
	"apifunction/pgcatalogo"
	"apifunction/service/contrato"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Tarifa struct {
	IDCentral       string  `json:"id_central"`
	Canal           string  `json:"canal"`
	ValorTentativa  float64 `json:"valor_tentativa"`
	ValorMinuto     float64 `json:"valor_minuto"`
	ValorUnidade    float64 `json:"valor_unidade"`
	PisoTentativa   float64 `json:"piso_tentativa,omitempty"`
	PisoMinuto      float64 `json:"piso_minuto,omitempty"`
	PisoUnidade     float64 `json:"piso_unidade,omitempty"`
	ModoPreco       string  `json:"modo_preco,omitempty"`
}

type TarifaRep struct {
	IDRepresentante string  `json:"id_representante"`
	IDCentral       string  `json:"id_central"`
	Canal           string  `json:"canal"`
	ValorTentativa  float64 `json:"valor_tentativa"`
	ValorMinuto     float64 `json:"valor_minuto"`
	ValorUnidade    float64 `json:"valor_unidade"`
	PisoTentativa   float64 `json:"piso_tentativa,omitempty"`
	PisoMinuto      float64 `json:"piso_minuto,omitempty"`
	PisoUnidade     float64 `json:"piso_unidade,omitempty"`
}

func ListarCentral(idCentral string) ([]Tarifa, string, error) {
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		idCentral = "CENTRAL"
	}
	modo, _ := pgcatalogo.ModoPreco(context.Background(), idCentral)
	rows, err := db.Conn.Query(`
SELECT ID_Central, Canal, ValorTentativa, ValorMinuto, ValorUnidade,
       COALESCE(PisoTentativa, 0), COALESCE(PisoMinuto, 0), COALESCE(PisoUnidade, 0)
FROM fp_tarifa_operacional WHERE ID_Central = ? AND Ativo = 'S' ORDER BY Canal`, idCentral)
	if err != nil {
		return nil, modo, err
	}
	defer rows.Close()
	var out []Tarifa
	for rows.Next() {
		var t Tarifa
		if err := rows.Scan(&t.IDCentral, &t.Canal, &t.ValorTentativa, &t.ValorMinuto, &t.ValorUnidade,
			&t.PisoTentativa, &t.PisoMinuto, &t.PisoUnidade); err != nil {
			return nil, modo, err
		}
		t.ModoPreco = modo
		out = append(out, t)
	}
	return out, modo, rows.Err()
}

func SalvarCentral(sess auth.SessaoAdm, t Tarifa) error {
	idCentral := strings.TrimSpace(t.IDCentral)
	if idCentral == "" {
		idCentral = sess.IDCentralCatalogo
	}
	if idCentral == "" {
		return errors.New("id_central obrigatorio")
	}
	canal := strings.TrimSpace(t.Canal)
	if canal == "" {
		return errors.New("canal obrigatorio")
	}

	bg := isBreakglass(sess)
	modo, _ := pgcatalogo.ModoPreco(context.Background(), idCentral)

	pisoT, pisoM, pisoU := t.PisoTentativa, t.PisoMinuto, t.PisoUnidade
	var atualT, atualM, atualU float64
	err := db.Conn.QueryRow(`
SELECT COALESCE(PisoTentativa, 0),
       COALESCE(PisoMinuto, 0),
       COALESCE(PisoUnidade, 0)
FROM fp_tarifa_operacional WHERE ID_Central = ? AND Canal = ?`, idCentral, canal).Scan(&atualT, &atualM, &atualU)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		if pisoT <= 0 {
			pisoT = atualT
		}
		if pisoM <= 0 {
			pisoM = atualM
		}
		if pisoU <= 0 {
			pisoU = atualU
		}
		if !bg && modo == "piso" {
			if pisoT > 0 && t.ValorTentativa < pisoT {
				return fmt.Errorf("Central em modo piso: valor minimo e R$ %.2f (tentativa)", pisoT)
			}
			if canal == "ligacao" && pisoM > 0 && t.ValorMinuto < pisoM {
				return fmt.Errorf("Central em modo piso: valor minimo e R$ %.2f (minuto)", pisoM)
			}
			if canal != "ligacao" && pisoU > 0 && t.ValorUnidade < pisoU {
				return fmt.Errorf("Central em modo piso: valor minimo e R$ %.2f (unidade)", pisoU)
			}
		}
	}
	if bg {
		pisoT = t.ValorTentativa
		pisoM = t.ValorMinuto
		pisoU = t.ValorUnidade
	} else if err == sql.ErrNoRows && (modo == "piso" || bg) {
		pisoT = t.ValorTentativa
		pisoM = t.ValorMinuto
		pisoU = t.ValorUnidade
	}

	_, err = db.Conn.Exec(`
INSERT INTO fp_tarifa_operacional (ID_Central, Canal, ValorTentativa, ValorMinuto, ValorUnidade, PisoTentativa, PisoMinuto, PisoUnidade, Ativo)
VALUES (?,?,?,?,?,?,?,?, 'S')
ON DUPLICATE KEY UPDATE
  ValorTentativa = VALUES(ValorTentativa),
  ValorMinuto = VALUES(ValorMinuto),
  ValorUnidade = VALUES(ValorUnidade),
  PisoTentativa = VALUES(PisoTentativa),
  PisoMinuto = VALUES(PisoMinuto),
  PisoUnidade = VALUES(PisoUnidade),
  Ativo = 'S'`, idCentral, canal, t.ValorTentativa, t.ValorMinuto, t.ValorUnidade, pisoT, pisoM, pisoU)
	if err != nil {
		return err
	}
	return ops.SyncTarifa(idCentral, canal, t.ValorTentativa, t.ValorMinuto, t.ValorUnidade)
}

func ListarRep(idRep string) ([]TarifaRep, error) {
	idRep = strings.TrimSpace(idRep)
	if idRep == "" {
		return nil, errors.New("id_representante obrigatorio")
	}
	rows, err := db.Conn.Query(`
SELECT ID_Representante, ID_Central, Canal, ValorTentativa, ValorMinuto, ValorUnidade
FROM fp_tarifa_operacional_rep WHERE ID_Representante = ? AND Ativo = 'S' ORDER BY Canal`, idRep)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TarifaRep
	var idCentral string
	for rows.Next() {
		var t TarifaRep
		if err := rows.Scan(&t.IDRepresentante, &t.IDCentral, &t.Canal, &t.ValorTentativa, &t.ValorMinuto, &t.ValorUnidade); err != nil {
			return nil, err
		}
		if idCentral == "" {
			idCentral = t.IDCentral
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if idCentral != "" {
		pisos, _, errP := ListarCentral(idCentral)
		if errP == nil {
			pisoPorCanal := map[string]Tarifa{}
			for _, p := range pisos {
				pisoPorCanal[p.Canal] = p
			}
			for i := range out {
				if p, ok := pisoPorCanal[out[i].Canal]; ok {
					out[i].PisoTentativa = p.PisoTentativa
					out[i].PisoMinuto = p.PisoMinuto
					out[i].PisoUnidade = p.PisoUnidade
				}
			}
		}
	}
	return out, nil
}

func SalvarRep(sess auth.SessaoAdm, t TarifaRep) error {
	idRep := strings.TrimSpace(t.IDRepresentante)
	if idRep == "" {
		idRep = sess.IDRepresentante
	}
	if idRep == "" {
		return errors.New("id_representante obrigatorio")
	}
	idCentral := strings.TrimSpace(t.IDCentral)
	if idCentral == "" {
		idCentral = sess.IDCentralCatalogo
	}
	canal := strings.TrimSpace(t.Canal)
	if canal == "" {
		return errors.New("canal obrigatorio")
	}
	usa, err := contrato.RepresentanteUsaAdm(idRep)
	if err != nil {
		return err
	}
	if usa != "S" && sess.UserTipo == "REP" {
		return errors.New("representante sem UsaAdmConfmonit")
	}

	var pisoT, pisoM, pisoU, valT, valM, valU float64
	err = db.Conn.QueryRow(`
SELECT COALESCE(PisoTentativa, 0), COALESCE(PisoMinuto, 0), COALESCE(PisoUnidade, 0),
       COALESCE(ValorTentativa, 0), COALESCE(ValorMinuto, 0), COALESCE(ValorUnidade, 0)
FROM fp_tarifa_operacional WHERE ID_Central = ? AND Canal = ? AND Ativo = 'S'`,
		idCentral, canal).Scan(&pisoT, &pisoM, &pisoU, &valT, &valM, &valU)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	pisoRepT, pisoRepM, pisoRepU := pisoT, pisoM, pisoU
	if pisoRepT <= 0 {
		pisoRepT = valT
	}
	if pisoRepM <= 0 {
		pisoRepM = valM
	}
	if pisoRepU <= 0 {
		pisoRepU = valU
	}
	if t.ValorTentativa < pisoRepT {
		return fmt.Errorf("valor tentativa nao pode ser menor que o piso da Central (R$ %.4f)", pisoRepT)
	}
	if canal == "ligacao" && t.ValorMinuto < pisoRepM {
		return fmt.Errorf("valor minuto nao pode ser menor que o piso da Central (R$ %.4f)", pisoRepM)
	}
	if canal != "ligacao" && t.ValorUnidade < pisoRepU {
		return fmt.Errorf("valor unidade nao pode ser menor que o piso da Central (R$ %.4f)", pisoRepU)
	}

	_, err = db.Conn.Exec(`
INSERT INTO fp_tarifa_operacional_rep (ID_Representante, ID_Central, Canal, ValorTentativa, ValorMinuto, ValorUnidade, Ativo)
VALUES (?,?,?,?,?,?, 'S')
ON DUPLICATE KEY UPDATE
  ValorTentativa = VALUES(ValorTentativa),
  ValorMinuto = VALUES(ValorMinuto),
  ValorUnidade = VALUES(ValorUnidade),
  Ativo = 'S'`, idRep, idCentral, canal, t.ValorTentativa, t.ValorMinuto, t.ValorUnidade)
	return err
}

func RepresentanteUsaAdmExport(idRep string) (string, error) {
	return contrato.RepresentanteUsaAdm(idRep)
}

func SeedTarifasCentral(idCentral string) error {
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		idCentral = "CENTRAL"
	}
	for _, canal := range []string{"ligacao", "sms", "whatsapp", "email"} {
		var n int
		_ = db.Conn.QueryRow(`SELECT COUNT(*) FROM fp_tarifa_operacional WHERE ID_Central = ? AND Canal = ?`, idCentral, canal).Scan(&n)
		if n > 0 {
			continue
		}
		_, err := db.Conn.Exec(`
INSERT INTO fp_tarifa_operacional (ID_Central, Canal, ValorTentativa, ValorMinuto, ValorUnidade, PisoTentativa, PisoMinuto, PisoUnidade, Ativo)
VALUES (?, ?, 0, 0, 0, 0, 0, 0, 'S')`, idCentral, canal)
		if err != nil {
			return err
		}
		_ = ops.SyncTarifa(idCentral, canal, 0, 0, 0)
	}
	return nil
}

var canaisOperacionais = []string{"ligacao", "sms", "whatsapp", "email"}

func ListarEfetivas(idFranqueado string) ([]TarifaRep, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, errors.New("id_franqueado obrigatorio")
	}
	out := make([]TarifaRep, 0, len(canaisOperacionais))
	for _, canal := range canaisOperacionais {
		t, err := GetTarifaEfetiva(idFranqueado, canal)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		t.Canal = canal
		out = append(out, t)
	}
	return out, nil
}

func GetTarifaEfetiva(idFranqueado, canal string) (TarifaRep, error) {
	out := TarifaRep{Canal: canal}
	idRep, idCentral, err := contrato.FranqueadoRepresentante(idFranqueado)
	if err != nil {
		return out, err
	}
	out.IDRepresentante = idRep
	out.IDCentral = idCentral
	err = db.Conn.QueryRow(`
SELECT ValorTentativa, ValorMinuto, ValorUnidade FROM fp_tarifa_operacional_rep
WHERE ID_Representante = ? AND Canal = ? AND Ativo = 'S'`, idRep, canal).Scan(
		&out.ValorTentativa, &out.ValorMinuto, &out.ValorUnidade)
	if err == nil {
		return out, nil
	}
	if err != sql.ErrNoRows {
		return out, err
	}
	if idCentral == "" {
		idCentral = "CENTRAL"
	}
	err = db.Conn.QueryRow(`
SELECT ValorTentativa, ValorMinuto, ValorUnidade FROM fp_tarifa_operacional
WHERE ID_Central = ? AND Canal = ? AND Ativo = 'S'`, idCentral, canal).Scan(
		&out.ValorTentativa, &out.ValorMinuto, &out.ValorUnidade)
	return out, err
}

func isBreakglass(sess auth.SessaoAdm) bool {
	return strings.EqualFold(strings.TrimSpace(sess.IDUsuario), "BREAKGLASS")
}

func PisoCentral(idCentral, canal string) (Tarifa, error) {
	lista, _, err := ListarCentral(idCentral)
	if err != nil {
		return Tarifa{}, err
	}
	for _, t := range lista {
		if t.Canal == canal {
			return t, nil
		}
	}
	return Tarifa{Canal: canal}, sql.ErrNoRows
}
