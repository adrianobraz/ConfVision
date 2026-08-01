package centralV4

import (
	"api/src/V4/config"
	usuariosV4 "api/src/V4/modulos/usuarios"
	"api/src/V4/seguranca"
	"api/src/auxiliar"
	"errors"
	"strings"
)

type centralLogin struct {
	Token string `json:"token"`
	usuariosV4.Usuario
}

func (cl *centralLogin) logar(email, senha string) error {

	// Valida se um email foi informado
	if email == "" {
		return errors.New("um email deve ser informado")

	}

	// Valida se uma senha foi informada
	if senha == "" {
		return errors.New("uma senha deve ser informada")

	}

	var err error
	cl.Email1 = email
	if err := cl.GetDadosByEmail1(); err != nil {
		return err
	}

	// Valida a senha
	if err = seguranca.VerificarSenha(cl.Senha, senha); err != nil {
		return err
	}

	// Valida se usuario nao esta bloqueado
	if cl.Ativo == "N" {
		return errors.New("usuario bloqueado")
	}

	// Isolamento multi-tenant: CENTRAL precisa de IDCentralUUID na sessao
	if cl.ID_Vinculo == "CENTRAL" {
		cl.IDCentralUUID = auxiliar.GarantirIDCentralUUID(cl.IDCentralUUID)
		if cl.IDCentralUUID == "" {
			return errors.New("usuario sem IDCentralUUID — rode o SQL de multi-tenant / vincule a Central")
		}
	}

	// Caso a senha esteja correta ele cria um token
	cl.Token, err = seguranca.CriarToken(cl.ID_Usuario)
	if err != nil {
		return err
	}

	// Retorna os dados para o requisitante
	return nil
}

// logarAdministrador autentica o break-glass do portal Central.
func (cl *centralLogin) logarAdministrador(usuario, senha string) error {
	usuario = strings.TrimSpace(usuario)
	senha = strings.TrimSpace(senha)

	if usuario == "" {
		return errors.New("um usuario deve ser informado")
	}
	if senha == "" {
		return errors.New("uma senha deve ser informada")
	}

	if usuario != config.CentralBreakglassUser || senha != config.CentralBreakglassPass {
		return errors.New("usuario ou senha invalidos")
	}

	cl.ID_Usuario = "BREAKGLASS"
	cl.ID_Vinculo = "CENTRAL"
	cl.IDCentralUUID = auxiliar.GarantirIDCentralUUID("")
	if cl.IDCentralUUID == "" {
		return errors.New("central legada sem IDCentralUUID — rode o SQL de multi-tenant")
	}
	cl.Tipo = "CEN"
	cl.Nome = "ADMINISTRATOR"
	cl.Nick = "ADMIN"
	cl.Email1 = usuario
	cl.UsuarioWeb = "S"
	cl.UsuarioTeminal = "N"
	cl.Master = "S"
	cl.Ativo = "S"

	token, err := seguranca.CriarToken(cl.ID_Usuario)
	if err != nil {
		return err
	}
	cl.Token = token
	cl.Senha = ""
	return nil
}
