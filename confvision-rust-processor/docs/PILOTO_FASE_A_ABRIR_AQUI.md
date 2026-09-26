# Fase A — abrir aqui (Postgres + EasyPanel)

**Não cria tabela nova.** Só atualiza a coluna `worker_id` na tabela existente **`vis_camera`**.

Worker Python (**confvision-worker**) fica **parado** até rust-pilot passar na validação (A7). Câmeras 3 e 4 voltam ao `worker_id` do Python no SQL, mas **não processam** enquanto o worker estiver off — isso é esperado.

---

## 1) Postgres — arquivo SQL

| Onde | Caminho |
|------|---------|
| **Cursor / disco (worktree rust-pilot)** | `c:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor\sql\piloto_fase_a_isolamento.sql` |
| **Repo (branch rust-pilot)** | `confvision-rust-processor/sql/piloto_fase_a_isolamento.sql` |

### Como executar

**Opção A — cliente `psql` (Postgres central ConfVision)**

Use a mesma URL que Go/worker (`POSTGRES_URL` no painel ou `VARIAVEIS_VPS.md`):

```bash
psql "postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable" \
  -f confvision-rust-processor/sql/piloto_fase_a_isolamento.sql
```

**Opção B — DBeaver / pgAdmin**

- Host: `191.96.156.116` (ou o host atual do `confmonit`)
- Database: `confmonit`
- Abrir o `.sql` → rodar primeiro os **SELECT**, depois o **BEGIN…COMMIT** (ajuste `worker-docker-21` e câmera `5`).

**Opção C — container `visionpsql` na VPS (se ainda usar Postgres local EasyPanel)**

```bash
docker ps | grep -i visionpsql
docker exec -i NOME_CONTAINER psql -U confmonit -d confmonit < piloto_fase_a_isolamento.sql
```

### Antes do UPDATE

1. Rodar só os **SELECT** do arquivo.
2. Substituir no SQL:
   - `worker-docker-21` → `WORKER_ID` real do worker (quando for subir).
   - Câmera piloto: `5` (ou `3`).
3. Executar **BEGIN … COMMIT**.

---

## 2) EasyPanel — rust-pilot (Environment)

| Onde | Caminho |
|------|---------|
| **Cursor / disco** | `c:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor\easypanel.env.fase-a.example` |
| **Repo** | `confvision-rust-processor/easypanel.env.fase-a.example` |

### Passos no painel

1. [EasyPanel foxpro → rust-pilot → Environment](https://ops.confhost.com.br/projects/foxpro/app/rust-pilot)
2. Copiar **todo** o conteúdo de `easypanel.env.fase-a.example`.
3. Colar no Environment; preencher `VIS_WORKER_API_KEY` e `RTMP_PUBLISH_SECRET` (iguais Go/MediaMTX).
4. **Salvar** → **Restart** (não precisa Implante).
5. Manter **confvision** (MediaMTX) **on** — RTSP da câmera piloto depende disso.
6. **confvision-worker:** deixar **stop** até A7 ok.

---

## 3) Validar (depois do SQL + Restart)

```bash
bash confvision-rust-processor/scripts/validate-piloto-fase-a.sh \
  https://foxpro-rust-pilot.rkr351.easypanel.host
```

Checklist completo: [`PILOTO_FASE_A.md`](./PILOTO_FASE_A.md)

---

## 4) Só então subir o worker

Quando `cameras_total=1`, `cameras_online=1`, logs estáveis:

1. EasyPanel → **confvision-worker** → Play (1 réplica).
2. Confirmar câmeras 3 e 4 (e demais) no sync do Python (`WORKER_ID` igual ao SQL).
