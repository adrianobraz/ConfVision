package collect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

type mediaPath struct {
	Ready    bool   `json:"ready"`
	ConfName string `json:"confName"`
}

func ReadMediaMTX(ctx context.Context, base, user, pass string) []model.Reading {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return []model.Reading{model.Unavailable("mediamtx", "MEDIAMTX_API_BASE nao configurado")}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v3/paths/list", nil)
	if err != nil {
		return []model.Reading{model.Unavailable("mediamtx", "url invalida")}
	}
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return []model.Reading{model.Unavailable("mediamtx", "api sem resposta")}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return []model.Reading{model.Unavailable("mediamtx", "api recusou a leitura")}
	}
	var payload struct {
		ItemCount int         `json:"itemCount"`
		Items     []mediaPath `json:"items"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return []model.Reading{model.Unavailable("mediamtx", "resposta ilegivel")}
	}
	ready := 0
	for _, item := range payload.Items {
		if item.Ready {
			ready++
		}
	}
	return []model.Reading{
		model.Text("mediamtx", model.OrigemMedido, "disponivel"),
		model.Measured("paths", "count", float64(len(payload.Items))),
		model.Measured("paths_prontos", "count", float64(ready)),
		model.Measured("item_count", "count", float64(payload.ItemCount)),
	}
}
