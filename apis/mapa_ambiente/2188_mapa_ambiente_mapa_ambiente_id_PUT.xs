// Update mapa_ambiente record
query "mapa_ambiente/{mapa_ambiente_id}" verb=PUT {
  api_group = "mapaAmbiente"

  input {
    int mapa_ambiente_id? filters=min:1
    dblink {
      table = "mapa_ambiente"
    }
  }

  stack {
    db.edit mapa_ambiente {
      field_name = "id"
      field_value = $input.mapa_ambiente_id
      enforce_hidden_fields = false
      data = {
        descricao     : $input.descricao
        imagem_url    : $input.imagem_url
        idCliente     : $input.idCliente
        idFranqueado  : $input.idFranqueado
        nomeCliente   : $input.nomeCliente
        nomeFranqueado: $input.nomeFranqueado
        ordem         : $input.ordem
        ativo         : $input.ativo
      }
    } as $model
  }

  response = $model
}