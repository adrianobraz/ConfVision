-- ConfService — prefixo cs_
-- MySQL 8+

CREATE DATABASE IF NOT EXISTS confservice
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE confservice;

CREATE TABLE IF NOT EXISTS cs_parceiro (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  razao_social VARCHAR(180) NOT NULL,
  nome_fantasia VARCHAR(180) NULL,
  cnpj VARCHAR(20) NULL,
  email VARCHAR(180) NOT NULL,
  telefone VARCHAR(30) NULL,
  software ENUM('MONI','SEGWARE','DGUARD','GENERICO') NOT NULL DEFAULT 'GENERICO',
  webhook_url VARCHAR(500) NOT NULL DEFAULT '',
  webhook_token VARCHAR(255) NOT NULL DEFAULT '',
  preco_cliente_quinzena DECIMAL(12,2) NOT NULL DEFAULT 0,
  comissao_pct DECIMAL(5,2) NULL COMMENT 'NULL = usa COMISSAO_PADRAO_PCT do .env',
  ativo TINYINT(1) NOT NULL DEFAULT 1,
  senha_hash VARCHAR(255) NOT NULL,
  UNIQUE KEY uk_cs_parceiro_email (email)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_cliente_vinculo (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  id_franqueado VARCHAR(64) NOT NULL,
  id_cliente VARCHAR(64) NOT NULL,
  nome_cliente VARCHAR(180) NULL,
  id_parceiro CHAR(36) NOT NULL,
  conta_externa VARCHAR(120) NULL COMMENT 'codigo/conta no software do parceiro',
  preco_congelado DECIMAL(12,2) NOT NULL,
  ativo TINYINT(1) NOT NULL DEFAULT 1,
  inicio_em DATE NOT NULL,
  fim_em DATE NULL,
  KEY idx_cs_vinculo_cliente (id_franqueado, id_cliente, ativo),
  KEY idx_cs_vinculo_parceiro (id_parceiro),
  KEY idx_cs_vinculo_franqueado (id_franqueado),
  CONSTRAINT fk_cs_vinculo_parceiro FOREIGN KEY (id_parceiro) REFERENCES cs_parceiro(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_webhook_fila (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  id_vinculo CHAR(36) NOT NULL,
  id_parceiro CHAR(36) NOT NULL,
  id_franqueado VARCHAR(64) NOT NULL,
  id_cliente VARCHAR(64) NOT NULL,
  alarm_events_id BIGINT NULL,
  payload_json JSON NOT NULL,
  status ENUM('PENDENTE','ENVIANDO','OK','ERRO') NOT NULL DEFAULT 'PENDENTE',
  tentativas INT NOT NULL DEFAULT 0,
  proximo_em DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  ultimo_erro TEXT NULL,
  http_status INT NULL,
  response_body MEDIUMTEXT NULL,
  KEY idx_cs_fila_status_proximo (status, proximo_em),
  KEY idx_cs_fila_parceiro (id_parceiro),
  CONSTRAINT fk_cs_fila_vinculo FOREIGN KEY (id_vinculo) REFERENCES cs_cliente_vinculo(id),
  CONSTRAINT fk_cs_fila_parceiro FOREIGN KEY (id_parceiro) REFERENCES cs_parceiro(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_fatura_franqueado (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  id_franqueado VARCHAR(64) NOT NULL,
  periodo_inicio DATE NOT NULL,
  periodo_fim DATE NOT NULL,
  valor_bruto DECIMAL(12,2) NOT NULL DEFAULT 0,
  status ENUM('ABERTA','PAGA','CANCELADA') NOT NULL DEFAULT 'ABERTA',
  UNIQUE KEY uk_cs_fat_franq_periodo (id_franqueado, periodo_inicio, periodo_fim)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_fatura_parceiro (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  id_parceiro CHAR(36) NOT NULL,
  periodo_inicio DATE NOT NULL,
  periodo_fim DATE NOT NULL,
  valor_bruto DECIMAL(12,2) NOT NULL DEFAULT 0,
  comissao DECIMAL(12,2) NOT NULL DEFAULT 0,
  valor_liquido DECIMAL(12,2) NOT NULL DEFAULT 0,
  status ENUM('ABERTA','PAGA','CANCELADA') NOT NULL DEFAULT 'ABERTA',
  UNIQUE KEY uk_cs_fat_parc_periodo (id_parceiro, periodo_inicio, periodo_fim),
  CONSTRAINT fk_cs_fat_parceiro FOREIGN KEY (id_parceiro) REFERENCES cs_parceiro(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_fatura_item (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  id_fatura_franqueado CHAR(36) NOT NULL,
  id_fatura_parceiro CHAR(36) NOT NULL,
  id_vinculo CHAR(36) NOT NULL,
  id_franqueado VARCHAR(64) NOT NULL,
  id_cliente VARCHAR(64) NOT NULL,
  id_parceiro CHAR(36) NOT NULL,
  preco DECIMAL(12,2) NOT NULL,
  comissao DECIMAL(12,2) NOT NULL,
  liquido_parceiro DECIMAL(12,2) NOT NULL,
  KEY idx_cs_item_fat_franq (id_fatura_franqueado),
  KEY idx_cs_item_fat_parc (id_fatura_parceiro),
  CONSTRAINT fk_cs_item_fat_franq FOREIGN KEY (id_fatura_franqueado) REFERENCES cs_fatura_franqueado(id),
  CONSTRAINT fk_cs_item_fat_parc FOREIGN KEY (id_fatura_parceiro) REFERENCES cs_fatura_parceiro(id),
  CONSTRAINT fk_cs_item_vinculo FOREIGN KEY (id_vinculo) REFERENCES cs_cliente_vinculo(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_parceiro_rep (
  id_central       VARCHAR(64) NOT NULL,
  id_representante VARCHAR(64) NOT NULL,
  id_parceiro      CHAR(36) NOT NULL,
  ativo            TINYINT(1) NOT NULL DEFAULT 1,
  created_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id_central, id_representante, id_parceiro),
  KEY idx_cs_pr_rep (id_representante, ativo),
  CONSTRAINT fk_cs_pr_parceiro FOREIGN KEY (id_parceiro) REFERENCES cs_parceiro(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_parceiro_franqueado (
  id_representante VARCHAR(64) NOT NULL,
  id_franqueado    VARCHAR(64) NOT NULL,
  id_parceiro      CHAR(36) NOT NULL,
  ativo            TINYINT(1) NOT NULL DEFAULT 1,
  created_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id_representante, id_franqueado, id_parceiro),
  KEY idx_cs_pf_fq (id_franqueado, ativo),
  CONSTRAINT fk_cs_pf_parceiro FOREIGN KEY (id_parceiro) REFERENCES cs_parceiro(id) ON DELETE CASCADE
) ENGINE=InnoDB;
