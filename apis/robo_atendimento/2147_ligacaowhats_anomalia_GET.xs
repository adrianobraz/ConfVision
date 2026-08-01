// gera um relatorio alarmes com problemas
query ligacaowhats_Anomalia verb=GET {
  api_group = "RoboAtendimento"

  input {
  }

  stack {
    db.direct_query {
      sql = """
        WITH base AS (
          SELECT
            x.*,
            COUNT(*) OVER (
              PARTITION BY x.nomecliente, x.tipoevento, x.zona
            ) AS total,
            ROW_NUMBER() OVER (
              PARTITION BY x.nomecliente, x.tipoevento, x.zona
              ORDER BY x.created_at DESC, x.id DESC
            ) AS rn
          FROM x1_66 x
          WHERE TO_TIMESTAMP(x.created_at / 1000) >= NOW() - INTERVAL '7 days'
        )
        SELECT
          nomecliente,
          tipoevento,
          zona,
          total,
          id,
          created_at,
          franqueado,
          telefone,
          empresanome,
          local,
          datahorario,
          alarm_events_id,
          ideventgo,
          falha,
          tentativas,
          "dtUltimaTentativa",
          exec,
          atendido,
          "idProcesso",
          "idDispositivo",
          notificarsempre,
          "DispNome",
          "DispDescricao",
          "DispTipo"
        FROM base
        WHERE total > 10
          AND rn = 1
        ORDER BY total DESC, nomecliente ASC;
        """
      parser = "template_engine"
      response_type = "list"
    } as $x1
  }

  response = $x1
}