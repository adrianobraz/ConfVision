# ConfVision — Disaster Recovery

Complementa [HIGH_AVAILABILITY.md](./HIGH_AVAILABILITY.md). RPO/RTO negócio: **A DEFINIR** (Fase 5).

---

## Cenário: perda de um Node (Rust + MTX local)

| Ativo | Recuperável? | Como |
|-------|--------------|------|
| Processamento analítico | Sim (rebuild) | Novo node, `worker_id` novo ou mesmo após wipe, reassign câmeras PG |
| Estado in-memory Rust | Não | Resync API |
| Fila Redis pendente | Parcial | Jobs perdidos se disco Redis perdido |
| Mídia temp local | Não crítico | CAPTURE_DIR |
| DVR disco local não uploaded | Risco | Segmentos em `/recordings` — upload retry |
| Metadata PG | Sim | Desde PG intacto |

**RTO típico manual:** tempo deploy + sync interval × câmeras ( **não medido** ).

---

## Cenário: perda Control Plane (Go) + PG

| Ativo | Efeito |
|-------|--------|
| Data plane RTSP | Continua temporariamente (ADR-012) |
| Novos eventos persistidos | **Não** |
| Assign / portal | **Down** |

Recuperação: restore PG backup + redeploy API. **Restore test pendente.**

---

## Cenário: perda cluster / região inteira

Recuperável desde:

- Backup PostgreSQL
- Credenciais S3 (mídia durável)
- Rebuild nodes MediaMTX + Rust registry `RUST_PROCESSOR_BASE_URLS`

**Multi-region:** não implementado — DR = procedimento manual cross-host.

---

## Backup mínimo checklist

- [ ] PostgreSQL backup automático verificado
- [ ] Restore test em staging
- [ ] Export env/registry processors (sem secrets no git)
- [ ] S3 lifecycle / versioning **A DEFINIR**

---

## Ordem de restore (sugerida)

1. PostgreSQL available + migration
2. Go API healthy
3. Redis
4. MediaMTX nodes
5. Rust processors (canary first)
6. Python workers DVR/motion
7. Validar ping + 1 câmera teste + d3-online-test quick
