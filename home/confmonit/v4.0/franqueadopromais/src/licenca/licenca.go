package licenca

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"franqueadopro/src/xanopro"
)

type Estado struct {
	Liberado  bool            `json:"liberado"`
	Motivo    string          `json:"motivo"`
	Produto   string          `json:"produto"`
	Plano     string          `json:"plano"`
	Retencao  int             `json:"retencao_dias"`
	Modulos   map[string]bool `json:"modulos_json"`
	Limites   map[string]int  `json:"limites_json"`
}

type cacheEntry struct {
	estado    Estado
	expiresAt time.Time
}

var (
	cacheMu sync.RWMutex
	cache   = map[string]cacheEntry{}
	ttl     = 30 * time.Second
)

func cacheKey(idFranqueado, produto string) string {
	return strings.TrimSpace(idFranqueado) + "|" + strings.TrimSpace(produto)
}

func parseEstado(raw []byte, produto string) (Estado, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return Estado{}, err
	}

	ass := mapObj(resp["assinatura"])

	liberado := parseLiberadoJSON(resp["liberado"])
	motivo := campoString(resp["motivo"])
	if !liberado && ass != nil {
		liberado, motivo = inferirLiberadoAssinatura(
			campoString(ass["status"]),
			ass["valido_ate"],
			motivo,
		)
	}

	plano := ""
	if ass != nil {
		plano = campoString(ass["plano"])
	}

	modRaw := mapObj(resp["modulos_json"])
	if modRaw == nil && ass != nil {
		modRaw = mapObj(ass["modulos_json"])
	}
	limRaw := mapObj(resp["limites_json"])
	if limRaw == nil && ass != nil {
		limRaw = mapObj(ass["limites_json"])
	}

	modFlat := flattenModulosJSON(modRaw)
	normalizarChaveLegacyModulo(modFlat)

	if liberado {
		pendentes := sliceTextoJSON(sliceListaJSON(resp["addons_pendentes"]))
		if len(pendentes) == 0 && ass != nil {
			pendentes = sliceTextoJSON(sliceListaJSON(ass["addons_pendentes_json"]))
		}
		contratados := sliceTextoJSON(sliceListaJSON(resp["addons_contratados"]))
		if len(contratados) == 0 && ass != nil {
			contratados = sliceTextoJSON(sliceListaJSON(ass["addons_json"]))
		}
		mesclarAddonsPagos(modFlat, contratados, pendentes)
		mesclarAddonsEfetivos(modFlat, sliceTextoJSON(sliceListaJSON(resp["addons_efetivos"])))
	}

	est := Estado{
		Liberado: liberado,
		Motivo:   motivo,
		Produto:  produto,
		Plano:    plano,
		Modulos:  mergeModulos(plano, modFlat),
		Limites:  mergeLimites(plano, parseLimitesJSON(limRaw)),
	}
	if p := campoString(resp["produto"]); p != "" {
		est.Produto = p
	}
	if rd := campoInt(resp["retencao_dias"]); rd > 0 {
		est.Retencao = rd
	} else {
		est.Retencao = retencaoPorPlano(est.Plano)
	}
	return est, nil
}

func mapObj(v interface{}) map[string]interface{} {
	m, ok := v.(map[string]interface{})
	if ok {
		return m
	}
	return nil
}

func sliceListaJSON(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []interface{}:
		return t
	case map[string]interface{}:
		return []interface{}{}
	default:
		return nil
	}
}

func campoString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func campoInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	}
	return 0
}

func retencaoPorPlano(plano string) int {
	switch plano {
	case "lite":
		return 30
	case "pro":
		return 90
	case "pro_plus":
		return 180
	default:
		return 30
	}
}

func parseLiberadoJSON(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		return s == "true" || s == "1" || s == "ok" || s == "sim"
	default:
		return false
	}
}

func inferirLiberadoAssinatura(status string, validoAte interface{}, motivo string) (bool, string) {
	if !strings.EqualFold(strings.TrimSpace(status), "ativa") {
		return false, motivo
	}
	if assinaturaVencida(validoAte) {
		return false, motivo
	}
	if motivo == "" || motivo == "sem_assinatura" || motivo == "pendente" || motivo == "xano_indisponivel" || motivo == "parse_erro" {
		return true, "ok"
	}
	return true, motivo
}

