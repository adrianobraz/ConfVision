-- Liga fatura MySQL (contrato) à fatura contábil Xano (fp_fatura)
ALTER TABLE fp_fatura_contrato
  ADD COLUMN ID_FaturaContabil INT NULL AFTER CicloRef;

CREATE INDEX idx_fatura_contabil ON fp_fatura_contrato (ID_FaturaContabil);
