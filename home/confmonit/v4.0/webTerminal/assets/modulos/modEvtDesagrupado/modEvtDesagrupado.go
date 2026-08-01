package modEvtDesagrupado

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/modEvtDesagrupado/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso   string `json:"idProcesso"`
		IdEvento     string `json:"idEvento"`
		IdDispositivo string `json:"idDispositivo"`
		Conta        string `json:"conta"`
		Codigo       string `json:"codigo"`
		Particao     string `json:"particao"`
		ZonaUser     string `json:"zonaUser"`
		NomeZoneUser string `json:"nomeZonaUser"`
		Camera       string `json:"camera"`
		Entrada      string `json:"entrada"`
		Descricao    string `json:"descricao"`
		CodFranq     string `json:"codFranq"`
		ProvedorVideo string `json:"provedorVideo"`
		UsaConfVision string `json:"usaConfVision"`
		CameraOn     string `json:"cameraOn"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	fmt.Println(obj)

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT 
			evento.Codigo,
			evento.Particao,
			evento.ZonaUser,
			evento.DataEntrada,
			evento.Img,
			evento.ID_Evento,
			
			processo.ID_Processo,
			
			dispositivo.ID_Dispositivo,	
			dispositivo.Conta,
			dispositivo.ProvedorVideo,

			franqueado.ID_Franqueado,
			franqueado.CodBenuvem,
			franqueado.UsaConfVision,

			setorAlarme.Camera

		FROM evento		
		
		LEFT JOIN processo 
		ON evento.ID_Processo = processo.ID_Processo
		
		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN setorAlarme
		ON setorAlarme.ID_Dispositivo = dispositivo.ID_Dispositivo
		AND setorAlarme.Numero = evento.ZonaUser
		AND setorAlarme.Particao = evento.Particao

		WHERE evento.ID_Processo = ?
		AND evento.Codigo = ?
		AND evento.ZonaUser = ?

	`, obj.IdProcesso, obj.Codigo, obj.ZonaUser)

	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto

	for tab.Next() {
		var (
			item          objeto
			entrada       sql.NullTime
			idDispositivo sql.NullString
			idProcesso    sql.NullString
			idEvento      sql.NullString
			idFraqueado   sql.NullString
			codVoip       sql.NullString
			provedorVideo sql.NullString
			usaConfVision sql.NullString
			cameraOn      sql.NullString
		)
		if erro := tab.Scan(
			&item.Codigo,
			&item.Particao,
			&item.ZonaUser,
			&entrada,
			&item.Camera,
			&idEvento,

			&idProcesso,

			&idDispositivo,
			&item.Conta,
			&provedorVideo,

			&idFraqueado,
			&codVoip,
			&usaConfVision,

			&cameraOn,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.IdProcesso = idProcesso.String
		item.IdEvento = idEvento.String
		item.IdDispositivo = idDispositivo.String
		item.Entrada = entrada.Time.Format("02/01/2006 15:04:05")
		item.CodFranq = codVoip.String
		item.ProvedorVideo = strings.ToLower(strings.TrimSpace(provedorVideo.String))
		item.UsaConfVision = usaConfVision.String
		item.CameraOn = "N"
		if cameraOn.Valid {
			item.CameraOn = cameraOn.String
		}

		var grupo string
		if err := aux.GetCtiGrupoAndDescricao(
			idFraqueado.String, item.Codigo, &grupo, &item.Descricao,
		); err != nil {
			aux.RespostaErro(w, http.StatusBadRequest, err)
			return
		}

		if grupo == "ARME" || grupo == "DESARME" || grupo == "PANICO" {
			if erro = buscaUsuario(db, idDispositivo.String, item.ZonaUser, &item.NomeZoneUser); erro != nil {
				aux.RespostaErro(w, http.StatusBadRequest, erro)
				return
			}
		} else {
			if erro = buscaSetor(db, idDispositivo.String, item.ZonaUser, &item.NomeZoneUser); erro != nil {
				aux.RespostaErro(w, http.StatusBadRequest, erro)
				return
			}
		}

		lista = append(lista, item)
	}
	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}

func buscaSetor(db *sql.DB, idDisp, setor string, nomeSetor *string) error {
	fmt.Println("Setor:", idDisp, setor)
	tab, erro := db.Query(`
		SELECT setorAlarme.nome
		FROM setorAlarme
		WHERE setorAlarme.ID_Dispositivo = ?
		AND setorAlarme.numero = ?
	`, idDisp, setor)
	if erro != nil {
		return erro
	}

	if tab.Next() {
		if erro := tab.Scan(nomeSetor); erro != nil {
			return erro
		}
		return nil
	}
	*nomeSetor = "Não Cadastrado"
	return nil
}

func buscaUsuario(db *sql.DB, idDisp, usuario string, nomeUsuario *string) error {
	fmt.Println("Usuario", idDisp, usuario)
	tab, erro := db.Query(`
		SELECT usuariosAlarme.nome
		FROM usuariosAlarme
		WHERE usuariosAlarme.ID_Dispositivo = ?
		AND usuariosAlarme.codigo = ?
	`, idDisp, usuario)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	if tab.Next() {
		if erro := tab.Scan(nomeUsuario); erro != nil {
			return erro
		}
		return nil
	}
	*nomeUsuario = "Desconhecido"
	return nil
}
