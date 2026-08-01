// Lista eventos da Linha do Tempo (caixa-preta) por mapa e/ou processo
query co_timeline_listar verb=POST {
  api_group = "centerOperacion"

  input {
    int mapa_ambiente_id?
    text idProcesso? filters=trim
    text idCliente? filters=trim
    int limite?=200
  }

  stack {
    var $limite {
      value = $input.limite|first_notempty:200
    }
  
    conditional {
      if (($input.mapa_ambiente_id|is_empty) == false && ($input.idProcesso|is_empty) == false) {
        db.query co_timeline_evento {
          where = $db.co_timeline_evento.mapa_ambiente_id == $input.mapa_ambiente_id && $db.co_timeline_evento.idProcesso == $input.idProcesso
          sort = {co_timeline_evento.evento_ts: "asc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.mapa_ambiente_id|is_empty) == false) {
        db.query co_timeline_evento {
          where = $db.co_timeline_evento.mapa_ambiente_id == $input.mapa_ambiente_id
          sort = {co_timeline_evento.evento_ts: "asc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.idProcesso|is_empty) == false) {
        db.query co_timeline_evento {
          where = $db.co_timeline_evento.idProcesso == $input.idProcesso
          sort = {co_timeline_evento.evento_ts: "asc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.idCliente|is_empty) == false) {
        db.query co_timeline_evento {
          where = $db.co_timeline_evento.idCliente == $input.idCliente
          sort = {co_timeline_evento.evento_ts: "desc"}
          return = {type: "list"}
        } as $lista
      }
    
      else {
        var $lista {
          value = []
        }
      }
    }
  }

  response = $lista
}