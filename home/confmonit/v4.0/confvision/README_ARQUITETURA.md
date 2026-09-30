# ConfVision — API Go (backend)

## O que é esta pasta

```
C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision
```

Esta é a **aplicação/API principal em Go** do ConfVision — o **control plane** existente.

Inclui, entre outros:

- `app.go`, `go.mod`, `go.sum`
- `src/` e módulos da aplicação
- **`visapi`** — rotas HTTP da API
- **`visdata`** — acesso a dados / PostgreSQL

## Papel na arquitetura

- O **Go** expõe a API usada por workers (sync de câmeras, ping, configuração).
- O **PostgreSQL** é acessado **pelo Go** nesta arquitetura — processadores não substituem isso na Fase 1 do Rust.

```
                ┌──────────────────────┐
                │    Esta API (Go)     │
                │  PostgreSQL / vis_*  │
                └──────────┬───────────┘
                           │
          ┌────────────────┴────────────────┐
          │                                 │
 ┌────────▼────────┐              ┌────────▼────────┐
 │ Python Worker   │              │ Rust Processor   │
 │ (analítico)     │              │ (externo)        │
 └────────┬────────┘              └────────┬────────┘
          └──────────────┬──────────────────┘
                         │
                   ┌─────▼─────┐
                   │ MediaMTX  │
                   └───────────┘
```

## Rust — processador externo

O **ConfVision Rust Processor** **não substitui** esta API.

- Código Rust: `C:\sistemaconfmonit\core4\confvision-rust-processor`
- Documentação: [`confvision-rust-processor/README.md`](../../../../confvision-rust-processor/README.md)
- Integração típica: `GET /vis_camera_sync_ativas`, `POST /vis_worker_ping`, header `X-Vis-Worker-Key` (`VIS_WORKER_API_KEY` no Rust).

Use `worker_tipo=rust_processor` e `PROCESSOR_ID` distinto do Python para coexistência.

## O que NÃO fazer

- **Não** mover este projeto Go para:
  - `C:\sistemaconfmonit\ConfVision`
  - `C:\sistemaconfmonit\core4\confvision`
  - `C:\sistemaconfmonit\core4\confvision-rust-processor`
- **Não** tratar `core4\confvision` (Python) como esta API.
- **Não** assumir que o repositório GitHub [ConfVision](https://github.com/adrianobraz/ConfVision) contém o Go — ele contém **Python**, MediaMTX, SQL e o serviço Rust publicado.

Visão geral dos repositórios: `C:\sistemaconfmonit\ConfVision\ARQUITETURA_PROJETOS.md` e [`ARQUITETURA_CORE4.md`](../../../../ARQUITETURA_CORE4.md).

## ANTES DE ALTERAR O CONFVISION

Confirme o **caminho completo** e o **tipo** da alteração (Go vs Python vs Rust). Regras completas em [`ARQUITETURA_CORE4.md`](../../../../ARQUITETURA_CORE4.md) (seção homônima).
