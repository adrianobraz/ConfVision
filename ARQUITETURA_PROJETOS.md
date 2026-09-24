# ConfVision — Arquitetura dos repositórios e pastas

Este documento evita confusão entre **GitHub oficial**, **ambiente de desenvolvimento (`core4`)** e **API Go**, que são coisas diferentes.

## Repositório oficial (GitHub / Python)

| Item | Valor |
|------|--------|
| **Clone local** | `C:\sistemaconfmonit\ConfVision` |
| **Remote** | [https://github.com/adrianobraz/ConfVision](https://github.com/adrianobraz/ConfVision) |

Esta pasta é o **projeto ConfVision versionado no GitHub**. Contém principalmente:

- código **Python** (workers, MediaMTX guard, DVR, motion, timelapse, etc.)
- **MediaMTX** (Docker/deploy)
- **SQL** e migrações
- documentação de deploy (EasyPanel, variáveis VPS)
- **`confvision-rust-processor/`** — serviço **Rust** separado (processing plane), publicado **neste mesmo repositório**, mas **não substitui** o Python

O **Rust não substitui o Python** neste momento. Python e Rust são processadores que **coexistem** e conversam com a mesma API Go.

## Ambiente de desenvolvimento (`core4`)

O monorepo local **`C:\sistemaconfmonit\core4`** é um ambiente de desenvolvimento com vários componentes. **Não** deve ser tratado como cópia 1:1 do clone GitHub.

Detalhes (ambiente dev local): `C:\sistemaconfmonit\core4\ARQUITETURA_CORE4.md` — arquivo **fora** deste repositório GitHub.

**Regra crítica:** não fazer `git pull`, `merge` ou `rebase` entre o histórico do `core4` e o clone `C:\sistemaconfmonit\ConfVision` **sem análise explícita** — os históricos podem divergir mesmo com o mesmo URL remoto.

## API Go (fora deste repositório)

A aplicação/API principal em **Go** **não** vive neste repositório GitHub.

| Item | Valor |
|------|--------|
| **Caminho local** | `C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision` |

O Go continua sendo o **backend/API** (PostgreSQL, rotas `vis_*`, sync de câmeras, ping de workers).

O **Rust** é um processador externo que consome essa API — ver [`confvision-rust-processor/README.md`](confvision-rust-processor/README.md).

## Diagrama (visão geral)

```
                ┌──────────────────────┐
                │       Go API         │
                │  PostgreSQL / API    │
                └──────────┬───────────┘
                           │
                sincronização / worker
                           │
          ┌────────────────┴────────────────┐
          │                                 │
 ┌────────▼────────┐              ┌────────▼────────┐
 │ Python Worker   │              │ Rust Processor   │
 │ (este repo)     │              │ (confvision-     │
 │                 │              │  rust-processor) │
 └────────┬────────┘              └────────┬────────┘
          │                                 │
          └──────────────┬──────────────────┘
                         │
                   ┌─────▼─────┐
                   │ MediaMTX  │
                   │ RTSP      │
                   └───────────┘
```

## Onde alterar cada coisa

| Tipo de alteração | Onde trabalhar |
|-------------------|----------------|
| Python / MediaMTX / SQL / deploy deste repo | `C:\sistemaconfmonit\ConfVision` (e espelho dev em `core4\confvision`) |
| Rust processor | `confvision-rust-processor/` neste repo; fonte dev em `core4\confvision-rust-processor` |
| Go API | `core4\home\confmonit\v4.0\confvision` — **não mover** para ConfVision |
| Publicar Rust no GitHub | Somente a pasta `confvision-rust-processor/` no repositório oficial — **não** `git add .` no `core4` inteiro |

Documentação do Go (relação com Rust, local): `C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision\README_ARQUITETURA.md`.

## ANTES DE ALTERAR O CONFVISION

1. Identifique se a alteração é **Go**, **Python**, **Rust**, **MediaMTX** ou **infraestrutura/deploy**.
2. **Não** assuma que todas as pastas chamadas `confvision` são o mesmo projeto.
3. Sempre verifique o **caminho completo** no disco.
4. **Não** faça `git pull`, `merge` ou `rebase` entre `core4` e `C:\sistemaconfmonit\ConfVision` sem análise explícita.
5. **Não** use `git add .` no `core4` para publicar o Rust.
6. Para publicar o Rust no GitHub ConfVision, adicione **apenas** `confvision-rust-processor/` ao repositório oficial.
7. **Não** mover o Go atual para ConfVision ou para pastas Python/Rust.
8. **Não** remover o Python atual.
9. **Não** criar Xano novo para esta arquitetura (Rust recusa `CONFVISION_API_URL` em `*.xano.io`).
10. Testes piloto do Rust no **EasyPanel** são feitos como serviço **separado** — ver `confvision-rust-processor/DEPLOY_EASYPANEL.md`.
