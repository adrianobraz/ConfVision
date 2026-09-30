-- Zera dados financeiros MySQL legado (apifunction / contrato / tarifas)
-- Executar APOS zerar Xano. Nao remove estrutura.
-- IMPORTANTE: executar o bloco INTEIRO (phpMyAdmin: colar tudo de uma vez).
-- Uso: mysql -u ... -p ... < 011_zero_financeiro_mysql.sql

SET FOREIGN_KEY_CHECKS = 0;

TRUNCATE TABLE fp_contrato_log;
TRUNCATE TABLE fp_fatura_contrato;
TRUNCATE TABLE fp_contrato_uso;
TRUNCATE TABLE fp_contrato_item;
TRUNCATE TABLE fp_contrato;
TRUNCATE TABLE fp_catalogo_cv_licenca;
TRUNCATE TABLE fp_pacote_cota;
TRUNCATE TABLE fp_catalogo_produto;

TRUNCATE TABLE fp_credito_recarga;
TRUNCATE TABLE fp_tarifa_operacional_rep;
TRUNCATE TABLE fp_tarifa_operacional;

SET FOREIGN_KEY_CHECKS = 1;
