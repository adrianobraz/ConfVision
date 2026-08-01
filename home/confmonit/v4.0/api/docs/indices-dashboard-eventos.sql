-- Indices recomendados para dashboard de eventos (menu principal FranqueadoPro)
-- Executar no MySQL/MariaDB de producao apos backup.
-- Verificar indices existentes: SHOW INDEX FROM evento; SHOW INDEX FROM processo;

-- 1) Filtro por franqueado (subquery antiga / JOIN atual)
CREATE INDEX IF NOT EXISTS idx_cliente_franqueado
    ON cliente (ID_Franqueado);

-- 2) JOIN processo -> dispositivo -> cliente
CREATE INDEX IF NOT EXISTS idx_dispositivo_cliente
    ON dispositivo (ID_Cliente);

CREATE INDEX IF NOT EXISTS idx_processo_dispositivo_data
    ON processo (ID_Dispositivo, DataCriacao);

-- Alternativa se a consulta filtrar muito por data antes do dispositivo:
-- CREATE INDEX idx_processo_data_dispositivo ON processo (DataCriacao, ID_Dispositivo);

-- 3) JOIN evento -> processo (critico: tabela evento costuma ser a maior)
CREATE INDEX IF NOT EXISTS idx_evento_processo
    ON evento (ID_Processo);

-- 4) ORDER BY no relatorio ao clicar no KPI
CREATE INDEX IF NOT EXISTS idx_evento_processo_dataentrada
    ON evento (ID_Processo, DataEntrada DESC);

-- 5) JOIN com contactId (grupo do evento)
CREATE INDEX IF NOT EXISTS idx_contactid_vinculo_codigo
    ON contactId (ID_Vinculo, Codigo);

CREATE INDEX IF NOT EXISTS idx_contactid_vinculo_grupo
    ON contactId (ID_Vinculo, Grupo);

-- 6) JOINs opcionais na listagem detalhada
CREATE INDEX IF NOT EXISTS idx_usuariosalarme_disp_codigo
    ON usuariosAlarme (ID_Dispositivo, Codigo);

CREATE INDEX IF NOT EXISTS idx_setoralarme_disp_numero
    ON setorAlarme (ID_Dispositivo, Numero);

-- Validar plano de execucao (substituir IDs e datas reais):
-- EXPLAIN SELECT COUNT(*) FROM evento
--   INNER JOIN processo ON evento.ID_Processo = processo.ID_Processo
--   INNER JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
--   INNER JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
--   WHERE cliente.ID_Franqueado = '...'
--   AND processo.DataCriacao >= '2026-06-28 00:00:00'
--   AND processo.DataCriacao <= '2026-07-05 23:59:59';
