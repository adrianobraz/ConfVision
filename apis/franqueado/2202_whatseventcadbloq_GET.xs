// Query all WhatsEventCadBloq records
query whatseventcadbloq verb=GET {
  api_group = "franqueado"

  input {
  }

  stack {
    db.query WhatsEventCadBloq {
      return = {type: "list"}
    } as $model
  }

  response = $model
}