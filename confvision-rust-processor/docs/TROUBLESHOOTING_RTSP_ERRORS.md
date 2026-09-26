# Troubleshooting — erros RTSP / decode (ConfVision Rust)

Guia para **suporte**, **ops** e **engenharia**: linguagem simples, detalhe técnico e o que o sistema faz **sozinho** (sem SSH/SQL manual).

Relacionado: [STREAM_RETRY_POLICY.md](./STREAM_RETRY_POLICY.md), [PAINEL_ATIVO_VS_STREAM_POLICY.md](./PAINEL_ATIVO_VS_STREAM_POLICY.md), [RUNBOOK_CAMERAS_404_ATIVO.md](./RUNBOOK_CAMERAS_404_ATIVO.md).

## Como ler a tabela

| Coluna | Significado |
|--------|-------------|
| **Simples** | O que o cliente vê / mensagem em linguagem de negócio |
| **Técnico** | Texto típico no log do processor ou causa provável |
| **`error_class`** | Código enviado no ping (`camera_stream_health`) quando classificado |
| **Política** | Classe usada pelo retry automático (`PathAbsent`, `Auth`, `Transient`) |
| **Ação automática** | Comportamento Rust + API Go sem intervenção manual |

## Erros comuns

| Simples | Técnico | `error_class` | Política | Ação automática |
|---------|---------|---------------|----------|-----------------|
| “Câmera sem vídeo no servidor” / path errado | RTSP **404**, `not found`, `unexpected rtsp response status: 404` | `rtsp_404` | PathAbsent | Backoff crescente (1 min → 60 min); contadores em `stream_falhas_consecutivas`; após limiar horário → `analitico_pausado` + `stream_motivo_pausa=sistema_stream_*`; evento `stream_failure` no ping |
| “Senha ou usuário RTSP incorreto” | **401/403**, `unauthorized` | `rtsp_auth` | Auth | Espera **30 min** entre tentativas; **não** pausa analítico só por auth (corrigir credencial no dispositivo/publisher) |
| “Vídeo chega quebrado / trava e reconecta” | **FU-A H.264**: `FU-A has start bit unset`, fragmentation unit RTP | `rtp_h264_fu_a` | Transient | Reconnect com backoff local; evento **`stream_incident`** (esboço) grava `stream_ultimo_erro` + `stream_erro_classe` no Postgres; **não** conta como 404 |
| “Rede instável ou MediaMTX lento” | `timeout`, `timed out`, `connection refused`, `broken pipe` | `network_timeout` | Transient | Idem transient + `stream_incident` |
| “Falha ao abrir stream (genérico)” | Outros erros de demux/decode/OpenCV | `unknown_transient` | Transient | Reconnect; incidente opcional se houver texto de erro |
| “Vídeo voltou” | Sessão RTSP OK de novo | — | — | `stream_ok` no ping → zera contadores, limpa motivo `sistema_stream%`, atualiza `ultimo_stream_ok_em` |
| “Nunca conectou após dias” | Sem `ultimo_stream_ok_em` + grace expirado | — | PathAbsent | `pause_analytic` → `sistema_stream_sem_historico` |
| “Muitas falhas seguidas (404)” | ≥ limiar horário (env `STREAM_HOURLY_*`) | `rtsp_404` | PathAbsent | `pause_analytic` → `sistema_stream_6h_horarias` |

## Onde ver no sistema

| Onde | Campo / sinal |
|------|----------------|
| **Processor (métricas / estado)** | `last_error`, `reconnect_count`, `rtsp_errors`, `status=offline` |
| **Postgres `vis_camera`** | `stream_falhas_consecutivas`, `stream_tentativas_horarias`, `stream_motivo_pausa`, `analitico_pausado`, `stream_policy_generation`, `stream_ultimo_erro`, `stream_erro_classe`, `stream_ultimo_erro_em` |
| **Ping worker** | `POST /vis_worker_ping` → `camera_stream_health[]`: `stream_ok`, `stream_failure`, `stream_incident`, `pause_analytic` |

## Diagnóstico rápido (ops)

1. **`ffprobe`** no **mesmo host** que o processor (VPS / container rust-pilot), URL interna `rtsp://foxpro_confvision:8554/cam/...`.
   - OK no ffprobe + offline no Rust → olhar **`error_class`** (`rtp_h264_fu_a` vs rede).
   - 404 no ffprobe → publisher/MediaMTX/path; política PathAbsent já reduz CPU.
2. **`rtsp_404_count=0`** nas métricas do processor → não misturar com FU-A (transient).
3. **CPU alta + muitas cams offline transient** → revisar carga (Fase C: `MAX_CAMERAS`, admission); FU-A pode ser sintoma de stream corrupto ou overload.

## Reset comercial (usuário)

Desligar/ligar **Ativo** no painel **não** reinicia container: incrementa **`stream_policy_generation`** e zera contadores — ver [PAINEL_ATIVO_VS_STREAM_POLICY.md](./PAINEL_ATIVO_VS_STREAM_POLICY.md).

Reativação só de analítico pausado pelo sistema: `POST /vis_camera_stream_reactivate?camera_id=`.

## Migration

Colunas de diagnóstico: `sql/migrations/20260326_vis_camera_stream_error_diag.sql` (após `20260326_vis_camera_stream_policy.sql`).

## Roadmap (automação)

- Relatório diário por `error_class` / franqueado.
- Ajuste adaptativo de decode/motion por câmera quando `rtp_h264_fu_a` persistir.
- Auto-assign `worker_id` e realocação entre processors (Fase C).
