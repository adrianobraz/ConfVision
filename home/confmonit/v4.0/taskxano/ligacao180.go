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

// Rotina ligacao180: migracao da task Xano ligacao_3 (freq 180s na origem).
// Mutex proprio (runningLigacao180), independente de ligacao060.

func runLigacao180Loop(client *http.Client, cfg Config, shutdown <-chan struct{}) {
	ticker := time.NewTicker(cfg.Ligacao180Interval)
	defer ticker.Stop()
	runLigacao180Once(client, cfg)
	for {
		select {
		case <-shutdown:
			log.Printf("ligacao180: loop encerrado")
			return
		case <-ticker.C:
			runLigacao180Once(client, cfg)
		}
	}
}

func runLigacao180Once(client *http.Client, cfg Config) {
	if !atomic.CompareAndSwapInt32(&runningLigacao180, 0, 1) {
		log.Printf("ligacao180: execucao anterior em andamento, pulando ciclo")
		return
	}
	defer atomic.StoreInt32(&runningLigacao180, 0)

	start := time.Now()
	attempts := cfg.Retries + 1
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			backoff := backoffForAttempt(attempt)
			log.Printf("ligacao180 retry=%d backoff=%s", attempt-1, backoff)
			time.Sleep(backoff)
		}
		err := processLigacao180Routine(client, cfg)
		if err == nil {
			log.Printf("ligacao180 sucesso | attempt=%d/%d | duracao=%s", attempt, attempts, time.Since(start))
			return
		}
		lastErr = err
		log.Printf("ligacao180 falha | attempt=%d/%d | erro=%v", attempt, attempts, err)
	}
	log.Printf("ligacao180 erro final | duracao=%s | erro=%v", time.Since(start), lastErr)
}

type whatsappLigarErroRow180 struct {
	ID                int
	Tentativas        int
	Falha             bool
	Exec              bool
	Atendido          bool
	IDProcesso        string
	Notificarsempre   bool
	DtUltimaTentativa float64 // unix ms no JSON Xano; 0 = ausente
}

func whatsappLigarErroQueryValues180(merged map[string]any) url.Values {
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

func queryWhatsappLigarErro180(client *http.Client, base string, merged map[string]any) ([]whatsappLigarErroRow180, error) {
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return nil, err
	}
	u.RawQuery = whatsappLigarErroQueryValues180(merged).Encode()
	body, status, err := doJSONRequest(client, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("ligacao180 query GET status=%d body=%s", status, compactBody(body))
	}
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("ligacao180 whatsappligarerro json invalido: %w", err)
	}
	return extractWhatsapp180ListFromAny(raw), nil
}

func extractWhatsapp180ListFromAny(v any) []whatsappLigarErroRow180 {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []any:
		var out []whatsappLigarErroRow180
		for _, el := range x {
			if row, ok := row180FromAny(el); ok {
				out = append(out, row)
				continue
			}
			out = append(out, extractWhatsapp180ListFromAny(el)...)
		}
		return out
	case map[string]any:
		if row, ok := row180FromAny(x); ok {
			return []whatsappLigarErroRow180{row}
		}
		keys := []string{"dados", "data", "items", "Items", "records", "result", "results", "response", "payload", "list", "rows", "Dados", "Data"}
		var out []whatsappLigarErroRow180
		for _, k := range keys {
			if vv, ok := x[k]; ok {
				out = append(out, extractWhatsapp180ListFromAny(vv)...)
			}
		}
		if len(out) > 0 {
			return out
		}
		for _, vv := range x {
			out = append(out, extractWhatsapp180ListFromAny(vv)...)
		}
		return out
	default:
		return nil
	}
}

func row180FromAny(v any) (whatsappLigarErroRow180, bool) {
	b, err := json.Marshal(v)
	if err != nil {
		return whatsappLigarErroRow180{}, false
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return whatsappLigarErroRow180{}, false
	}
	id := toInt(m["id"])
	if id == 0 {
		id = toInt(m["whatsappligarerro_id"])
	}
	if id == 0 {
		return whatsappLigarErroRow180{}, false
	}
	return whatsappLigarErroRow180{
		ID:                id,
		Tentativas:        toInt(m["tentativas"]),
		Falha:             toBool(m["falha"]),
		Exec:              toBool(m["exec"]),
		Atendido:          atendido060FromMap(m),
		IDProcesso:        stringFromAny(m["idProcesso"]),
		Notificarsempre:   toBool(m["notificarsempre"]),
		DtUltimaTentativa: floatFromAnyDt180(m["dtUltimaTentativa"]),
	}, true
}

func floatFromAnyDt180(v any) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		s := strings.TrimSpace(fmt.Sprint(x))
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	}
}

// filter180DtUltimaPlus3Min: (dtUltimaTentativa + 3min) <= now (stack Xano timestamp_add_minutes:3).
func filter180DtUltimaPlus3Min(rows []whatsappLigarErroRow180) []whatsappLigarErroRow180 {
	now := time.Now().UTC()
	out := make([]whatsappLigarErroRow180, 0, len(rows))
	for _, r := range rows {
		if r.DtUltimaTentativa <= 0 {
			continue
		}
		t := time.UnixMilli(int64(r.DtUltimaTentativa)).UTC().Add(3 * time.Minute)
		if !t.After(now) {
			out = append(out, r)
		}
	}
	return out
}

