query Relatorio_Falha_Franqueado verb=POST {
  api_group = "GraficoEventos"

  input {
    text idFranqueado filters=trim
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.idFranqueado == $input.idFranqueado && $db.alarm_events.ctiGrupo == "FALHAS"
      sort = {Total: "desc", alarm_events_nomeCliente1: "asc"}
      return = {
        type : "aggregate"
        group: {
          descricao                : $db.alarm_events.ctiDescricao
          alarm_events_nomeCliente1: $db.alarm_events.nomeCliente
        }
        eval : {Total: $db.alarm_events.ctiDescricao|count}
      }
    } as $resultado
  }

  response = $resultado
}