# Fase 4.1 — GPU / NVDEC (fechamento)

**Objetivo:** Rust preparado para decode **CPU** (foxpro VPS) e **NVIDIA NVDEC** quando o host tiver GPU + FFmpeg com `h264_cuvid`.

**Relação:** Fase D (foxpro CPU, 2 processors) **encerrada**. Fase 4.1 é **código + builds + política de aceleração**, sem exigir câmeras no sync.

---

## Entregáveis (checklist)

| Item | Status | Notas |
|------|--------|--------|
| `NvdecH264Decoder` + `HwDeviceCtxGuard` | ✅ | `Drop` libera `AVBufferRef`; guard em falha de `try_new` |
| `VIDEO_ACCELERATION` auto / cpu / gpu | ✅ | `src/decode/acceleration/` |
| `gpu` sem NVDEC no boot | ✅ | Erro explícito (`gpu_unavailable_error`), processo não sobe em modo gpu inválido |
| `auto` sem GPU | ✅ | Efetivo **CPU**, backend planejado CPU |
| `auto` com stack NVDEC | ✅ | Planeja **nvdec** (features `ffmpeg-decode` + `ffmpeg-nvdec`) |
| Fallback runtime HW→CPU | ✅ | Só **`auto`** + `DECODE_RUNTIME_FALLBACK=1`; **`gpu` proíbe** (`runtime_fallback_allowed`, `policy_forbids_cpu_fallback`) |
| Métricas / health | ✅ | `decode_backend_effective`, `gpu_decode_state`, `hw_fallback_to_cpu_count` |
| `cargo fmt` | ✅ | Gate no script `phase-4.1-verify.sh` |
| `cargo test` (default, sem FFmpeg) | ✅ | 124+ testes unitários |
| Build release **CPU** (sem features) | ✅ | EasyPanel foxpro atual |
| Build **`ffmpeg-decode`** | ✅ | Dockerfile Bookworm; validar em Linux/CI |
| Build **`ffmpeg-nvdec`** | ✅ | Mesmo Dockerfile + features; validar em host NVIDIA |

---

## Variáveis de ambiente

```env
# auto | cpu | gpu
VIDEO_ACCELERATION=auto
# auto | cuda | vaapi | qsv  (NVDEC usa cuda/auto)
VIDEO_GPU_BACKEND=auto
# 1 = auto pode fazer fallback runtime NVDEC→CPU após threshold de erros
DECODE_RUNTIME_FALLBACK=1
DECODE_HW_ERROR_THRESHOLD=10
```

| Modo | Boot | Runtime (erros NVDEC) |
|------|------|------------------------|
| **cpu** | Sempre FFmpeg CPU | N/A |
| **gpu** | Falha se sem NVDEC | Sem fallback para CPU |
| **auto** | NVDEC se stack OK, senão CPU | Fallback opcional (`DECODE_RUNTIME_FALLBACK`) |

Logs de boot: `video acceleration initialized (Fase 4.1)`.

---

## Builds

### VPS / EasyPanel (sem GPU)

```bash
cargo build --release
# ou imagem Docker padrão (ffmpeg-decode para decode CPU via libav):
cargo build --release --features ffmpeg-decode
```

### Servidor NVIDIA (GEX44 / GPU)

```bash
cargo build --release --features ffmpeg-decode,ffmpeg-nvdec
export VIDEO_ACCELERATION=gpu   # ou auto
```

Requisitos: driver NVIDIA, FFmpeg com `h264_cuvid`, probe CUDA (`NvdecCapability::probe_cuda_device`).

---

## Verificação local / CI

```bash
bash confvision-rust-processor/scripts/phase-4.1-verify.sh
```

No **Windows** dev: testes default OK; builds com FFmpeg exigem **Linux/Docker** (pkg-config + dev libs).

---

## Próxima fase

**4.1A — Alinhamento workers** — [FASE_4_1A_FECHAMENTO.md](./FASE_4_1A_FECHAMENTO.md) · `scripts/phase-4.1a-verify.sh` (CT111 sem cargo). Depois **5** (rampa de câmeras medida).

Ver também: [FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md), [OPCAO_A_RTSP_DIRETO.md](./OPCAO_A_RTSP_DIRETO.md).
