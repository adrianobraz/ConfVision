-- Servicos operacionais: tarifas + recarga de credito (MySQL confmonitV4)

CREATE TABLE IF NOT EXISTS fp_tarifa_operacional (
  ID_Tarifa         INT AUTO_INCREMENT PRIMARY KEY,
  -- IDCentralUUID obrigatorio na aplicacao; nao usar CENTRAL hardcoded
  ID_Central        VARCHAR(64) NOT NULL DEFAULT '',
  Canal             ENUM('ligacao','sms','whatsapp','email') NOT NULL,
  ValorTentativa    DECIMAL(12,4) NOT NULL DEFAULT 0,
  ValorMinuto       DECIMAL(12,4) NOT NULL DEFAULT 0,
  ValorUnidade      DECIMAL(12,4) NOT NULL DEFAULT 0,
  Ativo             CHAR(1) NOT NULL DEFAULT 'S',
  UpdatedAt         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tarifa_central_canal (ID_Central, Canal)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fp_tarifa_operacional_rep (
  ID_TarifaRep      INT AUTO_INCREMENT PRIMARY KEY,
  ID_Representante  CHAR(36) NOT NULL,
  ID_Central        VARCHAR(64) NOT NULL,
  Canal             ENUM('ligacao','sms','whatsapp','email') NOT NULL,
  ValorTentativa    DECIMAL(12,4) NOT NULL DEFAULT 0,
  ValorMinuto       DECIMAL(12,4) NOT NULL DEFAULT 0,
  ValorUnidade      DECIMAL(12,4) NOT NULL DEFAULT 0,
  Ativo             CHAR(1) NOT NULL DEFAULT 'S',
  UpdatedAt         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tarifa_rep_canal (ID_Representante, Canal)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fp_credito_recarga (
  ID_Recarga            INT AUTO_INCREMENT PRIMARY KEY,
  ID_Franqueado         CHAR(36) NOT NULL,
  ID_Central            VARCHAR(64) NOT NULL,
  ID_Representante      CHAR(36) NULL,
  Canal                 ENUM('credito','ligacao','sms','whatsapp','email') NOT NULL DEFAULT 'credito',
  Valor                 DECIMAL(12,2) NOT NULL,
  Status                ENUM('aberta','paga','cancelada') NOT NULL DEFAULT 'aberta',
  SolicitadoPor         ENUM('FRA','CEN','REP','BG') NOT NULL DEFAULT 'FRA',
  SolicitadoUsuario     VARCHAR(120) NULL,
  ID_FaturaContabil     INT NULL,
  CreditoAplicado       CHAR(1) NOT NULL DEFAULT 'N',
  CreatedAt             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PagoEm                DATETIME NULL,
  KEY idx_recarga_fra (ID_Franqueado, Status),
  KEY idx_recarga_fatura (ID_FaturaContabil)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
