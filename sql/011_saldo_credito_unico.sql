-- Saldo prepago unico por franqueado (canal = 'credito').
-- Consolida saldos antigos por servico (ligacao/sms/whatsapp/email) em uma carteira.

INSERT INTO ops_credito_saldo (id_franqueado, canal, saldo, saldo_inicial, updated_at)
SELECT
    id_franqueado,
    'credito',
    COALESCE(SUM(saldo), 0),
    GREATEST(COALESCE(SUM(saldo), 0), COALESCE(MAX(saldo_inicial), 0)),
    NOW()
FROM ops_credito_saldo
WHERE canal IN ('ligacao', 'sms', 'whatsapp', 'email')
GROUP BY id_franqueado
ON CONFLICT (id_franqueado, canal) DO UPDATE SET
    saldo = EXCLUDED.saldo,
    saldo_inicial = GREATEST(ops_credito_saldo.saldo_inicial, EXCLUDED.saldo_inicial),
    updated_at = NOW();

DELETE FROM ops_credito_saldo
WHERE canal IN ('ligacao', 'sms', 'whatsapp', 'email');
