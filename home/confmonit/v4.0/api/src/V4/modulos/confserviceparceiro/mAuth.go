package confserviceparceiroV4

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type authPayload struct {
	AdminToken        string   `json:"admin_token"`
	IDUsuario         string   `json:"idUsuario"`
	UserTipo          string   `json:"userTipo"`
	IDVinculo         string   `json:"idVinculo"`
	IDCentralCatalogo string   `json:"idCentralCatalogo"`
	Breakglass        bool     `json:"breakglass"`
	IDRepresentante   string   `json:"idRepresentante"`
	IDFranqueado      string   `json:"idFranqueado"`
	IDsParceiro       []string `json:"idsParceiro"`
}

func readAuthPayload(r *http.Request) (*authPayload, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var p authPayload
	if len(body) > 0 {
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, errors.New("json invalido")
		}
	}
	return &p, nil
}

func resolveEscopo(p *authPayload) (*admEscopo, error) {
	if p == nil {
		return nil, errors.New("sessao invalida")
	}
	token := strings.TrimSpace(p.AdminToken)
	if token == "" {
		return nil, errors.New("sessao invalida — faca login novamente")
	}

	userTipo := strings.ToUpper(strings.TrimSpace(p.UserTipo))
	idVinculo := strings.TrimSpace(p.IDVinculo)
	idCentral := strings.TrimSpace(p.IDCentralCatalogo)
	idUsuario := strings.TrimSpace(p.IDUsuario)

	if idUsuario == "BREAKGLASS" || p.Breakglass {
		if userTipo != "CEN" && userTipo != "REP" {
			return nil, errors.New("sessao sem perfil financeiro valido (CEN/REP)")
		}
		esc := &admEscopo{
			UserTipo:          userTipo,
			IDUsuario:         idUsuario,
			IDVinculo:         idVinculo,
			IDCentralCatalogo: idCentral,
		}
		if userTipo == "CEN" {
			if idCentral == "" {
				idCentral = idVinculo
			}
			if idCentral == "" {
				return nil, errors.New("Central nao selecionada (break-glass)")
			}
			esc.IDVinculo = idCentral
			esc.IDCentralCatalogo = idCentral
		} else if idVinculo == "" {
			return nil, errors.New("sessao Representante sem idVinculo")
		} else if idCentral == "" {
			return nil, errors.New("sessao Representante sem idCentral")
		}
		return esc, nil
	}

	if idUsuario == "" {
		return nil, errors.New("idUsuario obrigatorio na sessao")
	}

	escDB, err := carregarEscopoAdm(idUsuario)
	if err != nil {
		return nil, err
	}

	if userTipo != "" && userTipo != escDB.UserTipo {
		return nil, errors.New("sessao inconsistente (userTipo)")
	}

	if escDB.UserTipo == "REP" {
		if idVinculo != "" && idVinculo != escDB.IDVinculo {
			return nil, errors.New("sessao inconsistente (idVinculo)")
		}
		if idCentral != "" && idCentral != escDB.IDCentralCatalogo {
			return nil, errors.New("sessao inconsistente (idCentral)")
		}
		return escDB, nil
	}

	// CEN: permite idCentralCatalogo do body (break-glass sidebar na mesma sessao)
	if idCentral != "" {
		escDB.IDCentralCatalogo = idCentral
		escDB.IDVinculo = idCentral
	}
	return escDB, nil
}

func escopoCentralFromPayload(p *authPayload) (*admEscopo, error) {
	esc, err := resolveEscopo(p)
	if err != nil {
		return nil, err
	}
	if esc.UserTipo != "CEN" {
		return nil, errors.New("acesso apenas para Central")
	}
	if esc.IDCentralCatalogo == "" {
		return nil, errors.New("sessao Central sem idCentral")
	}
	return esc, nil
}

func escopoRepresentanteFromPayload(p *authPayload) (*admEscopo, error) {
	esc, err := resolveEscopo(p)
	if err != nil {
		return nil, err
	}
	if esc.UserTipo != "REP" {
		return nil, errors.New("acesso apenas para Representante")
	}
	if esc.IDVinculo == "" {
		return nil, errors.New("sessao Representante sem idVinculo")
	}
	return esc, nil
}
