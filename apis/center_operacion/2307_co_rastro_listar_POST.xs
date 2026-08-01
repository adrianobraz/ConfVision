// Lista pontos do rastro de disparos por mapa e/ou processo
query co_rastro_listar verb=POST {
  api_group = "centerOperacion"

  input {
    int mapa_ambiente_id?
    text idProcesso? filters=trim
    text idCliente? filters=trim
    int limite?=120
  }

  stack {
    var $limite {
      value = $input.limite|first_notempty:120
    }
  
    conditional {
      if (($input.mapa_ambiente_id|is_empty) == false && ($input.idProcesso|is_empty) == false) {
        db.query co_rastro_disparo {
          where = $db.co_rastro_disparo.mapa_ambiente_id == $input.mapa_ambiente_id && $db.co_rastro_disparo.idProcesso == $input.idProcesso
          sort = {co_rastro_disparo.evento_ts: "asc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.mapa_ambiente_id|is_empty) == false) {
        db.query co_rastro_disparo {
          where = $db.co_rastro_disparo.mapa_ambiente_id == $input.mapa_ambiente_id
          sort = {co_rastro_disparo.evento_ts: "asc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.idProcesso|is_empty) == false) {
        db.query co_rastro_disparo {
          where = $db.co_rastro_disparo.idProcesso == $input.idProcesso
          sort = {co_rastro_disparo.evento_ts: "asc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.idCliente|is_empty) == false) {
        db.query co_rastro_disparo {
          where = $db.co_rastro_disparo.idCliente == $input.idCliente
          sort = {co_rastro_disparo.evento_ts: "desc"}
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