package terminalV4

import (
	connV4 "api/src/V4/conexao"
	"database/sql"
)

type evtDetalhe struct {
	ID_Evento   string `json:"idEvento"`
	ID_Processo string `json:"idProcesso"`
	Codigo      string `json:"codigo"`
	ZonaUser    string `json:"zonaUser"`
	DataEntrada string `json:"dataEntrada"`
	Descricao   string `json:"descricao"`
	Grupo       string `json:"grupo"`
	UserSet     string `json:"userSet"`
}

func (ed *evtDetalhe) listarEvtDetalheByProc(lista *[]evtDetalhe) error {

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
			evento.ID_Evento,
			evento.ID_Processo,
			evento.Codigo,
			evento.ZonaUser,
			evento.DataEntrada,
            
            ctiPadrao.Descricao,
            ctiPadrao.Grupo,
            
            ctiPersonalizado.Descricao,			
            ctiPersonalizado.Grupo,	
			
			usuariosAlarme.Nome,

			setorAlarme.Nome
			
		FROM evento
		
		LEFT JOIN processo
		ON evento.ID_Processo = processo.ID_Processo

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente
        
        LEFT JOIN contactId AS ctiPadrao
        ON evento.Codigo = ctiPadrao.Codigo
        AND ctiPadrao.ID_Vinculo = 'CENTRAL' 
        
        LEFT JOIN contactId AS ctiPersonalizado
        ON evento.Codigo = ctiPersonalizado.Codigo
        AND ctiPersonalizado.ID_Vinculo = cliente.ID_Franqueado

		LEFT JOIN usuariosAlarme
        ON evento.ZonaUser = usuariosAlarme.Codigo
        AND dispositivo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo

		LEFT JOIN setorAlarme
        ON evento.ZonaUser = setorAlarme.Numero
        AND dispositivo.ID_Dispositivo = setorAlarme.ID_Dispositivo
        AND evento.Particao = setorAlarme.Particao


		WHERE evento.ID_Processo = ?

	`, ed.ID_Processo)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var (
			i                                       evtDetalhe
			dtEnt                                   sql.NullTime
			usuarioAlarme, setorAlarme              sql.NullString
			ctiPadrao, ctiPadraoGrupo               sql.NullString
			ctiPersonalizado, ctiPersonalizadoGrupo sql.NullString
		)
		if err := tab.Scan(
			&i.ID_Evento,
			&i.ID_Processo,
			&i.Codigo,
			&i.ZonaUser,
			&dtEnt,
			&ctiPadrao,
			&ctiPadraoGrupo,
			&ctiPersonalizado,
			&ctiPersonalizadoGrupo,
			&usuarioAlarme,
			&setorAlarme,
		); err != nil {
			return err
		}
		i.DataEntrada = dtEnt.Time.Format("02/01/2006 15:04:05")

		if ctiPersonalizado.Valid {
			i.Descricao = ctiPersonalizado.String
			i.Grupo = ctiPadraoGrupo.String

		} else {
			i.Descricao = ctiPadrao.String
		}

		if i.Grupo == "ARME" || i.Grupo == "DESARME" || i.Grupo == "PANICO" {
			i.UserSet = usuarioAlarme.String
		} else {
			i.UserSet = setorAlarme.String
		}

		*lista = append(*lista, i)
	}
	return nil
}
