package contrato

import (
	"apifunction/auth"
	"errors"
	"strings"
)

// AssertPodeEditarContrato valida UsaAdmConfmonit + carteira REP.
func AssertPodeEditarContrato(sess auth.SessaoAdm, idFranqueado string) error {
	if strings.TrimSpace(sess.IDUsuario) == "BREAKGLASS" {
		return nil
	}
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return errors.New("id_franqueado obrigatorio")
	}
	idRep, _, err := franqueadoRepresentante(idFranqueado)
	if err != nil {
		return err
	}
	usaAdm, err := representanteUsaAdm(idRep)
	if err != nil {
		return err
	}

	switch sess.UserTipo {
	case "CEN":
		if usaAdm == "S" {
			return errors.New("contrato gerido pelo representante (UsaAdmConfmonit = S)")
		}
		return nil
	case "REP":
		if usaAdm != "S" {
			return errors.New("contrato gerido pela central (UsaAdmConfmonit = N)")
		}
		if sess.IDRepresentante != idRep {
			return errors.New("franqueado fora da carteira deste representante")
		}
		return nil
	default:
		return errors.New("sem permissao")
	}
}
