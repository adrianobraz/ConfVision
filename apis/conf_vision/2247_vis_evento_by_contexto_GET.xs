// Eventos ConfVision no contexto do processo (sensor por id_processo + analitico por cliente/dispositivo/periodo)
query vis_evento_by_contexto verb=GET {
  api_group = "confVision"

  input {
    text id_processo? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    text id_franqueado? filters=trim
    timestamp? data_inicio?
  }

  stack {
    db.query vis_evento {
      where = $db.vis_evento.id_franqueado ==? $input.id_franqueado && (($db.vis_evento.id_processo ==? $input.id_processo) || ($db.vis_evento.id_cliente ==? $input.id_cliente && $db.vis_evento.id_dispositivo ==? $input.id_dispositivo && $db.vis_evento.created_at >=? $input.data_inicio))
      sort = {vis_evento.created_at: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista}
}