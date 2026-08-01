package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Rotina ligacao060: migracao da task Xano "Ligacao_0" (freq 60s na origem).
// Mutex proprio (runningLigacao060), independente da fila FilaLigacao.

func runLigacao060Loop(client *http.Client, cfg Config, shutdown <-chan struct{}) {
	ticker := time.NewTicker(cfg.Ligacao060Interval)
	defer ticker.Stop()
	runLigacao060Once(client, cfg)
	for {
		select {
		case <-shutdown:
			log.Printf("ligacao060: loop encerrado")
			return
		case <-ticker.C:
			runLigacao060Once(client, cfg)
		}
	}
}

func runLigacao060Once(client *http.Client, cfg Config) {
	if !atomic.CompareAndSwapInt32(&runningLigacao060, 0, 1) {
		log.Printf("ligacao060: execucao anterior em andamento, pulando ciclo")
		return
	}
	defer atomic.StoreInt32(&runningLigacao060, 0)

	start := time.Now()
	attempts := cfg.Retries + 1
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			backoff := backoffForAttempt(attempt)
			log.Printf("ligacao060 retry=%d backoff=%s", attempt-1, backoff)
			time.Sleep(backoff)
		}
		err := processLigacao060Routine(client, cfg)
		if err == nil {
			log.Printf("ligacao060 sucesso | attempt=%d/%d | duracao=%s", attempt, attempts, time.Since(start))
			return
		}
		lastErr = err
		log.Printf("ligacao060 falha | attempt=%d/%d | erro=%v", attempt, attempts, err)
	}
	log.Printf("ligacao060 erro final | duracao=%s | erro=%v", time.Since(start), lastErr)
}

type whatsappLigarErroRow060 struct {
	ID              int
	Tentativas      int
	Falha           bool
	Exec            bool
	Atendido        bool
	IDProcesso      string
	Notificarsempre bool
}

type ligacaoFilaPost060 struct {
	WhatsappligarerroID int `json:"whatsappligarerro_id"`
}

// dtUltimaAgora060: mesmo sentido do "now" no Xano para campo datetime (unix ms no JSON).
func dtUltimaAgora060() float64 {
	return float64(time.Now().UTC().UnixMilli())
}

// whatsappLigarErroQueryValues060: so envia na URL os campos presentes em merged; o Xano trata o resto como null.
func whatsappLigarErroQueryValues060(merged map[string]any) url.Values {
	order := []string{"falha", "exec", "tentativas", "atendido", "idProcesso", "limit"}
	v := url.Values{}
	for _, k := range order {
		val, ok := merged[k]
		if !ok || val == nil {
			continue
		}
		v.Set(k, queryParamString060(val))
	}
	return v
}

