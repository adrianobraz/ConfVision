package permissoesAcesso

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/xano"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-permissoes-acesso",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/permissoesListarByUsuario",
		Metodo: http.MethodPost,
		Funcao: listarByUsuario,
		Aberto: false,
	},
	{
		URI:    "/permissoesSalvar",
		Metodo: http.MethodPost,
		Funcao: salvar,
		Aberto: false,
	},
	{
		URI:    "/permissoesCarregarLogado",
		Metodo: http.MethodPost,
		Funcao: carregarLogado,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if !ehMaster(r) {
		http.Redirect(w, r, "/carregar-menu-gestao", http.StatusFound)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina
		d.TituloSite = config.TituloSite
		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
		return
	}

	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Permissões de Acesso"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-gestao"
	auxiliar.ExecutarTemplate(w, "permissoes-acesso.html", d)
}

func listarByUsuario(w http.ResponseWriter, r *http.Request) {
	proxyXano(w, r, "/fp_permissao_listar_by_usuario")
}

func carregarLogado(w http.ResponseWriter, r *http.Request) {
	proxyXano(w, r, "/fp_permissao_get_by_usuario")
}

func salvar(w http.ResponseWriter, r *http.Request) {
	if !ehMaster(r) {
		auxiliar.RespostaErro(w, http.StatusForbidden, fmt.Errorf("apenas usuario master pode alterar permissoes"))
		return
	}
	proxyXano(w, r, "/fp_permissao_salvar")
}

func proxyXano(w http.ResponseWriter, r *http.Request, path string) {
	if config.XanoApi == "" {
		auxiliar.RespostaErro(w, http.StatusBadGateway, fmt.Errorf("XANO_API_FRANQUEADO nao configurado no servidor"))
		return
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var payload map[string]interface{}
	if len(body) > 0 {
		if erro = json.Unmarshal(body, &payload); erro != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
	}

	raw, erro := xano.Post(path, payload)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, erro)
		return
	}

	var parsed interface{}
	if erro = json.Unmarshal(raw, &parsed); erro != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
		return
	}

	auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
		"status": "OK",
		"dados":  parsed,
	})
}

func ehMaster(r *http.Request) bool {
	return auxiliar.MasterDoCookie(r) == "S"
}

// CarregarPermissoesUsuario busca permissoes liberadas no Xano para middleware
func CarregarPermissoesUsuario(idFranqueado, idUsuario string) (map[string]bool, error) {
	chaves := map[string]bool{}

	if config.XanoApi == "" {
		return chaves, nil
	}

	raw, err := xano.Post("/fp_permissao_get_by_usuario", map[string]string{
		"idFranqueado": idFranqueado,
		"idUsuario":    idUsuario,
	})
	if err != nil {
		return nil, err
	}

	var lista []struct {
		ChaveMenu string `json:"chave_menu"`
		Liberado  string `json:"liberado"`
	}
	if err := json.Unmarshal(raw, &lista); err != nil {
		var wrapped struct {
			Dados []struct {
				ChaveMenu string `json:"chave_menu"`
				Liberado  string `json:"liberado"`
			} `json:"dados"`
		}
		if err2 := json.Unmarshal(raw, &wrapped); err2 != nil {
			return nil, err
		}
		for _, p := range wrapped.Dados {
			if p.Liberado == "S" {
				chaves[p.ChaveMenu] = true
			}
		}
		return chaves, nil
	}

	for _, p := range lista {
		if p.Liberado == "S" {
			chaves[p.ChaveMenu] = true
		}
	}
	return chaves, nil
}

// TemPermissao verifica chave exata ou se algum filho esta liberado (para menus pai)
func TemPermissao(chaves map[string]bool, chave string) bool {
	if len(chaves) == 0 {
		return true
	}
	if chaves[chave] {
		return true
	}
	prefix := chave + "."
	for k := range chaves {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// TemPermissaoExata exige a chave especifica (paginas internas)
func TemPermissaoExata(chaves map[string]bool, chave string) bool {
	if len(chaves) == 0 {
		return true
	}
	return chaves[chave]
}
