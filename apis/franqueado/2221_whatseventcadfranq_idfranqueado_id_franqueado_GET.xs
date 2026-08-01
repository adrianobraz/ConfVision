// Get WhatsEventCadFranq record
query "whatseventcadfranq/idfranqueado/{idFranqueado}" verb=GET {
  api_group = "franqueado"

  input {
    text idFranqueado? filters=trim
  }

  stack {
    db.query WhatsEventCadFranq {
      where = $db.WhatsEventCadFranq.idFranqueado == $input.idFranqueado
      return = {type: "single"}
    } as $WhatsEventCadFranq1
  }

  response = {dados: $WhatsEventCadFranq1}
}