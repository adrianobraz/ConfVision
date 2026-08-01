-- Indices para relatorios, listagens e dashboard (FranqueadoPro / API v4)
-- Executar no MySQL/MariaDB apos backup.
-- Verificar existentes: SHOW INDEX FROM nome_tabela;

-- ========== EVENTOS / PROCESSO ==========
CREATE INDEX IF NOT EXISTS idx_evento_processo ON evento (ID_Processo);
CREATE INDEX IF NOT EXISTS idx_evento_processo_dataentrada ON evento (ID_Processo, DataEntrada DESC);
CREATE INDEX IF NOT EXISTS idx_evento_codigo ON evento (Codigo);

CREATE INDEX IF NOT EXISTS idx_processo_dispositivo_data ON processo (ID_Dispositivo, DataCriacao);
CREATE INDEX IF NOT EXISTS idx_processo_dataatenfim ON processo (DataAtenFim);
CREATE INDEX IF NOT EXISTS idx_processo_dispositivo_dataatenfim ON processo (ID_Dispositivo, DataAtenFim);

-- ========== CLIENTE / DISPOSITIVO ==========
CREATE INDEX IF NOT EXISTS idx_cliente_franqueado ON cliente (ID_Franqueado);
CREATE INDEX IF NOT EXISTS idx_cliente_franqueado_ativo ON cliente (ID_Franqueado, Ativo);
CREATE INDEX IF NOT EXISTS idx_cliente_franqueado_nome ON cliente (ID_Franqueado, Nome);

CREATE INDEX IF NOT EXISTS idx_dispositivo_cliente ON dispositivo (ID_Cliente);
CREATE INDEX IF NOT EXISTS idx_dispositivo_dataultimoevento ON dispositivo (DataUltimoEvento);
CREATE INDEX IF NOT EXISTS idx_dispositivo_armado ON dispositivo (Armado);
CREATE INDEX IF NOT EXISTS idx_dispositivo_cliente_armado ON dispositivo (ID_Cliente, Armado);

-- ========== CONTACT ID (grupos de evento) ==========
CREATE INDEX IF NOT EXISTS idx_contactid_vinculo_codigo ON contactId (ID_Vinculo, Codigo);
CREATE INDEX IF NOT EXISTS idx_contactid_vinculo_grupo ON contactId (ID_Vinculo, Grupo);

-- ========== ATENDIMENTO / USUARIOS ==========
CREATE INDEX IF NOT EXISTS idx_usuariosalarme_disp_codigo ON usuariosAlarme (ID_Dispositivo, Codigo);
CREATE INDEX IF NOT EXISTS idx_setoralarme_disp_numero ON setorAlarme (ID_Dispositivo, Numero);

-- ========== SEM COMUNICACAO ==========
CREATE INDEX IF NOT EXISTS idx_listabloqueio_alvo ON listaBloqueio (ID_Alvo);

-- ========== CUSTO ATENDIMENTO (relatorio ligacoes - API legada) ==========
CREATE INDEX IF NOT EXISTS idx_custoatend_franqueado_data ON custoAtendimento (ID_Franqueado, DataOperacao);
CREATE INDEX IF NOT EXISTS idx_custoatend_cliente ON custoAtendimento (ID_Cliente, DataOperacao);
