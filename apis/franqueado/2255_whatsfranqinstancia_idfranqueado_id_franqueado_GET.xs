// Get instancia Evo GO (Dialyze) do franqueado em WhatsappFranqueadoNroTelefone
query "whatsfranqinstancia/idfranqueado/{idFranqueado}" verb=GET {
  api_group = "franqueado"

  input {
    text idFranqueado? filters=trim
  }

  stack {
    db.query WhatsappFranqueadoNroTelefone {
      where = $db.WhatsappFranqueadoNroTelefone.franqueado == $input.idFranqueado && $db.WhatsappFranqueadoNroTelefone.tipoApi == "G"
      return = {type: "single"}
    } as $WhatsappFranqueadoNroTelefone1
  }

  response = {dados: $WhatsappFranqueadoNroTelefone1}
}