func filter180(rows []whatsappLigarErroRow180, falha bool, tent int, exec bool) []whatsappLigarErroRow180 {
	out := make([]whatsappLigarErroRow180, 0, len(rows))
	for _, r := range rows {
		if r.Falha == falha && r.Tentativas == tent && r.Exec == exec {
			out = append(out, r)
		}
	}
	return out
}

func processLigacao180Routine(client *http.Client, cfg Config) error {
	base := strings.TrimRight(cfg.WhatsappLigarErroBase, "/")

	// Single: falha true, tentativas 1, exec true — se nao vazio, sai (equivale ao return "" do Xano).
	busy, err := queryWhatsappLigarErro180(client, base, map[string]any{
		"falha":      true,
		"tentativas": 1,
		"exec":       true,
	})
	if err != nil {
		return fmt.Errorf("ligacao180 query exec=true tentativas=1: %w", err)
	}
	busy = filter180(busy, true, 1, true)
	if len(busy) > 0 {
		log.Printf("ligacao180: candidato com exec=true (tentativas=1) em andamento, pulando ciclo")
		return nil
	}

	items, err := queryWhatsappLigarErro180(client, base, map[string]any{
		"falha":      true,
		"tentativas": 1,
		"exec":       false,
	})
	if err != nil {
		return fmt.Errorf("ligacao180 query lista: %w", err)
	}
	items = filter180(items, true, 1, false)
	items = filter180DtUltimaPlus3Min(items)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Tentativas != items[j].Tentativas {
			return items[i].Tentativas < items[j].Tentativas
		}
		return items[i].ID < items[j].ID
	})
	if len(items) == 0 {
		log.Printf("ligacao180: sem candidatos")
		return nil
	}

	for _, it := range items {
		if err := putWhatsappLigarErroPartial060(client, base, it.ID, map[string]any{"exec": true}); err != nil {
			return fmt.Errorf("ligacao180 put exec id=%d: %w", it.ID, err)
		}
	}

	for _, item := range items {
		verifiedSelf, errSelf := fetchWhatsappLigarErro060ByID(client, base, item.ID)
		if errSelf == nil && verifiedSelf.Atendido {
			log.Printf("ligacao180: id=%d ja atendido; resolvendo sem religar", item.ID)
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
		// Se notificarsempre=false: query atendido=true + idProcesso (limit 1); confirma com GET /{id} como ligacao060.
		if !item.Notificarsempre {
			if strings.TrimSpace(item.IDProcesso) != "" {
				proc := strings.TrimSpace(item.IDProcesso)
				outros, err := queryWhatsappLigarErro180(client, base, map[string]any{
					"atendido":   true,
					"idProcesso": proc,
					"limit":      1,
				})
				if err != nil {
					return fmt.Errorf("ligacao180 query atendido: %w", err)
				}
				for _, o := range outros {
					if o.ID == item.ID || o.ID == 0 {
						continue
					}
					// Lista GET pode misturar envelope; confirma pelo GET /{id} (atendido estrito em row060FromAny).
					verified, err := fetchWhatsappLigarErro060ByID(client, base, o.ID)
					if err != nil {
						return fmt.Errorf("ligacao180 confirma outro id=%d: %w", o.ID, err)
					}
					if strings.TrimSpace(verified.IDProcesso) != proc {
						continue
					}
					// So bloqueia se existir outro registo *mais recente* (id maior) no mesmo idProcesso (igual ligacao060).
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
					log.Printf("ligacao180: existe outro registro atendido mesmo idProcesso (outro_id=%d item_id=%d proc=%s)", verified.ID, item.ID, proc)
					break
				}
			}
		}

		if !temAtendido {
			t0 := item.Tentativas
			var patch map[string]any
			// Stack Xano: if tentativas==0 so tentativas+1; else +1 e dtUltima. Lista filtra tentativas==1 -> ramo else.
			if t0 == 0 {
				patch = map[string]any{"tentativas": t0 + 1}
			} else {
				patch = map[string]any{
					"tentativas":        t0 + 1,
					"dtUltimaTentativa": dtUltimaAgora060(),
				}
			}
			if err := putWhatsappLigarErroPartial060(client, base, item.ID, patch); err != nil {
				return fmt.Errorf("ligacao180 put tentativas id=%d: %w", item.ID, err)
			}
			if err := postLigacaoFila060(client, cfg.Ligacao180FilaPostURL, item.ID); err != nil {
				return fmt.Errorf("ligacao180 post ligacaofila id=%d: %w", item.ID, err)
			}
		} else {
			log.Printf("ligacao180: ramo 'resolver' (falha=false tentativas=99 ...) id=%d idProcesso=%q notificarsempre=%v",
				item.ID, strings.TrimSpace(item.IDProcesso), item.Notificarsempre)
			patch := map[string]any{
				"falha":             false,
				"tentativas":        99,
				"dtUltimaTentativa": dtUltimaAgora060(),
				"exec":              false,
			}
			if err := putWhatsappLigarErroPartial060(client, base, item.ID, patch); err != nil {
				return fmt.Errorf("ligacao180 put resolver id=%d: %w", item.ID, err)
			}
		}
	}
	return nil
}
