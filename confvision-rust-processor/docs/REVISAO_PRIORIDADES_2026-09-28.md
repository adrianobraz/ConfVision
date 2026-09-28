# Revisão de prioridades — 2026-09-28

Comparação do checklist “subir para clientes / Fase D” com **testes ao vivo** e decisões de produto (Rust-only, worker Python off).

**Como rever:** `bash confvision-rust-processor/scripts/phase-d6-verify.sh` + este doc.

---

## Legenda

| Símbolo | Significado |
|---------|-------------|
| ✅ | OK nos testes / evidência recente |
| 🟡 | Parcial ou depende de câmera/cliente |
| ❌ | Pendente ou falha conhecida |
| ➖ | N/A (obsoleto pela decisão Rust-only) |

---

## Prioridades altas

| # | Item | Status | Evidência / nota |
|---|------|--------|------------------|
| 1 | Foxpro: `confvision` + sidecar + Rust A/B **running** | ✅ | D6 **OK**; `/health` A+B `status=ok` (2026-09-28) |
| 2 | Proxmox: Go `confmonit4confvision` | ✅ | `/vis_health` → postgres ok |
| 3 | Go **D5** deployado | ✅ | `/vis_rust_processor_capacity` → **HTTP 200** |
| 4 | D6 cron CT111 | ✅ | `OK D6` + `OK Go /vis_rust_processor_capacity` (log operador) |
| 5 | Postgres / **D5 assign** | ✅ | Export 2026-09-27: **18 ativas**, todas com `worker_id` Rust A/B — ver [CAMERAS_AUDIT_2026-09-28.md](./CAMERAS_AUDIT_2026-09-28.md) |
| 5b | **Sync Go → Rust** | 🟡 | Maioria com **`analitico_pausado=true`** (excluída do sync); API retornou **0 cameras** — despausar + conferir SQL |
| 6 | **RTSP** sem worker | 🟡 | **0** `rtsp_url_sec` no CSV; **4** com rtsp_404; cam **4** depende MediaMTX — ver audit |
| 7 | **D3 aceite** E2E `vis_evento` | ❌ | `events_published=0` em A e B (health 2026-09-28) |

---

## Prioridades médias

| # | Item | Status | Nota |
|---|------|--------|------|
| 8 | Câmera 4 / RTMP 404 | 🟡 | Só relevante se câmera 4 estiver ativa no Rust; com 0 câmeras online, **não é blocker atual** |
| 9 | RAM / `capacity_state` | ✅ | A e B **healthy** com 0 câmeras; revisar de novo com câmeras online |
| 10 | YOLO timeout vs CPU sidecar | 🟡 | Sidecar responde (D6); inferência cold longa — monitorar sob carga real |
| 11 | Prometheus / alertas (C2 / D6.2) | 🟡 | Cron D6 ok; scrape Prometheus **opcional** |

---

## Backlog / N/A

| Item | Status |
|------|--------|
| `confvision-worker` Start | ➖ **Stop permanente** |
| Rollback Python (B3) | ➖ |
| Piloto 1 câmera + resto Python (A1, A7) | ➖ |
| D5.3 rebalance auto | ❌ backlog |
| D4 GPU | ➖ adiado |
| Tag Git / handoff formal (B6) | 🟡 opcional — `FASE_D_FECHAMENTO.md` existe |
| Baseline CPU anotado (A6) | 🟡 opcional |

---

## Fase D — veredito

| Bloco | Fechado? |
|-------|----------|
| D1, D2, D5, D6 (monitor) | **Sim** (operacional) |
| D3 aceite produto | **Não** — 1º `vis_evento` E2E |
| D4 | Fora de escopo |

**Conclusão:** A **infra de controle e monitor está OK**. O que falta para “clientes sentirem analítico” é **câmeras atribuídas + RTSP + primeiro evento** (linhas 5–7), não redeploy de D5/D6.

---

## Próximos passos mínimos (só o que ainda move agulha)

1. SQL: câmeras de teste dos clientes → `ativo`, `deteccao_humano`, `worker_id` Rust (ou D5 assign).  
2. Confirmar **RTSP** por câmera (`rtsp_url_sec` ou MediaMTX).  
3. Esperar movimento → `events_published` > 0 → conferir `vis_evento`.  
4. Manter cron D6; `git pull` no CT111 se fix jq estiver no remoto.

---

## Teste rápido (copiar)

```bash
curl -sS https://vision.confmonit2.com.br/vis_health
curl -sS -H "Authorization: Bearer $VIS_WORKER_API_KEY" \
  https://vision.confmonit2.com.br/vis_rust_processor_capacity | head -c 400
bash confvision-rust-processor/scripts/phase-d6-verify.sh
```
