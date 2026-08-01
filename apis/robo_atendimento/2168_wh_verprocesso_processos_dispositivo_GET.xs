query "wh_verprocesso/processos/dispositivo" verb=GET {
  api_group = "RoboAtendimento"

  input {
    text idDispositivo? filters=trim
  }

  stack {
    !db.direct_query {
      sql = """
        WITH eventos AS (
          SELECT
            e.*,
        
            CASE
              WHEN (e."created_at"::bigint) > 9999999999
                THEN to_timestamp((e."created_at"::bigint / 1000))
              ELSE to_timestamp(e."created_at"::bigint)
            END AS data_evento
        
          FROM x1_3 e
          WHERE e."idDispositivo" = '{{ $input.idDispositivo }}'
        ),
        
        processos AS (
          SELECT
            "idProcesso",
        
            MIN(data_evento) AS inicio_evento,
            MAX(data_evento) AS fim_evento,
        
            COUNT(*) AS total_eventos,
        
            MAX(CASE WHEN "ctiGrupo" = 'ALARME' THEN 1 ELSE 0 END) AS tem_alarme,
            MAX(CASE WHEN "ctiGrupo" = 'RESTAURE' THEN 1 ELSE 0 END) AS tem_restaure,
            MAX(CASE WHEN "ctiGrupo" = 'DESARME' THEN 1 ELSE 0 END) AS tem_desarme,
            MAX(CASE WHEN "ctiGrupo" = 'ARME' THEN 1 ELSE 0 END) AS tem_arme,
        
            MAX(CASE WHEN "ctiGrupo" = 'PANICO' THEN 1 ELSE 0 END) AS tem_panico,
            MAX(CASE WHEN "ctiGrupo" = 'MEDICO' THEN 1 ELSE 0 END) AS tem_medico,
            MAX(CASE WHEN "ctiGrupo" = 'EMERGENCIA' THEN 1 ELSE 0 END) AS tem_emergencia,
            MAX(CASE WHEN "ctiGrupo" = 'FALHAS' THEN 1 ELSE 0 END) AS tem_falha,
        
            STRING_AGG(DISTINCT "ctiGrupo", ', ') AS grupos,
            STRING_AGG(DISTINCT "ctiDescricao", ' | ') AS descricoes,
        
            MAX("nomeCliente") AS cliente,
            MAX("conta") AS conta,
            MAX("particao") AS particao,
            MAX("zonaUser") AS zona,
        
            MAX("img") AS imagem,
            MAX("carmeraAtiva") AS camera_ativa
        
          FROM eventos
          GROUP BY "idProcesso"
        ),
        
        historico AS (
          SELECT
            "idDispositivo",
        
            COUNT(*) FILTER (WHERE "ctiGrupo" = 'ALARME') AS qtd_alarme_hist,
            COUNT(*) FILTER (WHERE "ctiGrupo" = 'RESTAURE') AS qtd_restaure_hist
        
          FROM eventos
          GROUP BY "idDispositivo"
        )
        
        SELECT
          p."idProcesso",
        
          p.cliente,
          p.conta,
          p.particao,
          p.zona,
        
          p.inicio_evento,
          p.fim_evento,
        
          EXTRACT(EPOCH FROM (p.fim_evento - p.inicio_evento))::integer
            AS duracao_segundos,
        
          p.total_eventos,
        
          p.grupos,
          p.descricoes,
        
          p.imagem,
          p.camera_ativa,
        
          /* ===========================================
             PERFIL DO LOCAL
          =========================================== */
        
          CASE
            WHEN h.qtd_alarme_hist >= 40
             AND h.qtd_restaure_hist >= (h.qtd_alarme_hist * 0.7)
            THEN 'LOCAL_COM_MOVIMENTO_REPETITIVO'
        
            ELSE 'LOCAL_NORMAL'
          END AS perfil_local,
        
          /* ===========================================
             STATUS FINAL
          =========================================== */
        
          CASE
        
            /* CRÍTICOS */
        
            WHEN p.tem_panico = 1
              THEN 'CRITICO'
        
            WHEN p.tem_medico = 1
              THEN 'CRITICO'
        
            WHEN p.tem_emergencia = 1
              THEN 'CRITICO'
        
            /* FALHAS */
        
            WHEN p.tem_falha = 1
              THEN 'FALHA_TECNICA'
        
            /* ALARME SEM RESTAURE */
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 0
              THEN 'ATIVO'
        
            /* ALARME + RESTAURE */
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 0
              THEN 'NORMALIZADO'
        
            /* ALARME + RESTAURE + DESARME */
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 1
              THEN 'RESOLVIDO'
        
            /* OPERACIONAL */
        
            WHEN p.tem_arme = 1
              OR p.tem_desarme = 1
              THEN 'OPERACIONAL'
        
            ELSE 'INFORMATIVO'
        
          END AS status_final,
        
          /* ===========================================
             TITULO CLIENTE
          =========================================== */
        
          CASE
        
            WHEN p.tem_panico = 1
              THEN 'Botão de pânico acionado'
        
            WHEN p.tem_medico = 1
              THEN 'Emergência médica acionada'
        
            WHEN p.tem_emergencia = 1
              THEN 'Emergência acionada'
        
            WHEN p.tem_falha = 1
              THEN 'Falha técnica detectada'
        
            WHEN p.tem_alarme = 1
              THEN 'Movimento detectado'
        
            WHEN p.tem_arme = 1
              THEN 'Sistema armado'
        
            WHEN p.tem_desarme = 1
              THEN 'Sistema desarmado'
        
            ELSE 'Evento registrado'
        
          END AS titulo_cliente,
        
          /* ===========================================
             DESCRIÇÃO HUMANIZADA
          =========================================== */
        
          CASE
        
            WHEN p.tem_panico = 1
              THEN 'O sistema registrou um acionamento de pânico.'
        
            WHEN p.tem_medico = 1
              THEN 'O sistema registrou uma emergência médica.'
        
            WHEN p.tem_emergencia = 1
              THEN 'O sistema registrou uma emergência.'
        
            WHEN p.tem_falha = 1
              THEN 'O sistema identificou uma falha técnica ou de comunicação.'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 0
              THEN 'O sistema detectou atividade sem normalização registrada.'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 0
              THEN 'O evento foi normalizado automaticamente.'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 1
              THEN 'A ocorrência foi encerrada após interação do usuário.'
        
            WHEN p.tem_arme = 1
              THEN 'O sistema foi armado.'
        
            WHEN p.tem_desarme = 1
              THEN 'O sistema foi desarmado.'
        
            ELSE 'Evento operacional registrado.'
        
          END AS descricao_cliente
        
        FROM processos p
        LEFT JOIN historico h
          ON TRUE
        
        ORDER BY p.inicio_evento DESC
        LIMIT 100;
        """
      parser = "template_engine"
      response_type = "list"
    } as $x1
  
    !db.direct_query {
      sql = """
        WITH eventos AS (
          SELECT
            e.*,
        
            CASE
              WHEN (e."created_at"::bigint) > 9999999999
                THEN to_timestamp((e."created_at"::bigint / 1000))
              ELSE to_timestamp(e."created_at"::bigint)
            END AS data_evento
        
          FROM x1_3 e
          WHERE e."idDispositivo" = '{{ $input.idDispositivo }}'
        ),
        
        processos AS (
          SELECT
            "idProcesso",
        
            MIN(data_evento) AS inicio_evento,
            MAX(data_evento) AS fim_evento,
        
            COUNT(*) AS total_eventos,
        
            MAX(CASE WHEN "ctiGrupo" = 'ALARME' THEN 1 ELSE 0 END) AS tem_alarme,
            MAX(CASE WHEN "ctiGrupo" = 'RESTAURE' THEN 1 ELSE 0 END) AS tem_restaure,
            MAX(CASE WHEN "ctiGrupo" = 'DESARME' THEN 1 ELSE 0 END) AS tem_desarme,
            MAX(CASE WHEN "ctiGrupo" = 'ARME' THEN 1 ELSE 0 END) AS tem_arme,
        
            MAX(CASE WHEN "ctiGrupo" = 'PANICO' THEN 1 ELSE 0 END) AS tem_panico,
            MAX(CASE WHEN "ctiGrupo" = 'MEDICO' THEN 1 ELSE 0 END) AS tem_medico,
            MAX(CASE WHEN "ctiGrupo" = 'EMERGENCIA' THEN 1 ELSE 0 END) AS tem_emergencia,
            MAX(CASE WHEN "ctiGrupo" = 'FALHAS' THEN 1 ELSE 0 END) AS tem_falha,
        
            STRING_AGG(DISTINCT "ctiGrupo", ', ') AS grupos,
            STRING_AGG(DISTINCT "ctiDescricao", ' | ') AS descricoes,
        
            MAX("nomeCliente") AS cliente,
            MAX("conta") AS conta,
            MAX("particao") AS particao,
            MAX("zonaUser") AS zona,
        
            MAX("img") AS imagem,
            MAX("carmeraAtiva") AS camera_ativa
        
          FROM eventos
          GROUP BY "idProcesso"
        ),
        
        historico AS (
          SELECT
            COUNT(*) FILTER (WHERE "ctiGrupo" = 'ALARME') AS qtd_alarme_hist,
            COUNT(*) FILTER (WHERE "ctiGrupo" = 'RESTAURE') AS qtd_restaure_hist
          FROM eventos
        )
        
        SELECT
        
          /* ===========================================
             DADOS DO PROCESSO
          =========================================== */
        
          p."idProcesso",
        
          p.cliente,
          p.conta,
          p.particao,
          p.zona,
        
          p.inicio_evento,
          p.fim_evento,
        
          EXTRACT(EPOCH FROM (p.fim_evento - p.inicio_evento))::integer
            AS duracao_segundos,
        
          p.total_eventos,
        
          p.grupos,
          p.descricoes,
        
          p.imagem,
          p.camera_ativa,
        
          /* ===========================================
             PERFIL LOCAL
          =========================================== */
        
          CASE
            WHEN h.qtd_alarme_hist >= 40
             AND h.qtd_restaure_hist >= (h.qtd_alarme_hist * 0.7)
            THEN 'LOCAL_COM_MOVIMENTO_REPETITIVO'
        
            ELSE 'LOCAL_NORMAL'
          END AS perfil_local,
        
          /* ===========================================
             STATUS FINAL
          =========================================== */
        
          CASE
        
            WHEN p.tem_panico = 1
              THEN 'CRITICO'
        
            WHEN p.tem_medico = 1
              THEN 'CRITICO'
        
            WHEN p.tem_emergencia = 1
              THEN 'CRITICO'
        
            WHEN p.tem_falha = 1
              THEN 'FALHA_TECNICA'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 0
              THEN 'ATIVO'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 0
              THEN 'NORMALIZADO'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 1
              THEN 'RESOLVIDO'
        
            WHEN p.tem_arme = 1
              OR p.tem_desarme = 1
              THEN 'OPERACIONAL'
        
            ELSE 'INFORMATIVO'
        
          END AS status_final,
        
          /* ===========================================
             TITULO
          =========================================== */
        
          CASE
        
            WHEN p.tem_panico = 1
              THEN 'Botão de pânico acionado'
        
            WHEN p.tem_medico = 1
              THEN 'Emergência médica acionada'
        
            WHEN p.tem_emergencia = 1
              THEN 'Emergência acionada'
        
            WHEN p.tem_falha = 1
              THEN 'Falha técnica detectada'
        
            WHEN p.tem_alarme = 1
              THEN 'Movimento detectado'
        
            WHEN p.tem_arme = 1
              THEN 'Sistema armado'
        
            WHEN p.tem_desarme = 1
              THEN 'Sistema desarmado'
        
            ELSE 'Evento registrado'
        
          END AS titulo_cliente,
        
          /* ===========================================
             DESCRICAO
          =========================================== */
        
          CASE
        
            WHEN p.tem_panico = 1
              THEN 'O sistema registrou um acionamento de pânico.'
        
            WHEN p.tem_medico = 1
              THEN 'O sistema registrou uma emergência médica.'
        
            WHEN p.tem_emergencia = 1
              THEN 'O sistema registrou uma emergência.'
        
            WHEN p.tem_falha = 1
              THEN 'O sistema identificou uma falha técnica ou de comunicação.'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 0
              THEN 'O sistema detectou atividade sem normalização registrada.'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 0
              THEN 'O evento foi normalizado automaticamente.'
        
            WHEN p.tem_alarme = 1
             AND p.tem_restaure = 1
             AND p.tem_desarme = 1
              THEN 'A ocorrência foi encerrada após interação do usuário.'
        
            WHEN p.tem_arme = 1
              THEN 'O sistema foi armado.'
        
            WHEN p.tem_desarme = 1
              THEN 'O sistema foi desarmado.'
        
            ELSE 'Evento operacional registrado.'
        
          END AS descricao_cliente,
        
          /* ===========================================
             EVENTOS
          =========================================== */
        
          e."id"                  AS evento_id,
          e."idEvento",
          e."ctiGrupo",
          e."ctiDescricao",
          e."particao"            AS evento_particao,
          e."zonaUser"            AS evento_zona,
          e."nivel",
          e."img"                 AS evento_imagem,
          e."carmeraAtiva"        AS evento_camera,
        
          to_char(e.data_evento, 'DD/MM/YYYY') AS evento_data,
          to_char(e.data_evento, 'HH24:MI:SS') AS evento_hora,
          to_char(e.data_evento, 'DD/MM/YYYY HH24:MI:SS') AS evento_datahora
        
        FROM processos p
        
        INNER JOIN eventos e
          ON e."idProcesso" = p."idProcesso"
        
        LEFT JOIN historico h
          ON TRUE
        
        ORDER BY
          p.inicio_evento DESC,
          e.data_evento ASC
        
        LIMIT 500;
        """
      parser = "template_engine"
      response_type = "list"
    } as $x1
  
    db.direct_query {
      sql = """
        WITH eventos AS (
          SELECT
            e.*,
            CASE
              WHEN (e."created_at"::bigint) > 9999999999
                THEN to_timestamp((e."created_at"::bigint / 1000)) AT TIME ZONE 'America/Sao_Paulo'
              ELSE to_timestamp(e."created_at"::bigint) AT TIME ZONE 'America/Sao_Paulo'
            END AS data_evento
          FROM x1_3 e
          WHERE e."idDispositivo" = '{{ $input.idDispositivo }}'
        ),
        
        processos AS (
          SELECT
            "idProcesso",
        
            MIN(data_evento) AS inicio_evento,
            MAX(data_evento) AS fim_evento,
            COUNT(*) AS total_eventos,
        
            MAX(CASE WHEN "ctiGrupo" = 'ALARME' THEN 1 ELSE 0 END) AS tem_alarme,
            MAX(CASE WHEN "ctiGrupo" = 'RESTAURE' THEN 1 ELSE 0 END) AS tem_restaure,
            MAX(CASE WHEN "ctiGrupo" = 'DESARME' THEN 1 ELSE 0 END) AS tem_desarme,
            MAX(CASE WHEN "ctiGrupo" = 'ARME' THEN 1 ELSE 0 END) AS tem_arme,
        
            MAX(CASE WHEN "ctiGrupo" = 'PANICO' THEN 1 ELSE 0 END) AS tem_panico,
            MAX(CASE WHEN "ctiGrupo" = 'MEDICO' THEN 1 ELSE 0 END) AS tem_medico,
            MAX(CASE WHEN "ctiGrupo" = 'EMERGENCIA' THEN 1 ELSE 0 END) AS tem_emergencia,
            MAX(CASE WHEN "ctiGrupo" = 'FALHAS' THEN 1 ELSE 0 END) AS tem_falha,
        
            STRING_AGG(DISTINCT "ctiGrupo", ', ') AS grupos,
            STRING_AGG(DISTINCT "ctiDescricao", ' | ') AS descricoes,
        
            MAX("nomeCliente") AS cliente,
            MAX("conta") AS conta,
            MAX("particao") AS particao,
            MAX("zonaUser") AS zona,
            MAX("img") AS imagem,
            MAX("carmeraAtiva") AS camera_ativa
        
          FROM eventos
          GROUP BY "idProcesso"
        ),
        
        ultimos_processos AS (
          SELECT *
          FROM processos
          ORDER BY inicio_evento DESC
          LIMIT 25
        ),
        
        historico AS (
          SELECT
            COUNT(*) FILTER (WHERE "ctiGrupo" = 'ALARME') AS qtd_alarme_hist,
            COUNT(*) FILTER (WHERE "ctiGrupo" = 'RESTAURE') AS qtd_restaure_hist
          FROM eventos
        )
        
        SELECT
          p."idProcesso",
        
          p.cliente,
          p.conta,
          p.particao,
          p.zona,
        
          p.inicio_evento,
          p.fim_evento,
        
          EXTRACT(EPOCH FROM (p.fim_evento - p.inicio_evento))::integer AS duracao_segundos,
        
          p.total_eventos,
          p.grupos,
          p.descricoes,
          p.imagem,
          p.camera_ativa,
        
          CASE
            WHEN h.qtd_alarme_hist >= 40
             AND h.qtd_restaure_hist >= (h.qtd_alarme_hist * 0.7)
            THEN 'LOCAL_COM_MOVIMENTO_REPETITIVO'
            ELSE 'LOCAL_NORMAL'
          END AS perfil_local,
        
          CASE
            WHEN p.tem_panico = 1 THEN 'CRITICO'
            WHEN p.tem_medico = 1 THEN 'CRITICO'
            WHEN p.tem_emergencia = 1 THEN 'CRITICO'
            WHEN p.tem_falha = 1 THEN 'FALHA_TECNICA'
            WHEN p.tem_alarme = 1 AND p.tem_restaure = 0 THEN 'ATIVO'
            WHEN p.tem_alarme = 1 AND p.tem_restaure = 1 AND p.tem_desarme = 0 THEN 'NORMALIZADO'
            WHEN p.tem_alarme = 1 AND p.tem_restaure = 1 AND p.tem_desarme = 1 THEN 'RESOLVIDO'
            WHEN p.tem_arme = 1 OR p.tem_desarme = 1 THEN 'OPERACIONAL'
            ELSE 'INFORMATIVO'
          END AS status_final,
        
          CASE
            WHEN p.tem_panico = 1 THEN 'Botão de pânico acionado'
            WHEN p.tem_medico = 1 THEN 'Emergência médica acionada'
            WHEN p.tem_emergencia = 1 THEN 'Emergência acionada'
            WHEN p.tem_falha = 1 THEN 'Falha técnica detectada'
            WHEN p.tem_alarme = 1 THEN 'Movimento detectado'
            WHEN p.tem_arme = 1 THEN 'Sistema armado'
            WHEN p.tem_desarme = 1 THEN 'Sistema desarmado'
            ELSE 'Evento registrado'
          END AS titulo_cliente,
        
          CASE
            WHEN p.tem_panico = 1 THEN 'O sistema registrou um acionamento de pânico.'
            WHEN p.tem_medico = 1 THEN 'O sistema registrou uma emergência médica.'
            WHEN p.tem_emergencia = 1 THEN 'O sistema registrou uma emergência.'
            WHEN p.tem_falha = 1 THEN 'O sistema identificou uma falha técnica ou de comunicação.'
            WHEN p.tem_alarme = 1 AND p.tem_restaure = 0 THEN 'O sistema detectou atividade sem normalização registrada.'
            WHEN p.tem_alarme = 1 AND p.tem_restaure = 1 AND p.tem_desarme = 0 THEN 'O evento foi normalizado automaticamente.'
            WHEN p.tem_alarme = 1 AND p.tem_restaure = 1 AND p.tem_desarme = 1 THEN 'A ocorrência foi encerrada após interação do usuário.'
            WHEN p.tem_arme = 1 THEN 'O sistema foi armado.'
            WHEN p.tem_desarme = 1 THEN 'O sistema foi desarmado.'
            ELSE 'Evento operacional registrado.'
          END AS descricao_cliente,
        
          e."id" AS evento_id,
          e."idEvento",
          e."ctiGrupo",
          e."ctiDescricao",
          e."particao" AS evento_particao,
          e."zonaUser" AS evento_zona,
          e."nivel",
          e."img" AS evento_imagem,
          e."carmeraAtiva" AS evento_camera,
          e."idCliente",
        
          to_char(e.data_evento, 'DD/MM/YYYY') AS evento_data,
          to_char(e.data_evento, 'HH24:MI:SS') AS evento_hora,
          to_char(e.data_evento, 'DD/MM/YYYY HH24:MI:SS') AS evento_datahora
        
        FROM ultimos_processos p
        
        INNER JOIN eventos e
          ON e."idProcesso" = p."idProcesso"
        
        LEFT JOIN historico h
          ON TRUE
        
        ORDER BY
          p.inicio_evento DESC
        """
      parser = "template_engine"
      response_type = "list"
    } as $x1
  
    function.run WebLogarCache as $func1_token
    function.run ConfMonitCliente_DadosClienteID {
      input = {
        Authorization: $func1_token
        idcliente    : $var.x1.[1].idCliente
      }
    } as $func_1
  }

  response = {
    dados: {}|set:"DadosCliente":$func_1.response.result.dados|set:"Eventos":$x1
  }
}