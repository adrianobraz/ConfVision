-- Feedback de clientes (bug, melhoria, ideia) — ConfMonit
-- Executar no MySQL/MariaDB apos backup.

CREATE TABLE IF NOT EXISTS cm_feedback_cliente (
  ID_Feedback    VARCHAR(20)  NOT NULL,
  Software       VARCHAR(30)  NOT NULL COMMENT 'franqueadopro, confvision, webambiente, dialyze',
  Tipo           ENUM('bug', 'melhoria', 'ideia') NOT NULL,
  Descricao      TEXT         NOT NULL,
  URL_Pagina     VARCHAR(500) NULL,
  ID_Usuario     VARCHAR(20)  NULL,
  ID_Franqueado  VARCHAR(20)  NULL,

  Visto          CHAR(1)      NOT NULL DEFAULT 'N' COMMENT 'S = visto pela equipe, N = nao visto',
  Status         VARCHAR(200) NOT NULL DEFAULT 'novo' COMMENT 'ex: com a equipe de desenvolvimento',
  Resumo         TEXT         NULL COMMENT 'ex: foi adicionado essa melhoria no sistema',

  Data_Cadastro  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  Data_Visto     DATETIME     NULL,

  PRIMARY KEY (ID_Feedback),
  INDEX idx_software_data (Software, Data_Cadastro),
  INDEX idx_franqueado (ID_Franqueado),
  INDEX idx_visto (Visto),
  INDEX idx_status (Status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
