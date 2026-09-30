-- Log de transferencias de vinculo (admConfmonit / apifunction)
-- Rodar no banco confmonitV4

CREATE TABLE IF NOT EXISTS transferencia_vinculo_log (
  ID_Log         VARCHAR(64)  NOT NULL PRIMARY KEY,
  DataOperacao   DATETIME     NOT NULL,
  ID_Usuario     VARCHAR(64)  NOT NULL,
  Tipo           ENUM('CLI','FRA','REP') NOT NULL,
  ID_Origem      VARCHAR(64)  NOT NULL,
  ID_Destino     VARCHAR(64)  NOT NULL,
  VinculoAntigo  VARCHAR(64)  NULL,
  VinculoNovo    VARCHAR(64)  NULL,
  Motivo         TEXT         NULL,
  PreviewJSON    JSON         NULL,
  Status         ENUM('OK','ERRO') NOT NULL,
  ErroMsg        TEXT         NULL,
  INDEX idx_transferencia_data (DataOperacao DESC),
  INDEX idx_transferencia_tipo (Tipo, DataOperacao DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
