// Listar eventos do franqueado
query vis_evento_by_franqueado verb=GET {
  api_group = "confVision"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    db.query vis_evento {
      where = $db.vis_evento.id_franqueado == $input.id_franqueado
      sort = {vis_evento.created_at: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista}
}