-- Módulo Contrato (admConfmonit / apifunction) — MySQL confmonitV4
-- Rodar após transferencia_vinculo_log.sql

-- Catálogo de produtos (software/módulos)
CREATE TABLE IF NOT EXISTS fp_catalogo_produto (
  ID_CatalogoProduto  INT AUTO_INCREMENT PRIMARY KEY,
  -- IDCentralUUID obrigatorio na aplicacao; nao usar CENTRAL hardcoded
  ID_Central          VARCHAR(64) NOT NULL DEFAULT '',
  Produto             VARCHAR(32) NOT NULL,
  Plano               VARCHAR(32) NOT NULL,
  NomeExibicao        VARCHAR(120) NOT NULL,
  GrupoUI             VARCHAR(32) NULL,
  TipoSelecao         ENUM('radio','checkbox') NOT NULL DEFAULT 'checkbox',
  ValorMensal         DECIMAL(12,2) NOT NULL DEFAULT 0,
  RetencaoDias        INT NULL DEFAULT 30,
  ModulosJSON         JSON NULL,
  LimitesJSON         JSON NULL,
  Ordem               INT NOT NULL DEFAULT 0,
  Ativo               CHAR(1) NOT NULL DEFAULT 'S',
  CreatedAt           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cat_prod (ID_Central, Produto, Plano),
  KEY idx_cat_central (ID_Central, Ativo)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Pacotes de cotas
CREATE TABLE IF NOT EXISTS fp_pacote_cota (
  ID_PacoteCota   INT AUTO_INCREMENT PRIMARY KEY,
  -- IDCentralUUID obrigatorio na aplicacao; nao usar CENTRAL hardcoded
  ID_Central      VARCHAR(64) NOT NULL DEFAULT '',
  Nome            VARCHAR(80) NOT NULL,
  Quantidade      INT NOT NULL,
  Valor           DECIMAL(12,2) NOT NULL DEFAULT 0,
  Ativo           CHAR(1) NOT NULL DEFAULT 'S',
  Observacao      TEXT NULL,
  CreatedAt       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_pacote_central (ID_Central, Ativo)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Catálogo licenças ConfVision por câmera
CREATE TABLE IF NOT EXISTS fp_catalogo_cv_licenca (
  ID_CatalogoCV   INT AUTO_INCREMENT PRIMARY KEY,
  -- IDCentralUUID obrigatorio na aplicacao; nao usar CENTRAL hardcoded
  ID_Central      VARCHAR(64) NOT NULL DEFAULT '',
  Plano           VARCHAR(64) NOT NULL,
  NomeExibicao    VARCHAR(120) NOT NULL,
  Unidade         ENUM('camera','gravacao') NOT NULL DEFAULT 'camera',
  ValorMensal     DECIMAL(12,2) NOT NULL DEFAULT 0,
  Ordem           INT NOT NULL DEFAULT 0,
  Ativo           CHAR(1) NOT NULL DEFAULT 'S',
  CreatedAt       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cv_lic (ID_Central, Plano),
  KEY idx_cv_central (ID_Central, Ativo)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Contrato (1 ativo por franqueado — enforce na aplicação)
CREATE TABLE IF NOT EXISTS fp_contrato (
  ID_Contrato           INT AUTO_INCREMENT PRIMARY KEY,
  ID_Franqueado         CHAR(36) NOT NULL,
  ID_Central            VARCHAR(64) NOT NULL,
  ID_Representante      CHAR(36) NULL,
  Nome                  VARCHAR(120) NOT NULL,
  Status                ENUM('rascunho','aguardando_pagamento','ativo','suspenso','cancelado') NOT NULL DEFAULT 'rascunho',
  ValorBase             DECIMAL(12,2) NOT NULL DEFAULT 0,
  DescontoTipo          ENUM('percentual','valor') NULL,
  DescontoValor         DECIMAL(12,2) NULL DEFAULT 0,
  ValorFinal            DECIMAL(12,2) NOT NULL DEFAULT 0,
  Periodicidade         ENUM('quinzenal','mensal','bimestral','trimestral','semestral','anual') NOT NULL DEFAULT 'mensal',
  InicioEm              DATE NOT NULL,
  ProximoReajusteEm     DATE NULL,
  ProximaCobrancaEm     DATE NULL,
  VencimentoDia         TINYINT NULL,
  PermiteExcedente      CHAR(1) NOT NULL DEFAULT 'S',
  ValorUnitarioCamera   DECIMAL(12,2) NULL,
  Observacao            TEXT NULL,
  CriadoPor             CHAR(36) NOT NULL,
  CreatedAt             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UpdatedAt             DATETIME NULL ON UPDATE CURRENT_TIMESTAMP,
  Versao                INT NOT NULL DEFAULT 1,
  ModulosJSON           JSON NULL,
  LimitesJSON           JSON NULL,
  KEY idx_contrato_fra (ID_Franqueado),
  KEY idx_contrato_status (Status, ProximaCobrancaEm),
  KEY idx_contrato_central (ID_Central, Status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Itens do contrato
CREATE TABLE IF NOT EXISTS fp_contrato_item (
  ID_ContratoItem   INT AUTO_INCREMENT PRIMARY KEY,
  ID_Contrato       INT NOT NULL,
  Tipo              ENUM('produto','cota','cv_licenca') NOT NULL,
  Chave             VARCHAR(64) NOT NULL,
  Descricao         VARCHAR(200) NULL,
  Quantidade        INT NOT NULL DEFAULT 1,
  QtdPorPacote      INT NULL,
  ValorUnitario     DECIMAL(12,2) NOT NULL DEFAULT 0,
  ValorTotal        DECIMAL(12,2) NOT NULL DEFAULT 0,
  RefID             INT NULL,
  KEY idx_item_contrato (ID_Contrato),
  CONSTRAINT fk_item_contrato FOREIGN KEY (ID_Contrato) REFERENCES fp_contrato(ID_Contrato) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Faturas
CREATE TABLE IF NOT EXISTS fp_fatura_contrato (
  ID_Fatura         INT AUTO_INCREMENT PRIMARY KEY,
  ID_Contrato       INT NOT NULL,
  ID_Franqueado     CHAR(36) NOT NULL,
  ID_Central        VARCHAR(64) NOT NULL,
  ID_Representante  CHAR(36) NULL,
  Tipo              ENUM('contrato_inicial','renovacao','excedente','reajuste') NOT NULL,
  Referencia        VARCHAR(20) NOT NULL,
  Status            ENUM('aberta','paga','cancelada','vencida','estornada') NOT NULL DEFAULT 'aberta',
  ValorTotal        DECIMAL(12,2) NOT NULL,
  VencimentoEm      DATE NOT NULL,
  PagoEm            DATETIME NULL,
  CicloRef          VARCHAR(32) NOT NULL,
  ID_FaturaContabil INT NULL,
  Observacao        TEXT NULL,
  CreatedAt         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_fatura_fra (ID_Franqueado, Status),
  KEY idx_fatura_contrato (ID_Contrato),
  KEY idx_fatura_venc (Status, VencimentoEm),
  CONSTRAINT fk_fatura_contrato FOREIGN KEY (ID_Contrato) REFERENCES fp_contrato(ID_Contrato)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Uso / excedente por ciclo
CREATE TABLE IF NOT EXISTS fp_contrato_uso (
  ID_Contrato     INT NOT NULL,
  CicloRef        VARCHAR(32) NOT NULL,
  Chave           VARCHAR(64) NOT NULL,
  LimiteContrato  INT NOT NULL DEFAULT 0,
  UsoAtual        INT NOT NULL DEFAULT 0,
  Excedente       INT NOT NULL DEFAULT 0,
  PRIMARY KEY (ID_Contrato, CicloRef, Chave),
  CONSTRAINT fk_uso_contrato FOREIGN KEY (ID_Contrato) REFERENCES fp_contrato(ID_Contrato) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Log financeiro contrato
CREATE TABLE IF NOT EXISTS fp_contrato_log (
  ID_Log          INT AUTO_INCREMENT PRIMARY KEY,
  ID_Contrato     INT NULL,
  ID_Fatura       INT NULL,
  ID_Franqueado   CHAR(36) NULL,
  Acao            VARCHAR(64) NOT NULL,
  Detalhe         TEXT NULL,
  ID_Usuario      CHAR(36) NULL,
  CreatedAt       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_log_fra (ID_Franqueado)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
