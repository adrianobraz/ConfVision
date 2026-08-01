package feedback

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"errors"
	"strings"
)

type Feedback struct {
	ID_Feedback   string `json:"idFeedback"`
	Software      string `json:"software"`
	Tipo          string `json:"tipo"`
	Descricao     string `json:"descricao"`
	URL_Pagina    string `json:"urlPagina"`
	ID_Usuario    string `json:"idUsuario"`
	ID_Franqueado string `json:"idFranqueado"`
}

func (f *Feedback) Enviar() error {
	f.Software = strings.ToLower(strings.TrimSpace(f.Software))
	f.Tipo = strings.ToLower(strings.TrimSpace(f.Tipo))
	f.Descricao = strings.TrimSpace(f.Descricao)
	f.URL_Pagina = strings.TrimSpace(f.URL_Pagina)
	f.ID_Usuario = strings.TrimSpace(f.ID_Usuario)
	f.ID_Franqueado = strings.TrimSpace(f.ID_Franqueado)

	if f.Software == "" {
		return errors.New("software obrigatorio")
	}
	validosSoft := map[string]bool{
		"franqueadopro": true,
		"confvision":    true,
		"webambiente":   true,
		"dialyze":       true,
	}
	if !validosSoft[f.Software] {
		return errors.New("software invalido")
	}

	if f.Tipo != "bug" && f.Tipo != "melhoria" && f.Tipo != "ideia" {
		return errors.New("tipo invalido (bug|melhoria|ideia)")
	}

	if len(f.Descricao) < 5 {
		return errors.New("descricao obrigatorio (minimo 5 caracteres)")
	}

	if f.ID_Feedback == "" {
		f.ID_Feedback = idCurto()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO cm_feedback_cliente (
			ID_Feedback, Software, Tipo, Descricao, URL_Pagina,
			ID_Usuario, ID_Franqueado
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		f.ID_Feedback,
		f.Software,
		f.Tipo,
		f.Descricao,
		nullIfEmpty(f.URL_Pagina),
		nullIfEmpty(f.ID_Usuario),
		nullIfEmpty(f.ID_Franqueado),
	)
	return err
}

func idCurto() string {
	id := auxiliar.GeradorDeId()
	if len(id) > 20 {
		return id[len(id)-20:]
	}
	return id
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
