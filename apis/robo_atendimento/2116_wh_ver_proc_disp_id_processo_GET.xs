query "wh_VerProcDisp/{idProcesso}" verb=GET {
  api_group = "RoboAtendimento"

  input {
    text idProcesso? filters=trim
  }

  stack {
    db.direct_query {
      sql = """
        SELECT
            a."id",
            a."created_at",
            a."idEvento",
            a."codigo",
            a."particao",
            a."zonaUser",
            a."nivel",
            a."dataEntrada",
            a."img",
            a."idProcesso",
            a."idDispositivo",
            a."idCliente",
            a."nomeCliente",
            a."emailCliente",
            a."ctiGrupo",
            a."ctiDescricao",
            a."idFranqueado",
            a."codigoBenuvem",
            a."conta",
            a."carmeraAtiva",
            a."Data",
            COUNT(*) OVER (
                PARTITION BY
                    a."idProcesso",
                    a."idDispositivo",
                    a."particao",
                    a."zonaUser",
                    a."idCliente",
                    a."ctiGrupo"
            ) AS "totalprocessos"
        FROM x1_3 a
        WHERE a."idProcesso" = '{{$input.idProcesso}}'
        ORDER BY a."dataEntrada" DESC;
        """
      parser = "template_engine"
      response_type = "list"
    } as $x1
  
    precondition (($x1|is_empty) == false)
  }

  response = {Eventos: $x1}
}