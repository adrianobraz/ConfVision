// Listar eventos do cliente com paginação (50 por página)
query vis_evento_by_cliente_page verb=GET {
  api_group = "confVision"

  input {
    text id_cliente? filters=trim
    timestamp? data_de?
    timestamp? data_ate?
    int page?=1
  }

  stack {
    precondition (($input.id_cliente|is_empty) == false) {
      error = "id_cliente obrigatorio"
    }
  
    db.query vis_evento {
      where = $db.vis_evento.id_cliente == $input.id_cliente && $db.vis_evento.created_at >=? $input.data_de && $db.vis_evento.created_at <=? $input.data_ate
      sort = {vis_evento.created_at: "desc"}
      return = {type: "list", paging: {page: $input.page, per_page: 50}}
    } as $resultado
  }

  response = {dados: $resultado}
}