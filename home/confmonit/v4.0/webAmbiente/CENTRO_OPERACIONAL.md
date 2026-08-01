# Centro Operacional — webAmbiente

Módulo novo no **webAmbiente** (sistema separado). A home continua como versão lite; o Centro Operacional abre em tela dedicada por cliente.

## Acesso

- Home → selecionar cliente com mapas → botão **Centro operacional**
- URL direta: `/centroOperacional/page?idCliente=...&nome=...&idFranqueado=...`

## Itens novos (fora do escopo original do webAmbiente)

### Xano — publicar antes de usar em produção

| Item | Caminho |
|---|---|
| Tabela `co_cliente_config` | `core4/tables/127_co_cliente_config.xs` |
| API get config | `core4/apis/center_operacion/2303_co_cliente_config_get_POST.xs` |
| API salvar config | `core4/apis/center_operacion/2304_co_cliente_config_salvar_POST.xs` |
| Grupo API | `core4/apis/center_operacion/api_group.xs` (`centerOperacion`, canonical `CiSZf6eF`) |

Executar `push_all_changes_to_xano` após revisar os arquivos `.xs`.

### Backend Go (webAmbiente)

| Rota | Descrição |
|---|---|
| `GET /centroOperacional/page` | Tela principal |
| `POST /centroOperacional/eventos` | Fila de eventos (MySQL direto) |
| `POST /centroOperacional/mapasStatus` | Status batch dos mapas |
| `POST /centroOperacional/configCliente` | Avatar/cor (Xano) |
| `POST /centroOperacional/configCliente/salvar` | Salvar avatar/cor |
| `POST /centroOperacional/setoresMapas` | Índice setor→mapa |
| `POST /centroOperacional/dispositivoStatus` | Armado/desarmado |
| `POST /centroOperacional/comando` | Armar/desarmar remoto |
| `GET /audio/*` | Arquivos `audio/1.mp3` e `audio/2.mp3` |

Código: `src/modulos/centroOperacional/`

### Frontend

- `public/templates/centroOperacional/`
- `public/css/centro-operacional.css`
- `public/js/mapaMonitor.js` — suporte a múltiplas instâncias (`MapaMonitor.criar`)

### Variáveis `.env`

```env
XANO_CENTER_OPERACION_API="https://xpcy-oyme-lno7.b2.xano.io/api:CiSZf6eF"
URL_COMANDO="https://terminal.confmonit2.com.br/api-comando/armar"
SENHA_WEB_COMANDO="..."
```

## Funcionalidades implementadas

- Layout 3 colunas: mapas | abas+planta | fila de eventos
- Header: avatar com iniciais e cor, sirene, exportar CSV, relógio com data por extenso
- Faixa vermelha quando som desligado
- Foco automático no mapa do evento (com opção de fixar mapa)
- Piscar lista/aba em alarme quando não selecionados
- Toasts com barra 30s; vermelho piscando se não ACK
- Som: `audio/1.mp3` (padrão)
- Armar/desarmar por mapa (dispositivo do primeiro setor do mapa)
- Exportar eventos em CSV

## Próximas evoluções sugeridas

- Títulos de evento via tabela CTI (como no Terminal)
- Campo `idDispositivo` em `mapa_ambiente` no Xano
- ACK integrado ao fluxo de atendimento do Terminal
- Seletor de ícone no editor de setores
