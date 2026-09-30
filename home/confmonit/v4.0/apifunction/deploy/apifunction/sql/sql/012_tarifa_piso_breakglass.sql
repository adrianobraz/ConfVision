-- Piso Break-glass para tarifas operacionais (MySQL confmonitV4)
ALTER TABLE fp_tarifa_operacional
    ADD COLUMN IF NOT EXISTS PisoTentativa DECIMAL(12,4) NOT NULL DEFAULT 0 AFTER ValorUnidade,
    ADD COLUMN IF NOT EXISTS PisoMinuto DECIMAL(12,4) NOT NULL DEFAULT 0 AFTER PisoTentativa,
    ADD COLUMN IF NOT EXISTS PisoUnidade DECIMAL(12,4) NOT NULL DEFAULT 0 AFTER PisoMinuto;

UPDATE fp_tarifa_operacional
SET PisoTentativa = ValorTentativa,
    PisoMinuto = ValorMinuto,
    PisoUnidade = ValorUnidade
WHERE PisoTentativa = 0 AND PisoMinuto = 0 AND PisoUnidade = 0
  AND (ValorTentativa > 0 OR ValorMinuto > 0 OR ValorUnidade > 0);
