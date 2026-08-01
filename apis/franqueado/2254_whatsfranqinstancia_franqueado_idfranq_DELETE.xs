// Remove a instancia Evo GO (Dialyze) do franqueado em WhatsappFranqueadoNroTelefone
query "whatsfranqinstancia/franqueado/{idfranq}" verb=DELETE {
  api_group = "franqueado"

  input {
    text idfranq? filters=trim
  }

  stack {
    precondition (($input.idfranq|is_empty) == false) {
      error = "Franqueado Vazio"
    }
  
    db.query WhatsappFranqueadoNroTelefone {
      where = $db.WhatsappFranqueadoNroTelefone.franqueado == $input.idfranq && $db.WhatsappFranqueadoNroTelefone.tipoApi == "G"
      return = {type: "list"}
    } as $WhatsappFranqueadoNroTelefone1
  
    foreach ($WhatsappFranqueadoNroTelefone1) {
      each as $item {
        db.del WhatsappFranqueadoNroTelefone {
          field_name = "id"
          field_value = $item.id
        }
      }
    }
  }

  response = null
}