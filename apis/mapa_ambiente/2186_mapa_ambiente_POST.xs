// Add mapa_ambiente record
query mapa_ambiente verb=POST {
  api_group = "mapaAmbiente"

  input {
    dblink {
      table = "mapa_ambiente"
    }
  }

  stack {
    db.add mapa_ambiente {
      enforce_hidden_fields = false
      data = {
        created_at    : "now"
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