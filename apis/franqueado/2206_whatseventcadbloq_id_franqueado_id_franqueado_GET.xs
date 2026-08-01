// Get WhatsEventCadBloq record
query "whatseventcadbloq/idFranqueado/{idFranqueado}" verb=GET {
  api_group = "franqueado"

  input {
    text idFranqueado? filters=trim
  }

  stack {
    db.query WhatsEventCadBloq {
      where = $db.WhatsEventCadBloq.idFranqueado == $input.idFranqueado
      return = {type: "list"}
    } as $WhatsEventCadBloq1
  }

  response = {dados: $WhatsEventCadBloq1}
}