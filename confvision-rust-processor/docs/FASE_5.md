# Fase 5 — Multi-câmera (capacidade medida)

**Objetivo:** descobrir quantas câmeras **este** servidor suporta (CPU/GPU/RAM/rede), subindo carga de forma **progressiva** — não chutar 50/200/600.

**Pré-requisitos**

1. Fases **4.1 / 4.1A / D** OK (processors A+B, workers alinhados).
2. Câmeras **no sync**: `ativo`, `deteccao_humano`, `analitico_pausado=false`, `worker_id` A ou B.
3. **RTSP Opção A** (`rtsp_url_sec`) ou stream válido — senão `rtsp_404` distorce a medição.
4. Sidecar YOLO no ar se medir pipeline completo (D3).

---

## Degraus sugeridos

```text
1 → 5 → 10 → 20 → 40 → 50 → 100 → …
```

Pare ou desacelere se por **≥5 min**:

- `capacity_state=critical` estável
- `load_advisory=reject_admission`
- `rtsp_404_count` subindo
- `frames_dropped` crescendo rápido
- reconnect storm nos logs

---

## Ferramentas (repo)

| Script | Uso |
|--------|-----|
| `phase-f5-verify.sh` | CT111: `/health`, `/capacity-report`, `/metrics` OK |
| `phase-f5-snapshot.sh` | Uma linha CSV por processor |
| `phase-f5-ramp-run.sh` | Rampa interativa + arquivo `ramp-f5-*.csv` |
| `c1-ab-baseline.sh` | Amostras longas em um URL (legado C1) |
| `capacity-report.sh` | Relatório legível de um processor |

### CT111 / monitor (sem cargo)

```bash
bash confvision-rust-processor/scripts/phase-f5-verify.sh
```

### Rampa (operador com cadastro)

```bash
export F5_STAGES="1 5 10 20"
export F5_HOLD_SEC=300
export F5_SAMPLE_EVERY=60
bash confvision-rust-processor/scripts/phase-f5-ramp-run.sh
```

### Snapshot manual

```bash
F5_HEADER=1 bash confvision-rust-processor/scripts/phase-f5-snapshot.sh > meu.csv
F5_STAGE=5 F5_NOTE=pico bash confvision-rust-processor/scripts/phase-f5-snapshot.sh >> meu.csv
```

---

## O que registrar (por degrau)

- `cameras_online` / `cameras_total` (A e B)
- `capacity_state`, `limiting_resource`, `estimated_*_cameras`
- CPU, RAM, GPU, VRAM (`capacity-report`)
- `fps_total`, `frames_received`, `frames_dropped`
- `rtsp_404_count`, `decode_backend_effective`
- Estabilidade (reconnect, temperatura GPU se houver)

Template: [FASE_5_RESULTADO.template.md](./FASE_5_RESULTADO.template.md)

---

## Estado atual (foxpro)

Com **sync 0** (analítico pausado / sem elegíveis), a rampa **não mede decode real** — só infra. Isso é esperado até cadastro Opção A.

---

## Próxima fase

**Fase 6** — capacidade dinâmica centralizada (real / usada / disponível por worker).

Ver: [FASE_4_1A_FECHAMENTO.md](./FASE_4_1A_FECHAMENTO.md), [OPCAO_A_RTSP_DIRETO.md](./OPCAO_A_RTSP_DIRETO.md).
