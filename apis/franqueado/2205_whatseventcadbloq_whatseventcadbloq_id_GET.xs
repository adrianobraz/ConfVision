// Get WhatsEventCadBloq record
query "whatseventcadbloq/{whatseventcadbloq_id}" verb=GET {
  api_group = "franqueado"

  input {
    int whatseventcadbloq_id? filters=min:1
  }

  stack {
    db.get WhatsEventCadBloq {
      field_name = "id"
      field_value = $input.whatseventcadbloq_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = $model
}