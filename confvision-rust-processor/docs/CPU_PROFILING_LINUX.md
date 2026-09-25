# Profiling de CPU real (Linux)

Instrumentação `rtsp_hotpath` mede **tempo de parede** em `session.next()` (inclui `await`/I/O TCP). Para saber **onde os ciclos de CPU** são gastos no VPS, use `perf` no processo em execução.

## Pré-requisitos (VPS Linux)

- Kernel com `perf_event_open` (Debian/Ubuntu: pacote `linux-tools-$(uname -r)` ou `linux-perf`).
- Permissão para gravar no processo:
  - `sudo sysctl -w kernel.perf_event_paranoid=2` (ou `-1` temporário em janela controlada), **ou**
  - executar o script com `sudo`.
- Binário com símbolos (recomendado para stacks legíveis):
  ```bash
  cargo build --profile profiling --features ffmpeg-decode
  ```
  O profile `profiling` herda `release`, mantém LTO e **não faz strip** (`debug = line-tables-only`). O `cargo build --release` de produção **não muda**.

- Opcional (não instalado pelo projeto): [FlameGraph](https://github.com/brendangregg/FlameGraph) no `PATH` (`stackcollapse-perf.pl`, `flamegraph.pl`) para gerar SVG.

## Worker em execução

Use o mesmo deploy/processo que já atende `/health` e `/metrics`. Não é necessário carga artificial: capture com ~11 câmeras no estado atual.

Anote o PID ou deixe o script descobrir pelo nome `confvision-rust-processor`.

## Executar coleta (~30 s)

Na imagem de produção (após rebuild), os scripts ficam em **`/app/scripts/`** (`WORKDIR` = `/app`).

No repositório ou no container:

```bash
chmod +x scripts/profile-cpu.sh scripts/rust-process-threads.sh
./scripts/profile-cpu.sh          # auto-descobre PID
./scripts/profile-cpu.sh 12345    # PID explícito
sudo ./scripts/profile-cpu.sh     # se perf negar permissão
```

Dentro do container (threads / preparação; `perf` costuma rodar no **host** com PID do processo no namespace do host):

```bash
docker exec -it <container> /app/scripts/rust-process-threads.sh
# No host (exemplo): PID=$(docker inspect -f '{{.State.Pid}}' <container>)
# sudo perf record -F 99 -g -p "$PID" -o perf.data -- sleep 30
```

Só listar threads (PID, TID, nome, %CPU quando `ps` permitir):

```bash
./scripts/rust-process-threads.sh
./scripts/profile-cpu.sh --threads-only
```

Variáveis úteis: `DURATION=45`, `FREQ=99`, `OUT_DIR=/tmp/cv-prof`.

## Onde ficam os resultados

Diretório `profile-output/run_YYYYMMDD_HHMMSS/` (ou `OUT_DIR`):

| Arquivo | Conteúdo |
|---------|----------|
| `perf.data` | Gravação bruta |
| `perf-report.txt` | Top de símbolos (texto) |
| `perf-top-dso.txt` | Agrupado por DSO (`.so`, binário) |
| `perf-script.txt` | Stacks para flamegraph manual |
| `threads.txt` | Snapshot de threads no início |
| `flamegraph.svg` | Só se FlameGraph estiver no PATH |
| `meta.txt` | PID, duração, exe |

## Flamegraph

Se o script não gerar SVG:

```bash
cd profile-output/run_.../
perf script -i perf.data | stackcollapse-perf.pl | flamegraph.pl > flamegraph.svg
```

Abra `flamegraph.svg` no navegador. Largura ≈ fração amostral de CPU (não tempo bloqueado em I/O).

## O que procurar no relatório

Interpretação **sem assumir** gargalo antes da medição:

| Área | Sinais típicos no `perf report` |
|------|----------------------------------|
| Retina / RTP / H.264 | `retina::`, depacketize, `h264`, RTSP parse |
| Tokio / runtime | `tokio`, `runtime`, park/unpark, executor |
| FFmpeg / decode | `avcodec`, `ffmpeg`, `decode`, `h264_decoder` |
| Locks / sync | `pthread_mutex`, `futex`, `parking_lot`, `RwLock` |
| Rede / syscalls | `recv`, `read`, `epoll_wait`, `socket` (tempo **em syscall** ≠ só I/O wait, mas alto aqui indica custo kernel/polling) |
| Restante Rust | símbolos `confvision_rust_processor::` |

Compare threads em `threads.txt`: várias threads Tokio com CPU alta vs uma thread dominante.

## Interpretação

- **Alta CPU em funções Rust de depacketize** → trabalho ativo no caminho RTSP (diferente de `%` em `rtsp_hotpath.session_next_share`).
- **Alta CPU em `epoll_wait`/`recv` com pouco userland** → muitas wakeups/leituras; correlacionar com número de conexões RTSP.
- **Alta CPU em decode/motion** → downstream (hoje esperado baixo se throttle estiver ativo).
- **Perf sem símbolos** → rebuild com `--profile profiling` e repetir; ou `perf report` com `--symfs` apontando para o binário com debug.

## Segurança / produção

- Janela curta (~30 s), frequência moderada (99 Hz): impacto pequeno, mas prefira horário acordado.
- Não commitar `profile-output/` (artefatos locais).
- Não baixar FlameGraph automaticamente; instale manualmente se quiser SVG.
