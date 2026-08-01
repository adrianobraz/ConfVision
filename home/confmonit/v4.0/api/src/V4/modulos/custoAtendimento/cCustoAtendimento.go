package custoAtendimentoV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	reIdFra = regexp.MustCompile(`(?i)ID_Franqueado\s*=\s*'([^']+)'`)
	reIdCli = regexp.MustCompile(`(?i)ID_Cliente\s*=\s*'([^']+)'`)
	reDi    = regexp.MustCompile(`(?i)DataOperacao\s*>=\s*'([^']+)'`)
	reDf    = regexp.MustCompile(`(?i)DataOperacao\s*<=\s*'([^']+)'`)
)

func listarByFiltro(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IDFranqueado string `json:"idFranqueado"`
		IDCliente    string `json:"idCliente"`
		DataInicio   string `json:"dataInicio"`
		DataFim      string `json:"dataFim"`
		Filtro       string `json:"filtro"`
		Limit        int    `json:"limit"`
		Offset       int    `json:"offset"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	f := FiltroListar{
		IDFranqueado: strings.TrimSpace(req.IDFranqueado),
		IDCliente:    strings.TrimSpace(req.IDCliente),
		DataInicio:   strings.TrimSpace(req.DataInicio),
		DataFim:      strings.TrimSpace(req.DataFim),
		Limit:        req.Limit,
		Offset:       req.Offset,
	}

	// Compatibilidade com o payload antigo (SQL filtro)
	if f.IDFranqueado == "" && strings.TrimSpace(req.Filtro) != "" {
		parseFiltroLegado(req.Filtro, &f)
	}

	lista, total, err := ListarLigacoes(f)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Lista(w, http.StatusOK, lista, req.Limit, req.Offset, total, len(lista))
}

func parseFiltroLegado(filtro string, f *FiltroListar) {
	if m := reIdFra.FindStringSubmatch(filtro); len(m) > 1 {
		f.IDFranqueado = m[1]
	}
	if m := reIdCli.FindStringSubmatch(filtro); len(m) > 1 {
		f.IDCliente = m[1]
	}
	if m := reDi.FindStringSubmatch(filtro); len(m) > 1 {
		f.DataInicio = m[1]
	}
	if m := reDf.FindStringSubmatch(filtro); len(m) > 1 {
		f.DataFim = m[1]
	}
}
