// Delete WhatsEventCadFranq record
query "whatseventcadfranq/{whatseventcadfranq_id}" verb=DELETE {
  api_group = "franqueado"

  input {
    int whatseventcadfranq_id? filters=min:1
  }

  stack {
    db.del WhatsEventCadFranq {
      field_name = "id"
      field_value = $input.whatseventcadfranq_id
    }
  }

  response = null
}