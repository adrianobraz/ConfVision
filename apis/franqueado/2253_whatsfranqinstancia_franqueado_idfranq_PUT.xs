// Upsert da instancia Evo GO (Dialyze) do franqueado em WhatsappFranqueadoNroTelefone
query "whatsfranqinstancia/franqueado/{idfranq}" verb=PUT {
  api_group = "franqueado"

  input {
    text idfranq? filters=trim
    text telefone? filters=trim
    text txtApiToken? filters=trim
    text instancia? filters=trim
    text FranqueadoNome? filters=trim
  }

  stack {
    precondition (($input.idfranq|is_empty) == false) {
      error = "Franqueado Vazio"
    }
  
    db.query WhatsappFranqueadoNroTelefone {
      where = $db.WhatsappFranqueadoNroTelefone.franqueado == $input.idfranq && $db.WhatsappFranqueadoNroTelefone.tipoApi == "G"
      return = {type: "single"}
    } as $WhatsappFranqueadoNroTelefone1
  
    conditional {
      if ($WhatsappFranqueadoNroTelefone1|is_empty) {
        db.add WhatsappFranqueadoNroTelefone {
          enforce_hidden_fields = false
          data = {
            created_at    : "now"
            franqueado    : $input.idfranq
            telefone      : $input.telefone
            txtApiToken   : $input.txtApiToken
            instancia     : $input.instancia
            tipoApi       : "G"
            modulotexto   : true
            ModuloLigar   : false
            Grupo         : "G"
            FranqueadoNome: $input.FranqueadoNome
          }
        } as $model
      }
    
      else {
        db.edit WhatsappFranqueadoNroTelefone {
          field_name = "id"
          field_value = $WhatsappFranqueadoNroTelefone1.id
          enforce_hidden_fields = false
          data = {
            telefone      : $input.telefone
            txtApiToken   : $input.txtApiToken
            instancia     : $input.instancia
            tipoApi       : "G"
            modulotexto   : true
            FranqueadoNome: $input.FranqueadoNome
          }
        } as $model
      }
    }
  }

  response = {dados: $model}
}