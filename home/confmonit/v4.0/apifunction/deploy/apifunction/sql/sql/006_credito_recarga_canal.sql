-- Saldo unico: recarga credita carteira canal=credito (nao mais por servico).
-- Aplicar apos 005_servicos_credito_schema.sql

ALTER TABLE fp_credito_recarga
  MODIFY COLUMN Canal ENUM('credito','ligacao','sms','whatsapp','email') NOT NULL DEFAULT 'credito';
