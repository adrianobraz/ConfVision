# Deploy — Piso Break-glass (fluxo simples)

## Fluxo

1. Break-glass escolhe a **Central** no combo
2. Define modo **livre** ou **piso**
3. Gera **seed** (Catálogo / ConfVision / À la Carte) para essa Central
4. Edita preços nas telas de catálogo (valor salvo pelo BG = mínimo)
5. Central em modo piso: pode subir, não baixa; pagamento gera `repasse_central_breakglass`

## Push necessário

### Tabelas (ALTERAR — campo novo)

- `fp_produto_catalogo` → `valor_piso_breakglass`
- `fp_modulo_catalogo` → `valor_piso_breakglass`

### Functions

| Arquivo | Ação |
|---------|------|
| `490_fn_fp_central_modo_preco.xs` | criar/ok |
| `492_fn_fp_split_precos.xs` | ALTERAR (piso do item) |
| `493_fn_fp_catalogo_sync_piso_bg.xs` | CRIAR |
| `494_fn_fp_centrais_listar.xs` | CRIAR |
| `480_fn_fp_preco_efetivo.xs` | ALTERAR |
| `436_fn_fp_catalogo_upsert_item.xs` | ALTERAR |
| `455_fn_fp_modulo_catalogo_upsert_item.xs` | ALTERAR |

### APIs

| Path | Ação |
|------|------|
| `fp_catalogo_seed` | ALTERAR (`id_central`) |
| `fp_catalogo_listar` | ALTERAR |
| `fp_catalogo_salvar` | ALTERAR |
| `fp_modulo_catalogo_listar` | ALTERAR |
| `fp_modulo_catalogo_salvar` | ALTERAR |
| `fp_modulo_catalogo_seed` | ALTERAR |
| `fp_central_preco_config_listar` | ALTERAR |
| `fp_central_preco_config_salvar` | ALTERAR (sync piso) |

### Front

Rebuild `admConfmonit`.

## Login Central / Representante e catalogo (importante)

Masters de Central no MySQL usam `ID_Vinculo = "CENTRAL"` + `IDCentralUUID` da empresa.
O login `/v4/admfinanceiro/logar` deve devolver:

- **CEN:** `idVinculo` / `idCentralCatalogo` = **ID_Central** real
- **REP:** `idVinculo` = **ID_Representante** e `idCentralCatalogo` = **ID_Central** do representante
  (via `representante.IDCentralUUID`)

Sem isso, REP/CEN caem no catalogo da matriz (`"CENTRAL"`).

Sessao Xano (`fp_admin_sessao`) guarda `id_vinculo` + `id_central`.

### Deploy obrigatorio (nessa ordem)

1. **Redeploy API Go** (`mAdmFinanceiro.go`)
2. **Push Xano**: tabela `fp_admin_sessao` (`id_central`), `fn_fp_admin_login`, `fn_fp_admin_validar`,
   `fp_preco_rep_listar`, `fp_catalogo_seed`, `fp_catalogo_listar/salvar`, `fp_modulo_catalogo_*`
3. Rebuild `admConfmonit`
4. **Logout + login** (sessao antiga continua com id errado)
5. Na tela Catálogo, confira `Catalogo da Central <id>` — Central do usuario, nao outra

Se o catalogo da Central estiver vazio: no Break-glass escolha essa Central e rode o seed.

## Lista de Centrais (Xano → Go)

`fn_fp_centrais_listar` chama `POST {api_legada_url}/v4/central/lista`.

Erro `Unable to locate var: body.Dados` = bug de parse no Xano (acesso a chave ausente).
Corrigido com `|get:"dados":[]` — **push da function 494**.

Se o Xano cloud **não alcançar** a VPS, o front faz fallback via `/api-v4` / `api-v4-proxy.php`.

Teste: `curl -X POST http://185.130.61.4:2010/v4/central/lista -d "{}"`
