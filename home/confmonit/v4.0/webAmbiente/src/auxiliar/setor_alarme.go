package auxiliar

import (
	"database/sql"
	"strings"
)

// Icone do setor no mapa: usa somente setorAlarme.TipoSetor (CAMERA/SENSOR).
// O campo legado setorAlarme.camera (S/N) e ignorado de proposito.

func IconePorTipoSetor(tipo string) string {
	switch strings.ToUpper(strings.TrimSpace(tipo)) {
	case "CAMERA":
		return "camera-video-fill"
	default:
		return "bullseye"
	}
}

func ErroColunaTipoSetor(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown column") && strings.Contains(msg, "tiposetor")
}

func ResolverIconeSetor(tipoSetor string) string {
	return IconePorTipoSetor(tipoSetor)
}

// ResolverIconeSalvo usa o ícone gravado no mapa; se vazio, cai no padrão por tipo.
func ResolverIconeSalvo(iconeSalvo, tipoSetor string) string {
	if strings.TrimSpace(iconeSalvo) != "" {
		return strings.TrimSpace(iconeSalvo)
	}
	return ResolverIconeSetor(tipoSetor)
}

type DadosSetorAlarme struct {
	Camera    string
	Numero    string
	Particao  string
	TipoSetor string
}

func SetorTemCamera(camera string) bool {
	return strings.ToUpper(strings.TrimSpace(camera)) == "S"
}

func BuscarDadosSetorAlarme(idSetor string) (DadosSetorAlarme, error) {
	idSetor = strings.ToUpper(strings.TrimSpace(idSetor))
	if idSetor == "" {
		return DadosSetorAlarme{}, nil
	}

	db, err := Conectar()
	if err != nil {
		return DadosSetorAlarme{}, nil
	}
	defer db.Close()

	var out DadosSetorAlarme
	var camera, numero, particao, tipoSetor sql.NullString
	err = db.QueryRow(`
		SELECT setorAlarme.Camera, setorAlarme.Numero, setorAlarme.Particao, setorAlarme.TipoSetor
		FROM setorAlarme
		WHERE setorAlarme.ID_Setor = ?
	`, idSetor).Scan(&camera, &numero, &particao, &tipoSetor)
	if err != nil {
		if err == sql.ErrNoRows {
			return out, nil
		}
		if ErroColunaTipoSetor(err) {
			return buscarDadosSetorSemTipoSetor(db, idSetor)
		}
		return out, err
	}

	out.Camera = strings.TrimSpace(camera.String)
	out.Numero = strings.TrimSpace(numero.String)
	out.Particao = strings.TrimSpace(particao.String)
	out.TipoSetor = strings.TrimSpace(tipoSetor.String)
	return out, nil
}

func buscarDadosSetorSemTipoSetor(db *sql.DB, idSetor string) (DadosSetorAlarme, error) {
	var out DadosSetorAlarme
	var camera, numero, particao sql.NullString
	err := db.QueryRow(`
		SELECT setorAlarme.Camera, setorAlarme.Numero, setorAlarme.Particao
		FROM setorAlarme
		WHERE setorAlarme.ID_Setor = ?
	`, idSetor).Scan(&camera, &numero, &particao)
	if err != nil {
		if err == sql.ErrNoRows {
			return out, nil
		}
		return out, err
	}
	out.Camera = strings.TrimSpace(camera.String)
	out.Numero = strings.TrimSpace(numero.String)
	out.Particao = strings.TrimSpace(particao.String)
	return out, nil
}

func BuscarTipoSetor(idSetor string) (string, error) {
	dados, err := BuscarDadosSetorAlarme(idSetor)
	if err != nil {
		return "", err
	}
	return dados.TipoSetor, nil
}
