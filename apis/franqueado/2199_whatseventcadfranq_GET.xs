// Query all WhatsEventCadFranq records
query whatseventcadfranq verb=GET {
  api_group = "franqueado"

  input {
  }

  stack {
    db.query WhatsEventCadFranq {
      return = {type: "list"}
    } as $model
  }

  response = $model
}