func queryParamString060(val any) string {
	switch x := val.(type) {
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func processLigacao060Routine(client *http.Client, cfg Config) error {
	base := strings.TrimRight(cfg.WhatsappLigarErroBase, "/")

	busy, err := queryWhatsappLigarErro060(client, base, map[string]any{
		"falha":      true,
		"tentativas": 0,
		"exec":       true,
	})
	if err != nil {
		return fmt.Errorf("ligacao060 query exec=true: %w", err)
	}
	busy = filter060(busy, true, 0, true)
	if len(busy) > 0 {
		log.Printf("ligacao060: candidato com exec=true em andamento, pulando ciclo")
		return nil
	}

	items, err := queryWhatsappLigarErro060(client, base, map[string]any{
		"falha":      true,
		"tentativas": 0,
		"exec":       false,
	})
	if err != nil {
		return fmt.Errorf("ligacao060 query lista: %w", err)
	}
	items = filter060(items, true, 0, false)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Tentativas != items[j].Tentativas {
			return items[i].Tentativas < items[j].Tentativas
		}
		return items[i].ID < items[j].ID
	})
	if len(items) == 0 {
		log.Printf("ligacao060: sem candidatos")
		return nil
	}

	// foreach 1: data = { exec: true } apenas
	for _, it := range items {
		if err := putWhatsappLigarErroPartial060(client, base, it.ID, map[string]any{"exec": true}); err != nil {
			return fmt.Errorf("ligacao060 put exec id=%d: %w", it.ID, err)
		}
	}

	for _, item := range items {
		// Releitura: se webhook ja marcou atendido (chamada util), nao recoloca na fila.
		verifiedSelf, errSelf := fetchWhatsappLigarErro060ByID(client, base, item.ID)
		if errSelf == nil && verifiedSelf.Atendido {
			log.Printf("ligacao060: id=%d ja atendido; resolvendo sem religar", item.ID)
			_ = putWhatsappLigarErroPartial060(client, base, item.ID, map[string]any{
				"falha":             false,
				"tentativas":        99,
				"dtUltimaTentativa": dtUltimaAgora060(),
				"exec":              false,
			})
			continue
		}

		// No Xano: se notificarsempre=true, $whatsappLigarErro3 fica vazio (sem query); is_empty -> nao bloqueia.
		temAtendido := false
		// Se notificarsempre=false: equivale a query single idProcesso == item && atendido == true (limit 1 no endpoint).
		if !item.Notificarsempre {
			if strings.TrimSpace(item.IDProcesso) != "" {
				proc := strings.TrimSpace(item.IDProcesso)
				outros, err := queryWhatsappLigarErro060(client, base, map[string]any{
					"atendido":   true,
					"idProcesso": proc,
					"limit":      1,
				})
				if err != nil {
					return fmt.Errorf("ligacao060 query atendido: %w", err)
				}
				for _, o := range outros {
					if o.ID == item.ID || o.ID == 0 {
						continue
					}
					// A lista GET pode devolver envelope onde id e atendido nao sao do mesmo objeto; confirma pelo GET /{id}.
					verified, err := fetchWhatsappLigarErro060ByID(client, base, o.ID)
					if err != nil {
						return fmt.Errorf("ligacao060 confirma outro id=%d: %w", o.ID, err)
					}
					if strings.TrimSpace(verified.IDProcesso) != proc {
						continue
					}
					// So considera bloqueio um registo *mais recente* (id maior) no mesmo idProcesso.
					// Linhas antigas (ex. 914) nao devem esvaziar $whatsappLigarErro3 nem resolver o erro novo (ex. 932).
					if verified.ID < item.ID {
						continue
					}
					if !verified.Atendido {
						continue
					}
					if verified.Falha {
						continue
					}
					if verified.Tentativas >= 99 {
						continue
					}
					temAtendido = true
					log.Printf("ligacao060: existe outro registro atendido mesmo idProcesso (outro_id=%d item_id=%d proc=%s)", verified.ID, item.ID, proc)
					break
				}
			}
		}

		// $whatsappLigarErro3 is_empty -> incrementa + fila; senao -> data falha/tentativas 99/dtUltima/exec false
		if !temAtendido {
			t0 := item.Tentativas
			var patch map[string]any
			if t0 == 0 {
				patch = map[string]any{"tentativas": t0 + 1}
			} else {
				patch = map[string]any{
					"tentativas":        t0 + 1,
					"dtUltimaTentativa": dtUltimaAgora060(),
				}
			}
			if err := putWhatsappLigarErroPartial060(client, base, item.ID, patch); err != nil {
				return fmt.Errorf("ligacao060 put tentativas id=%d: %w", item.ID, err)
			}
			if err := postLigacaoFila060(client, cfg.Ligacao060FilaPostURL, item.ID); err != nil {
				return fmt.Errorf("ligacao060 post ligacaofila id=%d: %w", item.ID, err)
			}
		} else {
			log.Printf("ligacao060: ramo 'resolver' (falha=false tentativas=99 ...) id=%d idProcesso=%q notificarsempre=%v",
				item.ID, strings.TrimSpace(item.IDProcesso), item.Notificarsempre)
			patch := map[string]any{
				"falha":             false,
				"tentativas":        99,
				"dtUltimaTentativa": dtUltimaAgora060(),
				"exec":              false,
			}
			if err := putWhatsappLigarErroPartial060(client, base, item.ID, patch); err != nil {
				return fmt.Errorf("ligacao060 put resolver id=%d: %w", item.ID, err)
			}
		}
	}
	return nil
}

// fetchWhatsappLigarErro060ByID: GET um registo; evita misturar campos do envelope na resposta da lista.
func fetchWhatsappLigarErro060ByID(client *http.Client, base string, id int) (whatsappLigarErroRow060, error) {
	if id <= 0 {
		return whatsappLigarErroRow060{}, fmt.Errorf("id invalido %d", id)
	}
	target := strings.TrimRight(base, "/") + "/" + strconv.Itoa(id)
	body, status, err := doJSONRequest(client, http.MethodGet, target, nil)
	if err != nil {
		return whatsappLigarErroRow060{}, err
	}
	if status < 200 || status >= 300 {
		return whatsappLigarErroRow060{}, fmt.Errorf("GET id=%d status=%d body=%s", id, status, compactBody(body))
	}
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return whatsappLigarErroRow060{}, err
	}
	obj := extractFirstObject(raw)
	if row, ok := row060FromAny(obj); ok {
		return row, nil
	}
	return whatsappLigarErroRow060{}, fmt.Errorf("GET id=%d: registo invalido", id)
}