func assinaturaVencida(validoAte interface{}) bool {
	if validoAte == nil {
		return false
	}
	switch x := validoAte.(type) {
	case float64:
		if x > 1e12 {
			return time.UnixMilli(int64(x)).Before(time.Now())
		}
		return time.Unix(int64(x), 0).Before(time.Now())
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return false
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.Before(time.Now())
		}
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t.Before(time.Now())
		}
	}
	return false
}

// EhChaveMenuGrupo indica menu pai de area (entrada) — nao bloqueia o card do menu principal.
func EhChaveMenuGrupo(chave string) bool {
	grupos := map[string]bool{
		"configuracao":  true,
		"atendimento":   true,
		"minha-empresa": true,
		"relatorio":     true,
	}
	return grupos[chave]
}

// Verificar consulta licenca no Xano (com cache curto).
func Verificar(idFranqueado, produto string) (Estado, error) {
	return verificar(idFranqueado, produto, false)
}

// VerificarAtual ignora cache — uso em middleware e apos alteracao de plano.
func VerificarAtual(idFranqueado, produto string) (Estado, error) {
	return verificar(idFranqueado, produto, true)
}

func verificar(idFranqueado, produto string, forcarAtual bool) (Estado, error) {
	key := cacheKey(idFranqueado, produto)
	if !forcarAtual {
		cacheMu.RLock()
		if e, ok := cache[key]; ok && time.Now().Before(e.expiresAt) {
			cacheMu.RUnlock()
			return e.estado, nil
		}
		cacheMu.RUnlock()
	}

	raw, err := postLicencaFranqueado(idFranqueado, produto)
	if err != nil {
		return Estado{Liberado: false, Motivo: "xano_indisponivel"}, nil
	}

	est, err := parseEstado(raw, produto)
	if err != nil {
		return Estado{Liberado: false, Motivo: "parse_erro"}, nil
	}

	gravarCache(key, est)
	return est, nil
}

// postLicencaFranqueado consulta o Xano; usa resumo (mesma fonte da tela Meu Plano) com fallback em verificar.
func postLicencaFranqueado(idFranqueado, produto string) ([]byte, error) {
	payload := map[string]string{
		"id_franqueado": idFranqueado,
		"produto":       produto,
	}
	raw, err := xanopro.Post("/fp_assinatura_resumo", payload)
	if err == nil && len(raw) > 0 {
		return raw, nil
	}
	return xanopro.Post("/fp_licenca_verificar", payload)
}

// SincronizarCache atualiza o cache a partir da resposta JSON do Xano (ex.: assinatura_resumo).
func SincronizarCache(idFranqueado, produto string, raw []byte) {
	est, err := parseEstado(raw, produto)
	if err != nil {
		return
	}
	gravarCache(cacheKey(idFranqueado, produto), est)
}

// SincronizarCacheResumo atualiza cache com TTL estendido apos licencaResumo (evita divergencia no middleware).
func SincronizarCacheResumo(idFranqueado, produto string, raw []byte) {
	est, err := parseEstado(raw, produto)
	if err != nil {
		return
	}
	cacheMu.Lock()
	cache[cacheKey(idFranqueado, produto)] = cacheEntry{estado: est, expiresAt: time.Now().Add(5 * time.Minute)}
	cacheMu.Unlock()
}

func gravarCache(key string, est Estado) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if prev, ok := cache[key]; ok && prev.estado.Liberado && !est.Liberado {
		switch est.Motivo {
		case "xano_indisponivel", "parse_erro":
			if time.Now().Before(prev.expiresAt) {
				return
			}
		}
	}
	cache[key] = cacheEntry{estado: est, expiresAt: time.Now().Add(ttl)}
}

// ModuloLiberadoNoEstado verifica chave contra estado ja carregado (evita cache divergente).
func ModuloLiberadoNoEstado(est Estado, chaveMenu string) bool {
	if !est.Liberado {
		return false
	}
	if EhChaveMenuGrupo(chaveMenu) {
		return true
	}
	if moduloLiberadoNoMapa(est.Modulos, chaveMenu) {
		return true
	}
	// Central de Disparos: Pro+ (chave propria) OU addon Inteligencia Artificial
	if strings.HasPrefix(chaveMenu, "relatorio.central-disparos") {
		if moduloLiberadoNoMapa(est.Modulos, "relatorio.central-disparos") {
			return true
		}
		if moduloLiberadoNoMapa(est.Modulos, "atendimento.inteligencia-artificial") {
			return true
		}
		return false
	}
	return false
}

