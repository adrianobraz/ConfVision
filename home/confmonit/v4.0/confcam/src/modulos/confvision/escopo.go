package confvision

import (
	"confvision/src/modulos/visdata"
	"confvision/src/seguranca"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const cameraPingStaleMinutes = 15

// FranqueadoDaSessao devolve id_franqueado exclusivamente do cookie de sessão.
func FranqueadoDaSessao(r *http.Request) (string, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return "", fmt.Errorf("sessao invalida")
	}
	if seguranca.EhAdministrator(cookie) {
		return "", nil
	}
	id := strings.TrimSpace(seguranca.IdFranqueadoDoCookie(cookie))
	if id == "" {
		return "", fmt.Errorf("id_franqueado ausente na sessao")
	}
	return id, nil
}

// FranqueadoDaSessaoOuVazio — ADM não tem tenant; FRA/CLI exige id.
func FranqueadoDaSessaoOuVazio(r *http.Request) (string, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return "", fmt.Errorf("sessao invalida")
	}
	if seguranca.EhAdministrator(cookie) {
		return "", nil
	}
	return FranqueadoDaSessao(r)
}

func assertFranqueadoQuery(r *http.Request, idQuery string) (string, error) {
	idSessao, err := FranqueadoDaSessao(r)
	if err != nil {
		return "", err
	}
	idQuery = strings.TrimSpace(idQuery)
	if idQuery != "" && idQuery != idSessao {
		return "", fmt.Errorf("id_franqueado nao autorizado")
	}
	return idSessao, nil
}

func cameraPertenceAoFranqueado(cameraID, idFranqueado string) bool {
	cameraID = strings.TrimSpace(cameraID)
	idFranqueado = strings.TrimSpace(idFranqueado)
	if cameraID == "" || idFranqueado == "" {
		return false
	}
	id, err := strconv.Atoi(cameraID)
	if err != nil || id < 1 {
		return false
	}
	cam, err := visdata.GetCameraByID(context.Background(), id)
	if err != nil {
		cam, err = fetchVisCameraMap(id)
		if err != nil || cam == nil {
			return false
		}
	}
	idFra := strings.TrimSpace(fmt.Sprint(cam["id_franqueado"]))
	return idFra != "" && idFra != "<nil>" && idFra == idFranqueado
}

func fetchVisCameraMap(cameraID int) (map[string]any, error) {
	cam, ok := fetchVisCamera(cameraID)
	if !ok {
		return nil, fmt.Errorf("camera nao encontrada")
	}
	return cam, nil
}

func cameraIDFranqueado(cameraID int) (string, bool) {
	cam, err := visdata.GetCameraByID(context.Background(), cameraID)
	if err != nil {
		m, err2 := fetchVisCameraMap(cameraID)
		if err2 != nil {
			return "", false
		}
		cam = m
	}
	idFra := strings.TrimSpace(fmt.Sprint(cam["id_franqueado"]))
	if idFra == "" || idFra == "<nil>" {
		return "", false
	}
	return idFra, true
}

func responderEscopoProibido(w http.ResponseWriter, msg string) {
	http.Error(w, `{"status":"`+msg+`"}`, http.StatusForbidden)
}
