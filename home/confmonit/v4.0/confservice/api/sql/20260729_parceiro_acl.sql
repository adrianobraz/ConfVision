-- ACL parceiro monitoramento: Central -> Representante -> Franqueado
-- Banco: confservice

USE confservice;

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
