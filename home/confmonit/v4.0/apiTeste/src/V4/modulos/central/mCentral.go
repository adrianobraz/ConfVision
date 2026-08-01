package centralV4

import (
	usuariosV4 "api/src/V4/modulos/usuarios"
	"api/src/V4/seguranca"
	"errors"
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

	// Caso a senha esteja correta ele cria um token
	cl.Token, err = seguranca.CriarToken(cl.ID_Usuario)
	if err != nil {
		return err
	}

	// Retorna os dados para o requisitante
	return nil
}
