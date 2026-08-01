// Lista mapas ativos — com idFranqueado filtra por franqueado; sem filtro retorna todos
query mapa_ambiente_query_all verb=POST {
  api_group = "mapaAmbiente"

  input {
    text idFranqueado? filters=trim
  }

  stack {
    conditional {
      if (($input.idFranqueado|strlen) > 0) {
        db.query mapa_ambiente {
          where = $db.mapa_ambiente.idFranqueado == $input.idFranqueado && ($db.mapa_ambiente.ativo || $db.mapa_ambiente.ativo == null)
          return = {type: "list"}
        } as $model
      }
    
      else {
        db.query mapa_ambiente {
          where = $db.mapa_ambiente.ativo == true || $db.mapa_ambiente.ativo == null
          return = {type: "list"}
        } as $model
      }
    }
  }

  response = $model
}