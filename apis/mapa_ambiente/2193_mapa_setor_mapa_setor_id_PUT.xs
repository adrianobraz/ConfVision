// Update mapa_setor record
query "mapa_setor/{mapa_setor_id}" verb=PUT {
  api_group = "mapaAmbiente"

  input {
    int mapa_setor_id? filters=min:1
    dblink {
      table = "mapa_setor"
    }
  }

  stack {
    db.edit mapa_setor {
      field_name = "id"
      field_value = $input.mapa_setor_id
      enforce_hidden_fields = false
      data = {
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