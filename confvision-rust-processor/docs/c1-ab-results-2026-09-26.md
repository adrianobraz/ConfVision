# C1 — Resultado A/B Python vs Rust (2026-09-26)

| Campo | Valor |
|-------|--------|
| Data | 2026-09-26 |
| Câmera `id` | _pendente (sugerido: 5)_ |
| Janela parada | _pendente_ min |
| Janela movimento | _pendente_ min |

## Python (`worker_id` = _preencher_)

- CPU container/host (EasyPanel): % médio / pico — **pendente**
- RAM: MB — **pendente**
- Notas:

## Rust (`rust-processor-pilot-01`)

Fleet piloto: **8 câmeras** (5, 18, 19, 20, 21, 22, 26, 27) — amostra **fleet**, não câmera única.

Amostras rápidas (intervalo 15 s, 5 pontos, UTC):

```text
# timestamp_iso,fps_total,cameras_online,capacity_state,load_advisory,frames_received,cpu_percent
2026-09-26T18:21:55Z,73.09,8,healthy,normal,51471,50.87
2026-09-26T18:22:10Z,58.18,8,healthy,normal,52384,50.91
2026-09-26T18:22:25Z,59.89,8,healthy,caution,53300,50.92
2026-09-26T18:22:40Z,58.37,8,healthy,normal,54260,50.60
2026-09-26T18:22:56Z,67.13,8,healthy,normal,55228,50.43
```

- `capacity_state` médio: **healthy**
- `fps_total` médio: ~**63**
- CPU container: ~**51%**

Para comparativo justo 1 câmera: repetir `c1-ab-baseline.sh` após mover **só id 5** para Rust/Python alternado.

## Decisão

- [ ] Rust ≤ CPU Python com mesma utilidade motion
- [ ] Seguir escala Rust / aguardar dedicado GPU
- Observações: C3 admission ativo; C4 adiado enquanto healthy.
