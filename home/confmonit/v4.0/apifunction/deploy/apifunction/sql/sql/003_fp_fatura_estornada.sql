-- Estorno de pagamento — fatura contrato (MySQL confmonitV4)
ALTER TABLE fp_fatura_contrato
  MODIFY COLUMN Status ENUM('aberta','paga','cancelada','vencida','estornada') NOT NULL DEFAULT 'aberta';
