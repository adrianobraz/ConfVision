// Add mapa_setor record
query mapa_setor verb=POST {
  api_group = "mapaAmbiente"

  input {
    dblink {
      table = "mapa_setor"
    }
  }

  stack {
    db.add mapa_setor {
      enforce_hidden_fields = false
      data = {
        created_at      : "now"
        mapa_ambiente_id: $input.mapa_ambiente_id
        idSetor         : $input.idSetor
        idDispositivo   : $input.idDispositivo
        idCliente       : $input.idCliente
        setornome       : $input.setornome
        dispositivoNome : $input.dispositivoNome
        posX            : $input.posX
        poxY            : $input.poxY
        icone           : $input.icone
        label           : $input.label
        descricao       : $input.descricao
        clienteNome     : $input.clienteNome
      }
    } as $model
  }

  response = $model
}