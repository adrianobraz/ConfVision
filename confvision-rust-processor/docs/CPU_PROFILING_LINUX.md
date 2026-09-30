# Profiling de CPU real (Linux)

Instrumentação `rtsp_hotpath` mede **tempo de parede** em `session.next()` (inclui `await`/I/O TCP). Para **ciclos de CPU**, use `perf` no **host** Linux com um binário **não stripped** e DWARF.

## Por que o perf mostra só `0x...`?

O `Dockerfile` de produção gera `release` com `strip = true`. O `perf` precisa do **mesmo build-id** do processo em execução **e** de debug info (profile `profiling`).

**Regra:** durante a janela de profiling, o processo profileado deve ser o binário **`profiling`** (mesmo commit), não o release stripped copiado depois.

## Profile Cargo `[profile.profiling]`

- Herda `release` (LTO, `codegen-units = 1`).
- `strip = false`, `debug = 2` (DWARF para `perf`).
- `cargo build --release` de produção **não muda**.

## Build do binário com símbolos

### No VPS / CI Linux (cargo)

```bash
cd confvision-rust-processor
chmod +x scripts/build-profiling.sh
./scripts/build-profiling.sh
# ou:
cargo build --profile profiling --features ffmpeg-decode
file target/profiling/confvision-rust-processor   # deve conter "not stripped"
readelf -S target/profiling/confvision-rust-processor | grep debug_info
```

### Imagem Docker isolada (não substitui produção normal)

```bash
docker build -f Dockerfile.profiling -t confvision-rust-processor:profiling .
docker create --name cv-prof-tmp confvision-rust-processor:profiling
docker cp cv-prof-tmp:/app/confvision-rust-processor-profiling /tmp/confvision-rust-processor-profiling
docker rm cv-prof-tmp
file /tmp/confvision-rust-processor-profiling
```

No EasyPanel: serviço temporário ou redeploy curto usando **`Dockerfile.profiling`** (mesma tag/branch), depois voltar ao `Dockerfile` normal.

## Worker em execução

1. Subir o worker com imagem/binário **`profiling`** (mesmo código, mesmas envs).
2. Anote PID no host: `PID=$(docker inspect -f '{{.State.Pid}}' <container_id>)`.
3. Exporte o binário com símbolos (caminho no host):

```bash
export CONFVISION_BINARY=/tmp/confvision-rust-processor-profiling
# ou target/profiling/confvision-rust-processor se build local no host
```

## Executar coleta (~30 s)

Scripts em **`/app/scripts/`** na imagem.

```bash
sudo CONFVISION_BINARY=/tmp/confvision-rust-processor-profiling \
  ./scripts/profile-cpu.sh "$PID"
```

Dentro do container (só threads):

```bash
docker exec -it <container> /app/scripts/rust-process-threads.sh
```

Variáveis: `DURATION=30`, `FREQ=99`, `OUT_DIR=/tmp/cv-prof`, `PERF=perf`.

## Artefatos

`profile-output/run_*/`: `perf.data`, `perf-report.txt`, `perf-top-dso.txt`, `threads.txt`, `perf-script.txt`, `binary-file.txt` (se `CONFVISION_BINARY`), `flamegraph.svg` (se FlameGraph no PATH).

## Flamegraph manual

```bash
perf script -i perf.data --buildid-all | stackcollapse-perf.pl | flamegraph.pl > flamegraph.svg
```

## O que procurar

| Área | Sinais no `perf report` |
|------|-------------------------|
| Retina / RTP | `retina::`, depacketize, h264 |
| Tokio | `tokio`, executor, park |
| FFmpeg | `avcodec`, decode |
| Locks | `futex`, mutex |
| Syscalls | `epoll_wait`, `recv`, `read` |
| App | `confvision_rust_processor::` |

## Segurança

- Janela curta (~30 s); não commitar `profile-output/`.
- `perf` só no host; não instalar no container de produção.
