package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type BillingWorker struct {
	xano *XanoClient
	cfg  Config
}

type pendenteAssinatura struct {
	Assinatura struct {
		ID              int    `json:"id"`
		IDFranqueado    string `json:"id_franqueado"`
		Produto         string `json:"produto"`
		Plano           string `json:"plano"`
		ProximaCobranca string `json:"proxima_cobranca_em"`
	} `json:"assinatura"`
	CicloRef string `json:"ciclo_ref"`
}

type listarPendentesResp struct {
	Dados []pendenteAssinatura `json:"dados"`
	Total int                  `json:"total"`
}

type confvisionAgrupadoResp struct {
	Agrupado map[string][]struct {
		Licenca struct {
			ID           int    `json:"id"`
			IDFranqueado string `json:"id_franqueado"`
			Plano        string `json:"plano"`
			ValidoAte    string `json:"valido_ate"`
		} `json:"licenca"`
		CicloRef string `json:"ciclo_ref"`
	} `json:"agrupado"`
}

func (w *BillingWorker) RunWeekly() error {
	log.Println("=== inicio ciclo billing ===")

	if err := w.suspenderInadimplentes(); err != nil {
		log.Printf("avisar: suspender inadimplentes: %v", err)
	}

	if err := w.expirarConfVision(); err != nil {
		log.Printf("avisar: expirar confvision: %v", err)
	}

	if err := w.gerarFaturasAssinaturas(); err != nil {
		return err
	}

	if err := w.gerarFaturasConfVision(); err != nil {
		return err
	}

	log.Println("=== fim ciclo billing ===")
	return nil
}

func (w *BillingWorker) workerPayload(extra map[string]any) map[string]any {
	base := map[string]any{
		"worker_key": w.cfg.WorkerKey,
	}
	for k, v := range extra {
		base[k] = v
	}
	return base
}

func (w *BillingWorker) suspenderInadimplentes() error {
	if w.cfg.DryRun {
		log.Println("[dry-run] fp_fatura_suspender_vencidas")
		return nil
	}
	raw, err := w.xano.Post("/fp_fatura_suspender_vencidas", w.workerPayload(nil))
	if err != nil {
		return err
	}
	log.Printf("assinaturas suspensas por fatura vencida: %s", string(raw))
	return nil
}

func (w *BillingWorker) expirarConfVision() error {
	if w.cfg.DryRun {
		log.Println("[dry-run] fp_vis_licenca_expirar_vencidas")
		return nil
	}
	raw, err := w.xano.Post("/fp_vis_licenca_expirar_vencidas", w.workerPayload(nil))
	if err != nil {
		return err
	}
	log.Printf("confvision expiradas: %s", string(raw))
	return nil
}

func (w *BillingWorker) gerarFaturasAssinaturas() error {
	payload := w.workerPayload(map[string]any{
		"dias_antecedencia": w.cfg.DiasAntecedencia,
	})
	raw, err := w.xano.Post("/fp_assinatura_listar_pendentes_fatura", payload)
	if err != nil {
		return fmt.Errorf("listar pendentes assinatura: %w", err)
	}

	var resp listarPendentesResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("parse pendentes: %w", err)
	}

	log.Printf("assinaturas pendentes fatura: %d", resp.Total)

	for _, p := range resp.Dados {
		if w.cfg.DryRun {
			log.Printf("[dry-run] gerar fatura assinatura_id=%d ciclo=%s", p.Assinatura.ID, p.CicloRef)
			continue
		}
		genPayload := w.workerPayload(map[string]any{
			"assinatura_id": p.Assinatura.ID,
			"ciclo_ref":     p.CicloRef,
		})
		out, err := w.xano.Post("/fp_fatura_gerar_automatica", genPayload)
		if err != nil {
			log.Printf("erro gerar fatura assinatura %d: %v", p.Assinatura.ID, err)
			continue
		}
		log.Printf("fatura gerada assinatura %d: %s", p.Assinatura.ID, string(out))
	}
	return nil
}

func (w *BillingWorker) gerarFaturasConfVision() error {
	payload := w.workerPayload(map[string]any{
		"dias_antecedencia": w.cfg.DiasAntecedencia,
	})
	raw, err := w.xano.Post("/fp_vis_licenca_listar_pendentes_renovacao", payload)
	if err != nil {
		return fmt.Errorf("listar renovacao confvision: %w", err)
	}

	var resp confvisionAgrupadoResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("parse confvision: %w", err)
	}

	for idFranq, itens := range resp.Agrupado {
		if len(itens) == 0 {
			continue
		}
		ids := make([]int, 0, len(itens))
		cicloPrefix := ""
		for _, it := range itens {
			ids = append(ids, it.Licenca.ID)
			if cicloPrefix == "" {
				cicloPrefix = it.CicloRef
			}
		}
		if w.cfg.DryRun {
			log.Printf("[dry-run] fatura confvision franqueado=%s licencas=%v", idFranq, ids)
			continue
		}
		genPayload := w.workerPayload(map[string]any{
			"id_franqueado":    idFranq,
			"vis_licenca_ids":  ids,
			"ciclo_ref_prefix": cicloPrefix,
		})
		out, err := w.xano.Post("/fp_fatura_gerar_automatica_confvision", genPayload)
		if err != nil {
			log.Printf("erro fatura confvision %s: %v", idFranq, err)
			continue
		}
		log.Printf("fatura confvision %s: %s", idFranq, string(out))
	}
	return nil
}
