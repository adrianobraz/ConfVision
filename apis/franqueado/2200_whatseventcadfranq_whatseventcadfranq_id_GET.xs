// Get WhatsEventCadFranq record
query "whatseventcadfranq/{whatseventcadfranq_id}" verb=GET {
  api_group = "franqueado"

  input {
    int whatseventcadfranq_id? filters=min:1
  }

  stack {
    db.get WhatsEventCadFranq {
      field_name = "id"
      field_value = $input.whatseventcadfranq_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = $model
}