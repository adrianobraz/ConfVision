// Ultimos 25 eventos ConfVision do dispositivo (terminal — LIMIT no banco)
query vis_evento_ultimos25_dispositivo verb=GET {
  api_group = "confVision"

  input {
    text id_dispositivo? filters=trim
    int offset?
    int limit?=25
  }

  stack {
    var $lim {
      value = 25
    }
  
    conditional {
      if ($input.limit != null && $input.limit > 0 && $input.limit <= 100) {
        var.update $lim {
          value = $input.limit
        }
      }
    }
  
    db.query vis_evento {
      where = $db.vis_evento.id_dispositivo == $input.id_dispositivo
      sort = {vis_evento.created_at: "desc"}
      return = {
        type  : "list"
        paging: {page: 1, per_page: $lim, metadata: false}
      }
    } as $lista
  }

  response = {
    dados : $lista
    total : $lista|count
    offset: 0
    limit : $lim
  }
}