// ModuloLiberado verifica chave de menu contra plano (consulta Xano com cache curto).
func ModuloLiberado(idFranqueado, chaveMenu string) (bool, string) {
	est, _ := VerificarAtual(idFranqueado, "franqueadopro")
	if !est.Liberado {
		return false, est.Motivo
	}
	if EhChaveMenuGrupo(chaveMenu) {
		return true, ""
	}
	if ModuloLiberadoNoEstado(est, chaveMenu) {
		return true, ""
	}
	return false, "modulo_bloqueado_plano"
}

// DeveBloquearAcesso indica redirect para Meu Plano / Faturas.
func DeveBloquearAcesso(est Estado) bool {
	if est.Liberado {
		return false
	}
	switch est.Motivo {
	case "sem_assinatura", "vencida", "suspensa", "pendente", "modulo_bloqueado_plano", "xano_indisponivel", "parse_erro":
		return true
	default:
		return true
	}
}

// DevePrenderNaFatura — suspensão/vencimento: usuário só acessa a aba de faturas até regularizar.
func DevePrenderNaFatura(est Estado) bool {
	if est.Liberado {
		return false
	}
	switch est.Motivo {
	case "suspensa", "vencida", "pendente":
		return true
	default:
		return false
	}
}

const URLFaturasBloqueio = "/carregar-meu-plano?tab=faturas"

// URLRedirecionamentoBloqueio destino do redirect quando o acesso global está bloqueado.
func URLRedirecionamentoBloqueio(est Estado) string {
	if DevePrenderNaFatura(est) {
		return URLFaturasBloqueio
	}
	return "/carregar-meu-plano"
}

func normalizarPath(path string) string {
	path = strings.ToLower(strings.Split(path, "?")[0])
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}
	return path
}

// RotaPermitidaQuandoBloqueado rotas acessíveis com licença inativa.
func RotaPermitidaQuandoBloqueado(path, tab string, est Estado) bool {
	path = normalizarPath(path)
	livres := []string{
		"/logout",
		"/licencaResumo",
		"/licencaMinhasFaturas",
		"/licencaContratar",
		"/licencaCupomValidar",
		"/licencaCupomAplicarFatura",
		"/licencaEcossistema",
		"/licencaContratarProduto",
		"/carregar-alterar-senha",
		"/carregar-dados",
	}
	for _, r := range livres {
		if path == r {
			return true
		}
	}
	if path == "/carregar-meu-plano" {
		if DevePrenderNaFatura(est) {
			return strings.EqualFold(tab, "faturas")
		}
		return true
	}
	return false
}

// RotasLivres nao exigem licenca ativa (quando o acesso global está liberado).
func RotasLivres(path string) bool {
	path = normalizarPath(path)
	livres := []string{
		"/logout",
		"/licencaResumo",
		"/licencaMinhasFaturas",
		"/licencaContratar",
		"/licencaCupomValidar",
		"/licencaCupomAplicarFatura",
		"/licencaEcossistema",
		"/licencaContratarProduto",
		"/carregar-alterar-senha",
		"/carregar-dados",
	}
	for _, r := range livres {
		if path == r {
			return true
		}
	}
	return false
}

// ResolverEstadoNavegacao usa cache liberado pelo /licencaResumo; so reconsulta Xano se cache indicar bloqueio.
func ResolverEstadoNavegacao(idFranqueado, produto string) Estado {
	est, _ := Verificar(idFranqueado, produto)
	if est.Liberado {
		return est
	}
	est, _ = VerificarAtual(idFranqueado, produto)
	return est
}

// InvalidarCache limpa cache apos alteracao de plano.
func InvalidarCache(idFranqueado, produto string) {
	cacheMu.Lock()
	delete(cache, cacheKey(idFranqueado, produto))
	cacheMu.Unlock()
}
