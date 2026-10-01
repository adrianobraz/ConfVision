-- Bonificacao go-live: R$ 9999 canal ligacao para franqueados existentes (NAO contabiliza).
-- Rodar UMA vez antes do deploy. Novos franqueados nascem com saldo zero (aplicacao).

DO $$
DECLARE
    v_count INT;
    v_done INT;
BEGIN
    SELECT COUNT(*) INTO v_done FROM ops_credito_go_live;
    IF v_done > 0 THEN
        RAISE NOTICE '010_bonificacao_go_live ja executado — abortando';
        RETURN;
    END IF;

    INSERT INTO ops_credito_saldo (id_franqueado, canal, saldo, saldo_inicial, updated_at)
    SELECT DISTINCT f.id_franqueado, 'ligacao', 9999, 9999, NOW()
    FROM (
        SELECT id_franqueado FROM ops_franqueado_atendimento_politica
        UNION
        SELECT id_franqueado FROM ops_stg_cliente_ativo
        UNION
        SELECT id_franqueado FROM ops_credito_saldo
    ) f
    WHERE TRIM(f.id_franqueado) <> ''
    ON CONFLICT (id_franqueado, canal) DO UPDATE
        SET saldo = EXCLUDED.saldo,
            saldo_inicial = EXCLUDED.saldo_inicial,
            updated_at = NOW();

    GET DIAGNOSTICS v_count = ROW_COUNT;

    INSERT INTO ops_credito_movimento (
        id_franqueado, canal, tipo, quantidade, valor_total, observacao, criado_por
    )
    SELECT DISTINCT f.id_franqueado, 'ligacao', 'bonificacao', 9999, 9999,
           'Bonificacao go-live servicos — nao contabiliza', 'sistema'
    FROM (
        SELECT id_franqueado FROM ops_franqueado_atendimento_politica
        UNION
        SELECT id_franqueado FROM ops_stg_cliente_ativo
    ) f
    WHERE TRIM(f.id_franqueado) <> ''
      AND NOT EXISTS (
          SELECT 1 FROM ops_credito_movimento m
          WHERE m.id_franqueado = f.id_franqueado
            AND m.canal = 'ligacao'
            AND m.tipo = 'bonificacao'
            AND m.observacao LIKE 'Bonificacao go-live%'
      );

    INSERT INTO ops_credito_go_live (qtd_franqueados, valor_bonificado, observacao)
    VALUES (v_count, 9999, 'bonificacao ligacao go-live');

    RAISE NOTICE 'Bonificacao aplicada a % registros saldo', v_count;
END $$;
