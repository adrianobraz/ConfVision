package roteador

import (
	"log"
	"net/http"
	"strings"

	permissoesAcesso "franqueadopro/recursos/modulos/minha-empresa/permissoes-acesso"
	"franqueadopro/src/licenca"
	"franqueadopro/src/seguranca"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n %s %s %s", r.Method, r.RequestURI, r.Host)
		next(w, r)
	}
}

func Autenticar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, erro := seguranca.LerCookies(r)
		if erro != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if r.Method == http.MethodGet {
			path := strings.Split(r.URL.Path, "?")[0]
			path = strings.TrimSuffix(path, "/")
			if path == "" {
				path = "/"
			}
			tab := r.URL.Query().Get("tab")

			est := licenca.ResolverEstadoNavegacao(cookie["idFranqueado"], "franqueadopro")
			if licenca.DeveBloquearAcesso(est) {
				log.Printf("licenca bloqueada: franqueado=%s path=%s motivo=%s", cookie["idFranqueado"], path, est.Motivo)
				if !licenca.RotaPermitidaQuandoBloqueado(path, tab, est) {
					http.Redirect(w, r, licenca.URLRedirecionamentoBloqueio(est), http.StatusFound)
					return
				}
				next(w, r)
				return
			}

			// Rotas livres (ex.: Meu Plano) nao passam por gate de modulo — evita loop de redirect
			if licenca.RotasLivres(path) {
				next(w, r)
				return
			}

			if chave, ok := permissoesAcesso.ChavePorRotaComTab(path, tab); ok {
				if cookie["master"] != "S" {
					chaves, err := permissoesAcesso.CarregarPermissoesUsuario(
						cookie["idFranqueado"],
						cookie["idUsuario"],
					)
					if err == nil && len(chaves) > 0 {
						if !permissoesAcesso.TemPermissaoExata(chaves, chave) &&
							!permissoesAcesso.TemPermissao(chaves, chave) {
							http.Redirect(w, r, "/carregar-menu-principal", http.StatusFound)
							return
						}
					}
				}

				okMod := licenca.ModuloLiberadoNoEstado(est, chave)
				if !okMod {
					http.Redirect(w, r, "/carregar-menu-principal", http.StatusFound)
					return
				}
			}
		}

		next(w, r)
	}
}