func filter060(rows []whatsappLigarErroRow060, falha bool, tent int, exec bool) []whatsappLigarErroRow060 {
	out := make([]whatsappLigarErroRow060, 0, len(rows))
	for _, r := range rows {
		if r.Falha == falha && r.Tentativas == tent && r.Exec == exec {
			out = append(out, r)
		}
	}
	return out
}

// queryWhatsappLigarErro060: GET; so os campos do mapa vao na query — os outros o Xano assume como null.
func queryWhatsappLigarErro060(client *http.Client, base string, merged map[string]any) ([]whatsappLigarErroRow060, error) {
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return nil, err
	}
	u.RawQuery = whatsappLigarErroQueryValues060(merged).Encode()
	body, status, err := doJSONRequest(client, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("query GET status=%d body=%s", status, compactBody(body))
	}
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("whatsappligarerro json invalido: %w", err)
	}
	return extractWhatsapp060ListFromAny(raw), nil
}

func extractWhatsapp060ListFromAny(v any) []whatsappLigarErroRow060 {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []any:
		var out []whatsappLigarErroRow060
		for _, el := range x {
			if row, ok := row060FromAny(el); ok {
				out = append(out, row)
				continue
			}
			out = append(out, extractWhatsapp060ListFromAny(el)...)
		}
		return out
	case map[string]any:
		if row, ok := row060FromAny(x); ok {
			return []whatsappLigarErroRow060{row}
		}
		keys := []string{"dados", "data", "items", "Items", "records", "result", "results", "response", "payload", "list", "rows", "Dados", "Data"}
		var out []whatsappLigarErroRow060
		for _, k := range keys {
			if vv, ok := x[k]; ok {
				out = append(out, extractWhatsapp060ListFromAny(vv)...)
			}
		}
		if len(out) > 0 {
			return out
		}
		for _, vv := range x {
			out = append(out, extractWhatsapp060ListFromAny(vv)...)
		}
		return out
	default:
		return nil
	}
}

// atendido060FromMap: so considera atendido=true com valores explicitos.
// toBool(float64) usa !=0: qualquer numero !=0 virava true (errado para atendido).
// O GET pode devolver linha do mesmo idProcesso mesmo quando o filtro atendido=true falha no servidor.
func atendido060FromMap(m map[string]any) bool {
	for _, k := range []string{"atendido", "Atendido"} {
		if v, ok := m[k]; ok {
			return parseAtendidoEstrito060(v)
		}
	}
	return false
}

func parseAtendidoEstrito060(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		return s == "true" || s == "1" || s == "t" || s == "yes" || s == "sim"
	case float64:
		return x == 1
	case int:
		return x == 1
	case int64:
		return x == 1
	case json.Number:
		s := strings.TrimSpace(x.String())
		return s == "1" || strings.EqualFold(s, "true")
	default:
		return false
	}
}

func row060FromAny(v any) (whatsappLigarErroRow060, bool) {
	b, err := json.Marshal(v)
	if err != nil {
		return whatsappLigarErroRow060{}, false
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return whatsappLigarErroRow060{}, false
	}
	id := toInt(m["id"])
	if id == 0 {
		id = toInt(m["whatsappligarerro_id"])
	}
	if id == 0 {
		return whatsappLigarErroRow060{}, false
	}
	return whatsappLigarErroRow060{
		ID:              id,
		Tentativas:      toInt(m["tentativas"]),
		Falha:           toBool(m["falha"]),
		Exec:            toBool(m["exec"]),
		Atendido:        atendido060FromMap(m),
		IDProcesso:      stringFromAny(m["idProcesso"]),
		Notificarsempre: toBool(m["notificarsempre"]),
	}, true
}

func stringFromAny(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return strconv.FormatInt(int64(s), 10)
	default:
		return strings.TrimSpace(fmt.Sprint(s))
	}
}

// putWhatsappLigarErroPartial060 envia no PUT apenas o JSON do patch (equivalente ao data={...} do db.edit no Xano).
func putWhatsappLigarErroPartial060(client *http.Client, base string, id int, patch map[string]any) error {
	target := strings.TrimRight(base, "/") + "/" + strconv.Itoa(id)
	respBody, status, err := doJSONRequest(client, http.MethodPut, target, patch)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("put id=%d status=%d body=%s", id, status, compactBody(respBody))
	}
	return nil
}

func postLigacaoFila060(client *http.Client, rawURL string, whatsappligarerroID int) error {
	payload := ligacaoFilaPost060{WhatsappligarerroID: whatsappligarerroID}
	respBody, status, err := doJSONRequest(client, http.MethodPost, strings.TrimRight(rawURL, "/"), payload)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("post ligacaofila status=%d body=%s", status, compactBody(respBody))
	}
	return nil
}
